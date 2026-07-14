package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerJsonImport handles bulk creation of database records of type T from a JSON file upload on incoming POST requests.
// On non-POST requests, it passes through downstream.
//
// On a POST trigger, it extracts the uploaded file from FileField, decodes it as a JSON array of type T,
// and runs batch creations of 100 records inside a transaction. Upon successful completion, the imported count is saved as "$count" on the request context.
//
// Use Cases:
//   - Importing bulk data sets from structured JSON files (e.g. uploading user spreadsheets or importing inventories).
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerJsonImport[Product]{
//	            FileField:  "import_file",
//	            SuccessURL: lago.RoutePath("products.List", nil),
//	        },
//	    },
//	}
type LayerJsonImport[T any] struct {
	// FileField represents the form parameter key name representing the uploaded JSON file.
	FileField string
	// SuccessURL represents the dynamic Getter resolving to the redirection target URL upon successful bulk imports.
	SuccessURL getters.Getter[string]
}

// Next wraps the downstream HTTP request handlers executing bulk JSON file imports.
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
