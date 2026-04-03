package views

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

type MiddlewareSingleton[T any] struct {
	SuccessURL getters.Getter[string]
}

func (m MiddlewareSingleton[T]) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db := ctx.Value("$db").(*gorm.DB)

		if r.Method != http.MethodPost {
			instance := new(T)
			db.FirstOrCreate(instance)
			ctx = context.WithValue(ctx, getters.ContextKeyIn, getters.MapFromStruct(instance))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			ctx = ContextWithErrorsAndValues(ctx, values, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if len(fieldErrors) != 0 {
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		regularValues, associationValues := splitAssociationValues(values)

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
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx = ContextWithErrorsAndValues(ctx, values, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if m.SuccessURL == nil {
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
