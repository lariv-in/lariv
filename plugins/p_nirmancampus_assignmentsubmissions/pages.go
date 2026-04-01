package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignmentsubmissions.Menu", &components.SidebarMenu{
		Title: getters.Static("Assignment Submissions"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Submissions"),
				Url:   lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("assignmentsubmissions.DetailMenu", &components.SidebarMenu{
		Title: getters.Format("Submission: %s", getters.Any(getters.Key[string]("assignmentsubmission.AssignmentTitle"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all submissions"),
			Url:   lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Submission detail"),
				Url: lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit submission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete submission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("assignmentsubmissions.Filter", &components.FormComponent[AssignmentSubmission]{
		Url:    lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Assignment title",
				Name:   "AssignmentTitle",
				Getter: getters.Key[string]("$get.AssignmentTitle"),
			},
			&components.InputText{
				Label:  "Submission status",
				Name:   "SubmissionStatus",
				Getter: getters.Key[string]("$get.SubmissionStatus"),
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func assignmentSubmissionFormFields() *components.ContainerColumn {
	return &components.ContainerColumn{
		Page: components.Page{Key: "assignmentsubmissions.FormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.AssignmentTitle"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Assignment title",
								Name:     "AssignmentTitle",
								Required: true,
								Getter:   getters.Key[string]("$in.AssignmentTitle"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.SubmissionStatus"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Submission status",
								Name:     "SubmissionStatus",
								Required: true,
								Getter:   getters.Key[string]("$in.SubmissionStatus"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.MaxMarks"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Max marks",
								Name:     "MaxMarks",
								Required: true,
								Getter:   getters.Key[int]("$in.MaxMarks"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Marks"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Marks",
								Name:     "Marks",
								Required: true,
								Getter:   getters.Key[int]("$in.Marks"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.CourseID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_courses.Course]{
								Label:       "Course",
								Name:        "CourseID",
								Required:    true,
								Url:         lago.GetterRoutePath("courses.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select a course...",
								Getter: getters.Association[p_nirmancampus_courses.Course](
									getters.Key[uint]("$in.CourseID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.AcademicRecordID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_academicrecords.AcademicRecord]{
								Label:       "Academic record",
								Name:        "AcademicRecordID",
								Required:    true,
								Url:         lago.GetterRoutePath("academicrecords.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Student.StudentNo"),
								Placeholder: "Select an academic record...",
								Getter: getters.Association[p_nirmancampus_academicrecords.AcademicRecord](
									getters.Key[uint]("$in.AcademicRecordID"),
								),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Assets"),
				Children: []components.PageInterface{
					&p_filesystem.InputMultiVNode{
						Label:            "Assets",
						Name:             "Assets",
						VNode:            getters.Key[[]p_filesystem.VNode]("$in.Assets"),
						AllowedFiletypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".webp", ".doc", ".docx"},
					},
				},
			},
		},
	}
}

func assignmentSubmissionCreateURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.GetterRoutePath("assignmentsubmissions.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("assignmentsubmissions.FormFields", assignmentSubmissionFormFields())

	lago.RegistryPage.Register("assignmentsubmissions.CreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AssignmentSubmission]{
				Url:      lago.GetterRoutePath("assignmentsubmissions.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create submission",
				Subtitle: "Create a new assignment submission",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentSubmissionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save submission"},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentsubmissions.UpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AssignmentSubmission]{
				Getter: getters.Key[AssignmentSubmission]("assignmentsubmission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit submission",
				Subtitle: "Update assignment submission details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentSubmissionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save submission"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("assignmentsubmissions.Table", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AssignmentSubmission]{
				Page:            components.Page{Key: "assignmentsubmissions.TableBody"},
				UID:             "assignment-submissions-table",
				Classes:         "w-full",
				Data:            getters.Key[components.ObjectList[AssignmentSubmission]]("assignmentsubmissions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "assignmentsubmissions.Filter"}},
					&components.TableButtonCreate{Link: assignmentSubmissionCreateURLGetter()},
				},
				OnClick: getters.NavigateGetter(lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Assignment",
						Name:  "AssignmentTitle",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.AssignmentTitle")},
						},
					},
					{
						Label: "Course",
						Name:  "Course.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Course.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "SubmissionStatus",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.SubmissionStatus")},
						},
					},
				},
			},
		},
	})
}

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

	lago.RegistryPage.Register("assignmentsubmissions.DeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this submission?",
				CancelUrl: lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}
