package views

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

// MiddlewareUpdate handles row updates for type T on POST requests.
// On non-POST methods it passes through to next unchanged.
//
// It expects the record to already be in the context under Key, typically
// placed there by a preceding MiddlewareDetail. On POST it parses the view's
// form, applies FormPatchers, then within a transaction updates the record's
// columns and replaces any many-to-many associations.
//
// QueryPatchers are applied to the UPDATE query, allowing callers to add
// tenant filters or scopes.
//
// If SuccessURL is set, a successful update redirects to the resolved URL.
// If SuccessURL is nil, next is called so a downstream handler can decide
// the response.
//
// Form and field errors go into getters.ContextKeyError under "_form" or the
// field name. Internal errors (missing context record, getter failures) go
// under "_global". In all error cases next is called, never a raw HTTP
// response.
type MiddlewareUpdate[T any] struct {
	Key           getters.Getter[string]
	SuccessURL    getters.Getter[string]
	FormPatchers  FormPatchers
	QueryPatchers QueryPatchers[T]
}

func (m MiddlewareUpdate[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		values, fieldErrors = m.FormPatchers.Apply(view, r, values, fieldErrors)
		if len(fieldErrors) != 0 {
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db := ctx.Value("$db").(*gorm.DB)
		regularValues, associationValues := splitAssociationValues(values)
		key, err := m.Key(ctx)
		if err != nil {
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_global": fmt.Errorf("failed to resolve context key: %w", err),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
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
				updateQuery := gorm.G[T](tx).Where("id = ?", id)
				updateQuery = m.QueryPatchers.Apply(view, r, updateQuery)
				_, err := updateQuery.Updates(ctx, record)
				if err != nil {
					return err
				}
			}

			return applyAssociationReplacements(tx, record, associationValues)
		})
		if err != nil {
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
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	})
}
