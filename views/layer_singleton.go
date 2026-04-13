package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerSingleton manages a single-row table of type T (e.g. site-wide
// settings). It combines the roles of detail-loading and updating in one
// layer because there is no URL primary key — the record is always
// obtained via FirstOrCreate.
//
// On GET (or any non-POST method) it loads or creates the singleton and places
// its fields into the context as getters.ContextKeyIn so form components can
// pre-fill values. Then it calls next.
//
// On POST it parses the view's form, then within a transaction loads the
// singleton, updates its columns from the submitted values, and replaces any
// many-to-many associations.
//
// If SuccessURL is set, a successful update redirects to the resolved URL.
// If SuccessURL is nil, next is called so a downstream handler can decide the
// response.
//
// All errors (form parsing, validation, DB) are placed into
// getters.ContextKeyError ("_form" for form/field errors) and next is called,
// so the page can re-render with error feedback.
type LayerSingleton[T any] struct {
	SuccessURL getters.Getter[string]
}

func (m LayerSingleton[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer singleton: db from context", "error", dberr)
			ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_global": dberr,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if r.Method != http.MethodPost {
			instance := new(T)
			if res := db.FirstOrCreate(instance); res.Error != nil {
				slog.Error("views: layer singleton: first or create", "error", res.Error)
				ctx = ContextWithErrorsAndValues(ctx, nil, map[string]error{
					"_global": fmt.Errorf("failed to load or create singleton: %w", res.Error),
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = context.WithValue(ctx, getters.ContextKeyIn, getters.MapFromStruct(instance))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("views: layer singleton: parse form", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if len(fieldErrors) != 0 {
			for fname, ferr := range fieldErrors {
				slog.Error("views: layer singleton: field error", "field", fname, "error", ferr)
			}
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		regularValues, associationValues := SplitAssociationValues(values)

		instance := new(T)
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.FirstOrCreate(instance).Error; err != nil {
				return err
			}
			if len(regularValues) > 0 {
				id, err := modelPrimaryKeyValue(instance)
				if err != nil {
					return err
				}
				if err := tx.Model(new(T)).Where("id = ?", id).Updates(regularValues).Error; err != nil {
					return err
				}
			}
			return applyAssociationReplacements(tx, instance, associationValues)
		})
		if err != nil {
			slog.Error("views: layer singleton: transaction", "error", err)
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("views: layer singleton: resolve success URL", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		HtmxRedirect(w, r, successUrl, http.StatusSeeOther)
	})
}
