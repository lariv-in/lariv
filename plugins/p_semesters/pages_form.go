package p_semesters

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func semesterFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "semesters.SemesterFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{Error: getters.Key[error]("$error.Code"), Children: []components.PageInterface{&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.Key[string]("$in.Code")}}},
			&components.ContainerError{Error: getters.Key[error]("$error.Name"), Children: []components.PageInterface{&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")}}},
			&components.ContainerError{Error: getters.Key[error]("$error.Start"), Children: []components.PageInterface{&components.InputDatetime{Label: "Start", Name: "Start", Required: true, Getter: getters.Key[time.Time]("$in.Start")}}},
			&components.ContainerError{Error: getters.Key[error]("$error.End"), Children: []components.PageInterface{&components.InputDatetime{Label: "End", Name: "End", Required: true, Getter: getters.Key[time.Time]("$in.End")}}},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("semesters.SemesterDeleteForm")
	lago.RegistryPage.Register("semesters.SemesterCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "semesters.SemesterMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("semesters.SemesterCreateForm"), ActionURL: lago.RoutePath("semesters.CreateRoute", nil), Children: []components.PageInterface{
				&components.FormComponent[Semester]{Attr: getters.FormBubbling(getters.Static("semesters.SemesterCreateForm")), Title: "Create Semester", ChildrenInput: []components.PageInterface{semesterFormFields()}, ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save Semester"}}},
			}},
		},
	})
	lago.RegistryPage.Register("semesters.SemesterUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "semesters.SemesterDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("semesters.SemesterUpdateForm"), ActionURL: lago.RoutePath("semesters.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))}), Children: []components.PageInterface{
				&components.FormComponent[Semester]{Getter: getters.Key[Semester]("semester"), Attr: getters.FormBubbling(getters.Static("semesters.SemesterUpdateForm")), Title: "Edit Semester", ChildrenInput: []components.PageInterface{semesterFormFields()}, ChildrenAction: []components.PageInterface{
					&components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
						&components.ButtonSubmit{Label: "Save Semester"},
						&components.ButtonModalForm{Label: "Delete", Icon: "trash", Name: deleteFormName, Url: lago.RoutePath("semesters.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))}), FormPostURL: lago.RoutePath("semesters.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))}), ModalUID: "semester-delete-modal", Classes: "btn-error"},
					}},
				}},
			}},
		},
	})
}
