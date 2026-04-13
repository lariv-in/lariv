package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerJsonImport handles bulk creation of type T from a JSON file upload
// on POST requests. On non-POST methods it passes through to next unchanged.
//
// On POST it parses the view's form, validates and extracts the uploaded file
// from the field named FileField, decodes it as a JSON array of T, and inserts
// all records in batches of 100 within a single transaction. On success the
// number of imported records is stored in the context as "$count".
//
// If SuccessURL is set, a successful import redirects to the resolved URL.
// If SuccessURL is nil, next is called with the enriched context.
//
// All errors (form parsing, file validation, JSON decoding, DB) are placed
// into getters.ContextKeyError under "_form" and next is called, so the page
// can re-render with error feedback.
type LayerJsonImport[T any] struct {
	FileField  string
	SuccessURL getters.Getter[string]
}

func (m LayerJsonImport[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("views: layer json import: parse form", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("views: layer json import: field error", "field", fname, "error", ferr)
			}
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		fileHeader, err := uploadedJSONFile(values, m.FileField)
		if err != nil {
			slog.Error("views: layer json import: uploaded file", "error", err)
			fieldErrors["_form"] = err
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		records, err := decodeJSONArrayFile[T](fileHeader)
		if err != nil {
			slog.Error("views: layer json import: decode json", "error", err)
			fieldErrors["_form"] = fmt.Errorf("invalid json import: %w", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer json import: db from context", "error", dberr)
			fieldErrors["_form"] = dberr
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if len(records) > 0 {
			if err := db.Transaction(func(tx *gorm.DB) error {
				return gorm.G[T](tx).CreateInBatches(r.Context(), &records, 100)
			}); err != nil {
				slog.Error("views: layer json import: batch create", "error", err)
				fieldErrors["_form"] = fmt.Errorf("%v", err)
				ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		ctx = context.WithValue(ctx, "$count", len(records))
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("views: layer json import: resolve success URL", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		HtmxRedirect(w, r, successUrl, http.StatusSeeOther)
	})
}
