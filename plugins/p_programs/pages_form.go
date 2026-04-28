package p_programs

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_students"
	"github.com/lariv-in/lago/plugins/p_teachers"
)

func programFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "programs.ProgramFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{Error: getters.Key[error]("$error.Code"), Children: []components.PageInterface{&components.InputText{Label: "Code", Name: "Code", Required: true, Getter: getters.Key[string]("$in.Code")}}},
			&components.ContainerError{Error: getters.Key[error]("$error.Name"), Children: []components.PageInterface{&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")}}},
			&components.InputText{Label: "Standard / grade", Name: "Standard", Getter: getters.Key[string]("$in.Standard")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 3, Getter: getters.Key[string]("$in.Description")},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
			&components.InputManyToMany[p_students.Student]{
				Label: "Students", Name: "Students", Required: false,
				Getter: getters.Key[[]p_students.Student]("$in.Students"), Display: getters.Key[string]("$in.StudentNo"),
				Url: lago.RoutePath("students.MultiSelectRoute", nil), Placeholder: "Select students…", Classes: "w-full",
			},
			&components.InputManyToMany[p_teachers.Teacher]{
				Label: "Teachers", Name: "Teachers", Required: false,
				Getter: getters.Key[[]p_teachers.Teacher]("$in.Teachers"), Display: getters.Key[string]("$in.Code"),
				Url: lago.RoutePath("teachers.MultiSelectRoute", nil), Placeholder: "Select teachers…", Classes: "w-full",
			},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("programs.ProgramDeleteForm")
	lago.RegistryPage.Register("programs.ProgramCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "programs.ProgramMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("programs.ProgramCreateForm"), ActionURL: lago.RoutePath("programs.CreateRoute", nil), Children: []components.PageInterface{
				&components.FormComponent[Program]{Attr: getters.FormBubbling(getters.Static("programs.ProgramCreateForm")), Title: "Create Program", ChildrenInput: []components.PageInterface{programFormFields()}, ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save Program"}}},
			}},
		},
	})
	lago.RegistryPage.Register("programs.ProgramUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "programs.ProgramDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("programs.ProgramUpdateForm"), ActionURL: lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}), Children: []components.PageInterface{
				&components.FormComponent[Program]{Getter: getters.Key[Program]("program"), Attr: getters.FormBubbling(getters.Static("programs.ProgramUpdateForm")), Title: "Edit Program", ChildrenInput: []components.PageInterface{programFormFields()}, ChildrenAction: []components.PageInterface{
					&components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
						&components.ButtonSubmit{Label: "Save Program"},
						&components.ButtonModalForm{Label: "Delete", Icon: "trash", Name: deleteFormName, Url: lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}), FormPostURL: lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}), ModalUID: "program-delete-modal", Classes: "btn-error"},
					}},
				}},
			}},
		},
	})
}
