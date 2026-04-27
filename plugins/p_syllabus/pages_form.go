package p_syllabus

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
)

func syllabusTopicFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "syllabus.SyllabusTopicFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.CourseID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_courses.Course]{Label: "Course", Name: "CourseID", Required: true, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Key[uint]("$in.CourseID"))},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Title"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
				},
			},
			&components.InputNumber[uint]{Label: "Sort order", Name: "SortOrder", Required: false, Getter: getters.Key[uint]("$in.SortOrder")},
			&components.InputText{Label: "Book", Name: "Book", Getter: getters.Key[string]("$in.Book")},
			&components.InputText{Label: "Page range", Name: "PageRange", Getter: getters.Key[string]("$in.PageRange")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 3, Getter: getters.Key[string]("$in.Description")},
			&components.InputCheckbox{Label: "Completed", Name: "IsCompleted", Getter: getters.Key[bool]("$in.IsCompleted")},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("syllabus.SyllabusTopicDeleteForm")

	lago.RegistryPage.Register("syllabus.SyllabusTopicCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "syllabus.SyllabusTopicMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("syllabus.SyllabusTopicCreateForm"),
				ActionURL: lago.RoutePath("syllabus.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[SyllabusTopic]{
						Attr:           getters.FormBubbling(getters.Static("syllabus.SyllabusTopicCreateForm")),
						Title:          "Create syllabus topic",
						ChildrenInput:  []components.PageInterface{syllabusTopicFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("syllabus.SyllabusTopicUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "syllabus.SyllabusTopicDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("syllabus.SyllabusTopicUpdateForm"),
				ActionURL: lago.RoutePath("syllabus.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("syllabus_topic.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[SyllabusTopic]{
						Getter:        getters.Key[SyllabusTopic]("syllabus_topic"),
						Attr:          getters.FormBubbling(getters.Static("syllabus.SyllabusTopicUpdateForm")),
						Title:         "Edit syllabus topic",
						ChildrenInput: []components.PageInterface{syllabusTopicFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: deleteFormName,
										Url:         lago.RoutePath("syllabus.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("syllabus_topic.ID"))}),
										FormPostURL: lago.RoutePath("syllabus.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("syllabus_topic.ID"))}),
										ModalUID:    "syllabus-topic-delete-modal", Classes: "btn-error",
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
