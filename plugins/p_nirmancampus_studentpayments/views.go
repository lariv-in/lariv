package p_nirmancampus_studentpayments

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// studentPaymentsAdminRoleLayer limits create/update/delete to admin;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var studentPaymentsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	lago.RegistryView.Register("studentpayments.ListView",
		lago.GetPageView("studentpayments.PaymentTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentpayments.list", views.LayerList[Payment]{
				Key: getters.Static("studentpayments"),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.preload", Value: views.QueryPatcherPreload[Payment]{Fields: []string{"Student"}}},
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.list_order", Value: PaymentListOrder},
				},
			}),
	)

	lago.RegistryView.Register("studentpayments.DetailView",
		lago.GetPageView("studentpayments.PaymentDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentpayments.detail", views.LayerDetail[Payment]{
				Key:          getters.Static("payment"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.preload", Value: views.QueryPatcherPreload[Payment]{Fields: []string{"Student"}}},
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("studentpayments.CreateView",
		lago.GetPageView("studentpayments.PaymentCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentpayments.admin_role", studentPaymentsAdminRoleLayer).
			WithLayer("studentpayments.create_query_defaults", paymentCreateQueryDefaultsLayer{}).
			WithLayer("studentpayments.create", views.LayerCreate[Payment]{
				SuccessURL: lago.RoutePath("studentpayments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}),
	)

	lago.RegistryView.Register("studentpayments.UpdateView",
		lago.GetPageView("studentpayments.PaymentUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentpayments.admin_role", studentPaymentsAdminRoleLayer).
			WithLayer("studentpayments.detail", views.LayerDetail[Payment]{
				Key:          getters.Static("payment"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.preload", Value: views.QueryPatcherPreload[Payment]{Fields: []string{"Student"}}},
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
				},
			}).
			WithLayer("studentpayments.update", views.LayerUpdate[Payment]{
				Key: getters.Static("payment"),
				SuccessURL: lago.RoutePath("studentpayments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("payment.ID")),
				}),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
				},
			}),
	)

	lago.RegistryView.Register("studentpayments.DeleteView",
		lago.GetPageView("studentpayments.PaymentDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("studentpayments.admin_role", studentPaymentsAdminRoleLayer).
			WithLayer("studentpayments.detail", views.LayerDetail[Payment]{
				Key:          getters.Static("payment"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.preload", Value: views.QueryPatcherPreload[Payment]{Fields: []string{"Student"}}},
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
				},
			}).
			WithLayer("studentpayments.delete", views.LayerDelete[Payment]{
				Key:        getters.Static("payment"),
				SuccessURL: lago.RoutePath("studentpayments.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Payment]{
					registry.Pair[string, views.QueryPatcher[Payment]]{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
				},
			}),
	)
}
