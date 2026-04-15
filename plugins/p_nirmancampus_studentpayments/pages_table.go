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

func paidAtDisplay(getter getters.Getter[*time.Time]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := getter(ctx)
		if err != nil || t == nil || t.IsZero() {
			return "", nil
		}
		loc, _ := ctx.Value("$tz").(*time.Location)
		if loc == nil {
			loc = time.UTC
		}
		return t.In(loc).Format(time.DateOnly), nil
	}
}

func paymentTableColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Student",
			Name:  "Student.Name",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: getters.Format(
						"%s (%s)",
						getters.Any(getters.Key[string]("$row.Student.Name")),
						getters.Any(getters.Key[string]("$row.Student.StudentNo")),
					),
				},
			},
		},
		{
			Label: "Amount",
			Name:  "Amount",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$row.Amount"))),
				},
			},
		},
		{
			Label: "Method",
			Name:  "PaymentMethod",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: registry.PairValueFromKey(getters.Key[string]("$row.PaymentMethod"), PaymentMethodChoices),
				},
			},
		},
		{
			Label: "Paid on",
			Name:  "PaidAt",
			Children: []components.PageInterface{
				&components.FieldText{Getter: paidAtDisplay(getters.Key[*time.Time]("$row.PaidAt"))},
			},
		},
		{
			Label: "Transaction ID",
			Name:  "TransactionID",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.TransactionID")},
			},
		},
	}
}

func registerFilterPages() {
	lago.RegistryPage.Register("studentpayments.PaymentFilter", &components.FormComponent[Payment]{
		Attr: getters.FormBoostedGet(lago.RoutePath("studentpayments.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputSelect[string]{
				Label:   "Method",
				Name:    "PaymentMethod",
				Choices: getters.Static(PaymentMethodChoices),
				Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
					s, err := getters.Key[string]("$get.PaymentMethod")(ctx)
					if err != nil || s == "" {
						return registry.Pair[string, string]{}, nil
					}
					if p, ok := registry.PairFromPairs(s, PaymentMethodChoices); ok {
						return p, nil
					}
					return registry.Pair[string, string]{Key: s, Value: s}, nil
				},
			},
			&components.InputForeignKey[p_nirmancampus_students.Student]{
				Label:       "Student",
				Name:        "StudentID",
				Url:         lago.RoutePath("students.SelectRoute", nil),
				Placeholder: "Filter by student…",
				Display:     getters.Key[string]("$in.StudentNo"),
				Getter: getters.Association[p_nirmancampus_students.Student](
					getters.Key[uint]("$get.StudentID"),
				),
			},
			&components.InputText{
				Label:  "Transaction ID",
				Name:   "TransactionID",
				Getter: getters.Key[string]("$get.TransactionID"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func registerTablePages() {
	createFormName := getters.Static("studentpayments.PaymentCreateForm")
	lago.RegistryPage.Register("studentpayments.PaymentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Payment]{
				Page:    components.Page{Key: "studentpayments.PaymentTableBody"},
				UID:     "studentpayments-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Payment]]("studentpayments"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "studentpayments.PaymentFilter"},
					},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Name:        createFormName,
						Url:         lago.RoutePath("studentpayments.CreateRoute", nil),
						FormPostURL: lago.RoutePath("studentpayments.CreateRoute", nil),
						ModalUID:    "studentpayments-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
						Attr:        getters.ModalRefreshList(getters.Static(""), getters.Static("#studentpayments-table")),
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("studentpayments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$row.ID")),
				})),
				Columns: paymentTableColumns(),
			},
		},
	})
}
