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
		Title: getters.GetterStatic("Assignment Submissions"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Submissions"),
				Url:   lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("assignmentsubmissions.DetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Submission: %s", getters.GetterAny(getters.GetterKey[string]("assignmentsubmission.AssignmentTitle"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all submissions"),
			Url:   lago.GetterRoutePath("assignmentsubmissions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Submission detail"),
				Url: lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit submission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Delete submission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentsubmission.ID")),
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
				Getter: getters.GetterKey[string]("$get.AssignmentTitle"),
			},
			&components.InputText{
				Label:  "Submission status",
				Name:   "SubmissionStatus",
				Getter: getters.GetterKey[string]("$get.SubmissionStatus"),
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
						Error: getters.GetterKey[error]("$error.AssignmentTitle"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Assignment title",
								Name:     "AssignmentTitle",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.AssignmentTitle"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.SubmissionStatus"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Submission status",
								Name:     "SubmissionStatus",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.SubmissionStatus"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.MaxMarks"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Max marks",
								Name:     "MaxMarks",
								Required: true,
								Getter:   getters.GetterKey[int]("$in.MaxMarks"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Marks"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Marks",
								Name:     "Marks",
								Required: true,
								Getter:   getters.GetterKey[int]("$in.Marks"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.CourseID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_courses.Course]{
								Label:       "Course",
								Name:        "CourseID",
								Required:    true,
								Url:         lago.GetterRoutePath("courses.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select a course...",
								Getter: getters.GetterAssociation[p_nirmancampus_courses.Course](
									getters.GetterKey[uint]("$in.CourseID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.AcademicRecordID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_academicrecords.AcademicRecord]{
								Label:       "Academic record",
								Name:        "AcademicRecordID",
								Required:    true,
								Url:         lago.GetterRoutePath("academicrecords.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Student.StudentNo"),
								Placeholder: "Select an academic record...",
								Getter: getters.GetterAssociation[p_nirmancampus_academicrecords.AcademicRecord](
									getters.GetterKey[uint]("$in.AcademicRecordID"),
								),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Assets"),
				Children: []components.PageInterface{
					&p_filesystem.InputMultiVNode{
						Label:            "Assets",
						Name:             "Assets",
						VNode:            getters.GetterKey[[]p_filesystem.VNode]("$in.Assets"),
						AllowedFiletypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".webp", ".doc", ".docx"},
					},
				},
			},
		},
	}
}

func assignmentSubmissionCreateURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
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
				Getter: getters.GetterKey[AssignmentSubmission]("assignmentsubmission"),
				Url: lago.GetterRoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
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
				Data:            getters.GetterKey[components.ObjectList[AssignmentSubmission]]("assignmentsubmissions"),
				CreateUrl:       assignmentSubmissionCreateURLGetter(),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "assignmentsubmissions.Filter"},
				DefaultView:     "Grid",
				Columns: []components.TableColumn{
					{
						Label: "Assignment",
						Name:  "AssignmentTitle",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.AssignmentTitle")},
						},
					},
					{
						Label: "Course",
						Name:  "Course.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "SubmissionStatus",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.SubmissionStatus")},
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
				Getter: getters.GetterKey[AssignmentSubmission]("assignmentsubmission"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "assignmentsubmissions.DetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.AssignmentTitle")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Course.Name")},
							&components.LabelInline{
								Title: "Submission status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.SubmissionStatus")},
								},
							},
							&components.LabelInline{
								Page:  components.Page{Roles: []string{"admin", "superuser"}},
								Title: "Marks",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterFormat("%d / %d", getters.GetterAny(getters.GetterKey[int]("$in.Marks")), getters.GetterAny(getters.GetterKey[int]("$in.MaxMarks")))},
								},
							},
							&components.LabelInline{
								Title: "Academic record",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterFormat("%s · %s", getters.GetterAny(getters.GetterKey[string]("$in.AcademicRecord.Student.StudentNo")), getters.GetterAny(getters.GetterKey[string]("$in.AcademicRecord.Program.Name")))},
								},
							},
							&components.LabelInline{
								Title: "Assets",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{VNode: getters.GetterKey[[]p_filesystem.VNode]("$in.Assets")},
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
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}
