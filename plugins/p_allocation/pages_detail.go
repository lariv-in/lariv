package p_allocation

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "allocation.CourseTeacherAssignmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[CourseTeacherAssignment]{
				Getter: getters.Key[CourseTeacherAssignment]("allocation"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "allocation.CourseTeacherAssignmentDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Role")},
							&components.LabelInline{Title: "Teacher ID", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.TeacherID")))},
							}},
							&components.LabelInline{Title: "Course ID", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.CourseID")))},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("allocation.CourseTeacherAssignmentDeleteForm", &components.Modal{
		UID: "allocation-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this allocation?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
