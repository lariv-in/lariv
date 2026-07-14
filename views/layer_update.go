package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// LayerUpdate handles database row updates for type T on incoming POST requests.
// It expects the target record to already reside inside the context under Key (typically loaded by a preceding [LayerDetail] layer).
//
// On intercepting a POST action, it parses form values, runs registered [FormPatchers],
// updates columns recursively inside a GORM transaction, syncs association properties, and handles redirection links.
//
// Use Cases:
//   - Driving resource edit/update views (e.g., editing user profiles, updating product information).
//   - Triggering background jobs or side-effects upon record update using transaction commit hooks.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.LayerDetail[User]{Key: getters.Static("$record")},
//	        views.LayerUpdate[User]{
//	            Key:        getters.Static("$record"),
//	            SuccessURL: lago.RoutePath("users.List", nil),
//	        },
//	    },
//	}
type LayerUpdate[T any] struct {
	// Key represents the context key pointing to the target loaded record to update.
	Key getters.Getter[string]
	// SuccessURL represents the dynamic Getter resolving to the redirection target URL upon successful updates.
	SuccessURL getters.Getter[string]
	// FormPatchers represents the collection of patch middleware rules to apply to form maps before updates.
	FormPatchers FormPatchers
	// QueryPatchers represents the slice of query modifications to restrict update scopes.
	QueryPatchers QueryPatchers[T]
}

// Next wraps the downstream HTTP request handlers executing row updates.
func (m LayerUpdate[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("views: layer update: parse form", "error", err)
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
				slog.Error("views: layer update: field error", "field", fname, "error", ferr)
			}
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			slog.Error("views: layer update: db from context", "error", dberr)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_global": dberr,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		regularValues, associationValues := SplitAssociationValues(values)
		key, err := m.Key(ctx)
		if err != nil {
			slog.Error("views: layer update: resolve context key", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_global": fmt.Errorf("failed to resolve context key: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
			slog.Error("views: layer update: record missing from context", "key", key)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_global": fmt.Errorf("record not found in context"),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		id := uint(reflect.ValueOf(record).FieldByName("ID").Uint())
		err = db.Transaction(func(tx *gorm.DB) error {
			if len(regularValues) > 0 {
				if err := PopulateFromMap(&record, regularValues); err != nil {
					return err
				}
				// GORM's update association callbacks require an addressable model; the
				// generic chain defaults to Model(zero T), which is not. Bind the loaded
				// row as Model before Updates (Dest may remain the struct value).
				updateQuery := gorm.G[T](tx).Scopes(func(stmt *gorm.Statement) {
					stmt.Model = &record
				}).Where("id = ?", id)
				updateQuery = m.QueryPatchers.Apply(view, r, updateQuery)
				updateQuery.Updates(ctx, record)
			}

			return applyAssociationReplacements(tx, &record, associationValues)
		})
		if err != nil {
			slog.Error("views: layer update: transaction", "error", err)
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if hook, ok := any(&record).(TxCommitHook); ok {
			hook.AfterTxCommit(db)
		}

		if m.SuccessURL == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			slog.Error("views: layer update: resolve success URL", "error", err)
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		HtmxRedirect(w, r, successUrl, http.StatusSeeOther)
	})
}
