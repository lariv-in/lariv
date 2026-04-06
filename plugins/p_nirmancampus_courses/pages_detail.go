package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("courses.CourseDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Course]{
				Getter: getters.Key[Course]("course"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{
							Key: "courses.CourseDetailContent",
						},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title:   "Type",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.CourseType")},
								},
							},
							&components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
								},
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Description")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "course-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this course?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
