package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerSingleton manages single-row settings database tables of type T (e.g. system configurations, site-wide parameters).
// It combines the functions of detail loading and updating because there is no primary key parameter passed via URLs.
// Instead, it operates on the single record returned from FirstOrCreate.
//
// On non-POST triggers (e.g. GET), it loads or creates the singleton row, storing its fields map in the request context
// under [getters.ContextKeyIn] to pre-fill template form input values.
// On POST triggers, it parses form parameters, runs GORM Updates inside a transaction, syncs association values, and handles redirects.
//
// Use Cases:
//   - Managing global system settings, administrator configs, or general app configuration forms.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerSingleton[SystemConfig]{
//	            SuccessURL: lago.RoutePath("admin.Dashboard", nil),
//	        },
//	    },
//	}
type LayerSingleton[T any] struct {
	// SuccessURL represents the dynamic Getter resolving to the redirection target URL upon successful singleton updates.
	SuccessURL getters.Getter[string]
}

// Next wraps the downstream HTTP request handlers executing singleton loading or updates.
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
