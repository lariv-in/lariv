package p_syllabus

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("syllabus.SyllabusTopicDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "syllabus.SyllabusTopicDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[SyllabusTopic]{
				Getter: getters.Key[SyllabusTopic]("syllabus_topic"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "syllabus.SyllabusTopicDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Course ID", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.CourseID")))},
							}},
							&components.LabelInline{Title: "Sort order", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.SortOrder")))},
							}},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("syllabus.SyllabusTopicDeleteForm", &components.Modal{
		UID: "syllabus-topic-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title: "Confirm deletion", Message: "Delete this syllabus topic?",
				Attr: getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
