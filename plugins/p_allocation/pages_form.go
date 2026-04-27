package p_allocation

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_teachers"
)

func allocationFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "allocation.CourseTeacherAssignmentFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_teachers.Teacher]{Label: "Teacher", Name: "TeacherID", Required: true, Url: lago.RoutePath("teachers.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select teacher...", Getter: getters.Association[p_teachers.Teacher](getters.Key[uint]("$in.TeacherID"))},
			&components.InputForeignKey[p_courses.Course]{Label: "Course", Name: "CourseID", Required: true, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Key[uint]("$in.CourseID"))},
			&components.InputText{Label: "Role", Name: "Role", Getter: getters.Key[string]("$in.Role")},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("allocation.CourseTeacherAssignmentDeleteForm")
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "allocation.CourseTeacherAssignmentMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("allocation.CourseTeacherAssignmentCreateForm"),
				ActionURL: lago.RoutePath("allocation.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[CourseTeacherAssignment]{
						Attr:           getters.FormBubbling(getters.Static("allocation.CourseTeacherAssignmentCreateForm")),
						Title:          "Create allocation",
						ChildrenInput:  []components.PageInterface{allocationFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "allocation.CourseTeacherAssignmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("allocation.CourseTeacherAssignmentUpdateForm"),
				ActionURL: lago.RoutePath("allocation.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[CourseTeacherAssignment]{
						Getter:        getters.Key[CourseTeacherAssignment]("allocation"),
						Attr:          getters.FormBubbling(getters.Static("allocation.CourseTeacherAssignmentUpdateForm")),
						Title:         "Edit allocation",
						ChildrenInput: []components.PageInterface{allocationFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("allocation.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))}),
										FormPostURL: lago.RoutePath("allocation.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))}),
										ModalUID:    "allocation-delete-modal", Classes: "btn-error",
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
