package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"gorm.io/gorm"
)

func init() {
	registerMenuPages()
	registerStudentsMenuAssignmentSubmissionsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerStudentsMenuAssignmentSubmissionsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Assignment Submissions"),
			Url:   lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignmentsubmissions.DetailMenu", &components.SidebarMenu{
		Title: getters.Format("Submission: %s", getters.Any(getters.Key[string]("assignmentsubmission.AssignmentTitle"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to assignment submissions"),
			Url:   lago.RoutePath("assignmentsubmissions.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Submission detail"),
				Url: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit submission"),
				Url: lago.RoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete submission"),
				Url: lago.RoutePath("assignmentsubmissions.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("assignmentsubmissions.Filter", &components.FormComponent[AssignmentSubmission]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("assignmentsubmissions.DefaultRoute", nil)),
		Method:   http.MethodGet,
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

// assignmentSubmissionFormAcademicRecordGetter loads the academic record for the FK input with
// preloads so Display (Student.StudentNo) works; plain getters.Association does not preload Student.
func assignmentSubmissionFormAcademicRecordGetter() getters.Getter[p_nirmancampus_academicrecords.AcademicRecord] {
	return func(ctx context.Context) (p_nirmancampus_academicrecords.AcademicRecord, error) {
		var zero p_nirmancampus_academicrecords.AcademicRecord
		id, err := getters.Key[uint]("$in.AcademicRecordID")(ctx)
		if err != nil || id == 0 {
			return zero, nil
		}
		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return zero, nil
		}
		var rec p_nirmancampus_academicrecords.AcademicRecord
		err = db.Preload("Student").Preload("Program").First(&rec, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return zero, nil
			}
			slog.Error("assignmentSubmissionFormAcademicRecordGetter: load failed", "error", err, "id", id)
			return zero, err
		}
		return rec, nil
	}
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
							&components.InputNumber[int]{
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
							&components.InputNumber[int]{
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
								Url:         lago.RoutePath("courses.SelectRoute", nil),
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
								Url:         lago.RoutePath("academicrecords.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Student.StudentNo"),
								Placeholder: "Select an academic record...",
								Getter:      assignmentSubmissionFormAcademicRecordGetter(),
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

func registerFormPages() {
	lago.RegistryPage.Register("assignmentsubmissions.FormFields", assignmentSubmissionFormFields())

	lago.RegistryPage.Register("assignmentsubmissions.CreateForm", &components.Modal{
		Page: components.Page{
			Key:   "assignmentsubmissions.CreateModal",
			Roles: []string{"admin", "superuser"},
		},
		UID: "assignmentsubmissions-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[AssignmentSubmission]{
				OnSubmit: getters.FormSubmitCloseModal(lago.RoutePath("assignmentsubmissions.CreateRoute", nil)),
				Method:   http.MethodPost,
				Title:    "Create submission",
				Subtitle: "Create a new assignment submission",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentSubmissionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save submission", Classes: "btn-primary"},
						},
					},
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
				OnSubmit: getters.FormSubmit(lago.RoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				})),
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
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AssignmentSubmission]{
				Page:    components.Page{Key: "assignmentsubmissions.TableBody"},
				UID:     "assignment-submissions-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AssignmentSubmission]]("assignmentsubmissions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "assignmentsubmissions.Filter"}},
					&components.ButtonModal{
						Page:    components.Page{Roles: []string{"admin", "superuser"}},
						Url:     lago.RoutePath("assignmentsubmissions.CreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
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
				CancelUrl: lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
			},
		},
	})
}
