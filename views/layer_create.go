package views

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/lariv-in/lariv/getters"
	"gorm.io/gorm"
)

// LayerCreate handles database record insertion transactions for type T on incoming POST requests.
// On non-POST requests, it passes execution to the next handler downstream.
//
// On a POST request, it parses the view's form parameters, executes registered [FormPatchers],
// populates a new record instance of type T with the non-association values, inserts it inside a transaction,
// and syncs many-to-many associations. Upon successful creation, the new primary key is stored as "$id" on the request context.
// If SuccessURL is set, a successful operation redirects the browser to the resolved path.
// Otherwise, it passes execution downstream with the enriched context.
//
// Use Cases:
//   - Handling model creation form submissions (e.g. creating users, items, or configurations).
//   - Triggering background jobs or side-effects upon record creation using transaction commit hooks.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        &views.PathLayer{Names: []string{"id"}},
//	        views.LayerCreate[User]{
//	            SuccessURL: lariv.RoutePath("users.List", nil),
//	            FormPatchers: views.FormPatchers{
//	                registry.NewPair("author", AuthorPatcher{}),
//	            },
//	        },
//	    },
//	}
type LayerCreate[T any] struct {
	// SuccessURL represents the dynamic Getter resolving to the redirection target URL upon successful record creation.
	SuccessURL getters.Getter[string]
	// FormPatchers represents the collection of patch middleware rules to apply to form maps before database insertion.
	FormPatchers FormPatchers
}

// Next wraps the downstream HTTP request handlers executing row creation on POST triggers.
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
