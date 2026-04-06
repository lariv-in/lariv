package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func registerDetailPages() {
	lago.RegistryPage.Register("assignmentsubmissions.Detail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AssignmentSubmission]{
				Getter: getters.Key[AssignmentSubmission]("assignmentsubmission"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "assignmentsubmissions.DetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.AssignmentTitle")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Course.Name")},
							&components.LabelInline{
								Title: "Submission status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.SubmissionStatus")},
								},
							},
							&components.LabelInline{
								Page:  components.Page{Roles: []string{"admin", "superuser"}},
								Title: "Marks",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d / %d", getters.Any(getters.Key[int]("$in.Marks")), getters.Any(getters.Key[int]("$in.MaxMarks")))},
								},
							},
							&components.LabelInline{
								Title: "Academic record",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%s · %s", getters.Any(getters.Key[string]("$in.AcademicRecord.Student.StudentNo")), getters.Any(getters.Key[string]("$in.AcademicRecord.Program.Name")))},
								},
							},
							&components.LabelInline{
								Title: "Assets",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{VNode: getters.Key[[]p_filesystem.VNode]("$in.Assets")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentsubmissions.DeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "assignmentsubmission-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this submission?",
				Attr:    getters.FormBubbling(nil),
			},
		},
	})
}
