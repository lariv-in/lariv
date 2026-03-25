package p_assignments

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func assignmentsOrderDueQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	if r.URL.Query().Get("sort") != "" {
		return query
	}
	return query.Order("due ASC")
}

func init() {
	lago.RegistryView.Register("assignments.ListView",
		views.ListView[Assignment]("assignments")(
			lago.GetPageView("assignments.AssignmentTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.order_due", assignmentsOrderDueQueryPatcher))

	lago.RegistryView.Register("assignments.DetailView",
		views.DetailView[Assignment]("assignment")(
			lago.GetPageView("assignments.AssignmentDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.CreateView",
		views.CreateView[Assignment](
			lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("assignments.AssignmentCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("assignments.UpdateView",
		views.DetailView[Assignment]("assignment")(
			views.UpdateView[Assignment](
				lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("assignments.AssignmentUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.DeleteView",
		views.DetailView[Assignment]("assignment")(
			views.DeleteView[Assignment](
				lago.GetterRoutePath("assignments.DefaultRoute", nil),
			)(
				lago.GetPageView("assignments.AssignmentDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.SelectView",
		views.ListView[Assignment]("assignments")(
			lago.GetPageView("assignments.AssignmentSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.order_due", assignmentsOrderDueQueryPatcher))
}
