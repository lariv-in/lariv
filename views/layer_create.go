package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerCreate handles row creation for type T on POST requests.
// On non-POST methods it passes through to next unchanged.
//
// On POST it parses the view's form, applies FormPatchers, then within a
// transaction populates a new T from the regular (non-association) values,
// inserts it, and replaces any many-to-many associations. On success the
// newly created record's ID is stored in the context as "$id".
//
// If SuccessURL is set, a successful create redirects to the resolved URL.
// If SuccessURL is nil, next is called with the enriched context, allowing a
// downstream handler to decide the response.
//
// All errors (form parsing, validation, DB) are placed into
// getters.ContextKeyError ("_form" for form/field errors) and next is called,
// so the page can re-render with error feedback.
type LayerCreate[T any] struct {
	SuccessURL   getters.Getter[string]
	FormPatchers FormPatchers
}

func (m LayerCreate[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("views: layer create: parse form", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		values, fieldErrors = m.FormPatchers.Apply(view, r, values, fieldErrors)
		ctx = r.Context()
		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("views: layer create: field error", "field", fname, "error", ferr)
			}
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer create: db from context", "error", dberr)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_global": dberr,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		regularValues, associationValues := SplitAssociationValues(values)
		record := new(T)
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := PopulateFromMap(record, regularValues); err != nil {
				return err
			}
			if err := gorm.G[T](tx).Create(r.Context(), record); err != nil {
				return err
			}
			return applyAssociationReplacements(tx, record, associationValues)
		})
		if err != nil {
			slog.Error("views: layer create: transaction", "error", err)
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if hook, ok := any(record).(TxCommitHook); ok {
			hook.AfterTxCommit(db)
		}

		id := uint(reflect.ValueOf(*record).FieldByName("ID").Uint())
		ctx = context.WithValue(ctx, "$id", id)
		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("views: layer create: resolve success URL", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		HtmxRedirect(w, r, successUrl, http.StatusSeeOther)
	})
}
