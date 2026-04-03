package views

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

type MiddlewareCreate[T any] struct {
	SuccessURL   getters.Getter[string]
	FormPatchers FormPatchers
}

func (m MiddlewareCreate[T]) Next(view View, next http.Handler) http.Handler {
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
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		id := uint(reflect.ValueOf(*record).FieldByName("ID").Uint())
		ctx = context.WithValue(ctx, "$id", id)
		if m.SuccessURL != nil {
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
