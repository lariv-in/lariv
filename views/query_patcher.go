package views

import (
	"net/http"

	"gorm.io/gorm"
)

type QueryPatcher = func(view *View, r *http.Request, db *gorm.DB) *gorm.DB

func QueryPatcherPreload(field string) QueryPatcher {
	return func(v *View, r *http.Request, query *gorm.DB) *gorm.DB {
		return query.Preload(field)
	}
}

func QueryPatcherOrderBy(order string) QueryPatcher {
	return func(v *View, r *http.Request, query *gorm.DB) *gorm.DB {
		return query.Order(order)
	}
}
