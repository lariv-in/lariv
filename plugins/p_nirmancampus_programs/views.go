package p_nirmancampus_programs

import (
	"net/http"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func fieldDBName[T any](db *gorm.DB, fieldName string) (string, bool) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", false
	}
	if stmt.Schema == nil {
		return "", false
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", false
	}
	return field.DBName, true
}

func queryPatcherUniversity(param string) views.QueryPatcher {
	return func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		getMap, ok := r.Context().Value("$get").(map[string]any)
		if !ok {
			return query
		}

		raw, ok := getMap[param]
		if !ok {
			return query
		}
		value, ok := raw.(string)
		if !ok {
			return query
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return query
		}

		col, ok := fieldDBName[Program](query, "University")
		if !ok {
			return query
		}

		return query.Where(col+" = ?", value)
	}
}

func init() {
	univPatcher := queryPatcherUniversity("University")

	lago.RegistryView.Register("programs.ListView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("nirmancampus_programs.filter_university", univPatcher))

	lago.RegistryView.Register("programs.DetailView",
		views.DetailView[Program]("program")(
			lago.GetPageView("programs.ProgramDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("programs.CreateView",
		views.CreateView[Program](
			lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("programs.ProgramCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("programs.UpdateView",
		views.DetailView[Program]("program")(
			views.UpdateView[Program](
				lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("programs.ProgramUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("programs.DeleteView",
		views.DetailView[Program]("program")(
			views.DeleteView[Program](
				lago.GetterRoutePath("programs.DefaultRoute", nil),
			)(
				lago.GetPageView("programs.ProgramDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("programs.SelectView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("nirmancampus_programs.filter_university", univPatcher))
}
