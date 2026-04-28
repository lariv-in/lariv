package p_finances

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_students"
)

func studentChargeFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "finances.StudentChargeFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_students.Student]{Label: "Student", Name: "StudentID", Required: true, Url: lago.RoutePath("students.SelectRoute", nil), Display: getters.Key[string]("$in.StudentNo"), Placeholder: "Select student...", Getter: getters.Association[p_students.Student](getters.Key[uint]("$in.StudentID"))},
			&components.InputNumber[int64]{Label: "Amount (cents)", Name: "AmountCents", Required: true, Getter: getters.Key[int64]("$in.AmountCents")},
			&components.InputText{Label: "Description", Name: "Description", Getter: getters.Key[string]("$in.Description")},
			&components.InputText{Label: "Purpose", Name: "Purpose", Getter: getters.Key[string]("$in.Purpose")},
			&components.InputTextarea{Label: "Remarks", Name: "Remarks", Rows: 2, Getter: getters.Key[string]("$in.Remarks")},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
			&components.InputDate{Label: "Due on", Name: "DueOn", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.DueOn"))},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("finances.StudentChargeDeleteForm")
	lago.RegistryPage.Register("finances.StudentChargeCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "finances.StudentChargeMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("finances.StudentChargeCreateForm"),
				ActionURL: lago.RoutePath("finances.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[StudentCharge]{
						Attr:           getters.FormBubbling(getters.Static("finances.StudentChargeCreateForm")),
						Title:          "Create charge",
						ChildrenInput:  []components.PageInterface{studentChargeFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("finances.StudentChargeUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "finances.StudentChargeDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("finances.StudentChargeUpdateForm"),
				ActionURL: lago.RoutePath("finances.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[StudentCharge]{
						Getter:        getters.Key[StudentCharge]("student_charge"),
						Attr:          getters.FormBubbling(getters.Static("finances.StudentChargeUpdateForm")),
						Title:         "Edit charge",
						ChildrenInput: []components.PageInterface{studentChargeFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("finances.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))}),
										FormPostURL: lago.RoutePath("finances.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))}),
										ModalUID:    "student-charge-delete-modal", Classes: "btn-error",
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
