package p_assignments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/registry"
)

func assignmentFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assignments.AssignmentFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_courses.Course]{Label: "Course", Name: "CourseID", Required: true, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Key[uint]("$in.CourseID"))},
			&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 4, Getter: getters.Key[string]("$in.Description")},
			&components.InputDatetime{Label: "Release at", Name: "ReleaseAt", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.ReleaseAt"))},
			&components.InputDatetime{Label: "Due at", Name: "DueAt", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.DueAt"))},
			&components.InputNumber[int]{Label: "Total marks", Name: "TotalMarks", Required: true, Getter: getters.Key[int]("$in.TotalMarks")},
			&components.InputForeignKey[p_semesters.Semester]{Label: "Semester (optional)", Name: "SemesterID", Required: false, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Semester…", Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID")))},
			&components.InputSelect[string]{Label: "Type", Name: "AssignmentType", Required: true, Choices: getters.Static(AssignmentTypeChoices), Getter: registry.PairFromGetter(getters.Key[string]("$in.AssignmentType"), AssignmentTypeChoices)},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("assignments.AssignmentDeleteForm")
	lago.RegistryPage.Register("assignments.AssignmentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assignments.AssignmentMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assignments.AssignmentCreateForm"),
				ActionURL: lago.RoutePath("assignments.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Assignment]{
						Attr:           getters.FormBubbling(getters.Static("assignments.AssignmentCreateForm")),
						Title:          "Create assignment",
						ChildrenInput:  []components.PageInterface{assignmentFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("assignments.AssignmentUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assignments.AssignmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assignments.AssignmentUpdateForm"),
				ActionURL: lago.RoutePath("assignments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Assignment]{
						Getter:        getters.Key[Assignment]("assignment"),
						Attr:          getters.FormBubbling(getters.Static("assignments.AssignmentUpdateForm")),
						Title:         "Edit assignment",
						ChildrenInput: []components.PageInterface{assignmentFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("assignments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))}),
										FormPostURL: lago.RoutePath("assignments.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))}),
										ModalUID:    "assignment-delete-modal", Classes: "btn-error",
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
