package p_nirmancampus_studentpayments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("studentpayments.PaymentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentpayments.PaymentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Payment]{
				Getter: getters.Key[Payment]("payment"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "studentpayments.PaymentDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format(
									"%s · %s",
									getters.Any(getters.Key[string]("$in.Student.Name")),
									getters.Any(getters.Format("%.2f", getters.Any(getters.Key[float64]("$in.Amount")))),
								),
							},
							&components.FieldSubtitle{
								Getter: registry.PairValueFromKey(getters.Key[string]("$in.PaymentMethod"), PaymentMethodChoices),
							},
							&components.LabelInline{
								Title: "Student record",
								Children: []components.PageInterface{
									&components.FieldLink{
										Href: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("$in.Student.ID")),
										}),
										Label: getters.Format(
											"%s (%s)",
											getters.Any(getters.Key[string]("$in.Student.Name")),
											getters.Any(getters.Key[string]("$in.Student.StudentNo")),
										),
										Classes: "link link-primary",
									},
								},
							},
							&components.LabelInline{
								Title: "Transaction ID",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.TransactionID")},
								},
							},
							&components.LabelInline{
								Title: "Paid on",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$in.PaidAt"))},
								},
							},
							&components.LabelNewline{
								Title: "Remarks",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Remarks")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentpayments.PaymentDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "studentpayments-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this payment?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
