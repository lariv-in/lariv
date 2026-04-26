package p_nirmancampus_studentpayments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerStudentsMenuPaymentsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerStudentsMenuPaymentsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("All Payments"),
			Url:   lago.RoutePath("studentpayments.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("studentpayments.PaymentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Payment #%d", getters.Any(getters.Key[uint]("payment.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all payments"),
			Url:   lago.RoutePath("studentpayments.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Payment detail"),
				Url: lago.RoutePath("studentpayments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("payment.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit payment"),
				Url: lago.RoutePath("studentpayments.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("payment.ID")),
				}),
			},
		},
	})
}
