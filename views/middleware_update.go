package views

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		record, ok := ctx.Value(key).(T)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
				rowsAffected, err := updateQuery.Updates(ctx, record)
				if err != nil {
					return err
				}
				fmt.Printf("Updated %d rows\n", rowsAffected)
			}

			return applyAssociationReplacements(tx, record, associationValues)
		})
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if m.SuccessURL != nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		successUrl, err := m.SuccessURL(ctx)
		if err != nil {
			ctx = ContextWithErrorsAndValues(ctx,values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	})
}
