package p_nirmancampus_studentpayments

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
)

func paymentMethodPairFromIn(ctx context.Context) (registry.Pair[string, string], error) {
	s, err := getters.Key[string]("$in.PaymentMethod")(ctx)
	if err != nil || s == "" {
		if p, ok := registry.PairFromPairs("cash", PaymentMethodChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: "cash", Value: "Cash"}, nil
	}
	if p, ok := registry.PairFromPairs(s, PaymentMethodChoices); ok {
		return p, nil
	}
	return registry.Pair[string, string]{Key: s, Value: s}, nil
}

func paymentCreateFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "studentpayments.PaymentCreateFormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.StudentID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_students.Student]{
								Label:       "Student",
								Name:        "StudentID",
								Required:    true,
								Url:         lago.RoutePath("students.SelectRoute", nil),
								Display:     getters.Key[string]("$in.StudentNo"),
								Placeholder: "Select a student…",
								Getter: getters.Association[p_nirmancampus_students.Student](
									getters.Key[uint]("$in.StudentID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Amount"),
						Children: []components.PageInterface{
							&components.InputNumber[float64]{
								Label:    "Amount",
								Name:     "Amount",
								Required: true,
								Getter:   getters.Key[float64]("$in.Amount"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.PaymentMethod"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Payment method",
								Name:     "PaymentMethod",
								Required: true,
								Choices:  getters.Static(PaymentMethodChoices),
								Getter:   paymentMethodPairFromIn,
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.TransactionID"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Transaction ID",
								Name:   "TransactionID",
								Getter: getters.Key[string]("$in.TransactionID"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.PaidAt"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:  "Paid on",
								Name:   "PaidAt",
								Getter: getters.Deref(getters.Key[*time.Time]("$in.PaidAt")),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Remarks"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Remarks",
								Name:   "Remarks",
								Rows:   3,
								Getter: getters.Key[string]("$in.Remarks"),
							},
						},
					},
				},
			},
		},
	}
}

func paymentUpdateFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "studentpayments.PaymentUpdateFormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Student",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format(
									"%s (%s)",
									getters.Any(getters.Key[string]("$in.Student.Name")),
									getters.Any(getters.Key[string]("$in.Student.StudentNo")),
								),
							},
						},
					},
					&components.InputForeignKey[p_nirmancampus_students.Student]{
						Hidden: true,
						Name:   "StudentID",
						Getter: getters.Association[p_nirmancampus_students.Student](
							getters.Key[uint]("$in.StudentID"),
						),
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Amount"),
						Children: []components.PageInterface{
							&components.InputNumber[float64]{
								Label:    "Amount",
								Name:     "Amount",
								Required: true,
								Getter:   getters.Key[float64]("$in.Amount"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.PaymentMethod"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Payment method",
								Name:     "PaymentMethod",
								Required: true,
								Choices:  getters.Static(PaymentMethodChoices),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.PaymentMethod")(ctx)
									if err != nil || s == "" {
										return registry.Pair[string, string]{}, nil
									}
									if p, ok := registry.PairFromPairs(s, PaymentMethodChoices); ok {
										return p, nil
									}
									return registry.Pair[string, string]{Key: s, Value: s}, nil
								},
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.TransactionID"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Transaction ID",
								Name:   "TransactionID",
								Getter: getters.Key[string]("$in.TransactionID"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.PaidAt"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:  "Paid on",
								Name:   "PaidAt",
								Getter: getters.Deref(getters.Key[*time.Time]("$in.PaidAt")),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Remarks"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Remarks",
								Name:   "Remarks",
								Rows:   3,
								Getter: getters.Key[string]("$in.Remarks"),
							},
						},
					},
				},
			},
		},
	}
}

func registerFormPages() {
	createFormName := getters.Static("studentpayments.PaymentCreateForm")
	updateFormName := getters.Static("studentpayments.PaymentUpdateForm")
	deleteFormName := getters.Static("studentpayments.PaymentDeleteForm")

	lago.RegistryPage.Register("studentpayments.PaymentCreateForm", &components.Modal{
		Page: components.Page{
			Key:   "studentpayments.PaymentCreateModal",
			Roles: []string{"admin", "superuser"},
		},
		UID: "studentpayments-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[Payment]{
				Attr: getters.FormBubbling(createFormName),

				Title:    "Record payment",
				Subtitle: "Log a payment for a student.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					paymentCreateFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save payment", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentpayments.PaymentUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentpayments.PaymentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      updateFormName,
				ActionURL: lago.RoutePath("studentpayments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("payment.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Payment]{
						Getter: getters.Key[Payment]("payment"),
						Attr:   getters.FormBubbling(updateFormName),

						Title:    "Edit payment",
						Subtitle: "Update amount, method, or references.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							paymentUpdateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save payment"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        deleteFormName,
												Url:         lago.RoutePath("studentpayments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("payment.ID"))}),
												FormPostURL: lago.RoutePath("studentpayments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("payment.ID"))}),
												ModalUID:    "studentpayments-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
