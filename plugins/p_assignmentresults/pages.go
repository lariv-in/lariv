package p_assignmentresults

import (
	"log"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_academicrecords"
	"github.com/lariv-in/lago/plugins/p_assignments"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
	registerAssignmentPatches()
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat(
			"%s — %s",
			getters.GetterAny(getters.GetterKey[string]("assignmentresult.Assignment.Name")),
			getters.GetterAny(getters.GetterKey[string]("assignmentresult.AcademicRecord.Student.User.Name")),
		),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Assignment"),
			Url: lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("assignmentresult.AssignmentID")),
			}),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Result Detail"),
				Url: lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentresult.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Result"),
				Url: lago.GetterRoutePath("assignmentresults.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentresult.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Result"),
				Url: lago.GetterRoutePath("assignmentresults.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentresult.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultFilter", &components.FormComponent[AssignmentResult]{
		Url:    lago.GetterRoutePath("assignmentresults.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Remarks",
				Name:   "Remarks",
				Getter: getters.GetterKey[string]("$get.Remarks"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentresults.AssignmentResultSelectionFilter", &components.FormComponent[AssignmentResult]{
		Url:    lago.GetterRoutePath("assignmentresults.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Remarks",
				Name:   "Remarks",
				Getter: getters.GetterKey[string]("$get.Remarks"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func assignmentResultSemesterEnvironment() *components.Environment[uint] {
	return &components.Environment[uint]{
		Label:   "Semester",
		Key:     getters.GetterStatic("semester"),
		Options: p_academicrecords.SemesterEnvironmentOptions,
		Default: p_academicrecords.SemesterEnvironmentDefaultGetter,
		Classes: "w-full max-w-md",
	}
}

func assignmentResultFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assignmentresults.AssignmentResultFormFieldsBody"},
		Children: []components.PageInterface{
			assignmentResultSemesterEnvironment(),
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.AssignmentID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_assignments.Assignment]{
								Label:       "Assignment",
								Name:        "AssignmentID",
								Required:    true,
								Url:         lago.GetterRoutePath("assignments.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select an assignment...",
								Getter: getters.GetterAssociation[p_assignments.Assignment](
									getters.GetterKey[uint]("$in.AssignmentID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.AcademicRecordID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_academicrecords.AcademicRecord]{
								Label:       "Academic record",
								Name:        "AcademicRecordID",
								Required:    true,
								Url:         lago.GetterRoutePath("academicrecords.SelectRoute", nil),
								Placeholder: "Select an academic record...",
								Display: getters.GetterFormat(
									"%s — %s",
									getters.GetterAny(getters.GetterKey[string]("$in.Student.User.Name")),
									getters.GetterAny(getters.GetterKey[string]("$in.Semester.Name")),
								),
								Getter: getters.GetterAssociation[p_academicrecords.AcademicRecord](
									getters.GetterKey[uint]("$in.AcademicRecordID"),
								),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
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
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Remarks"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Remarks",
						Name:   "Remarks",
						Rows:   4,
						Getter: getters.GetterKey[string]("$in.Remarks"),
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultFormFields", assignmentResultFormFields())

	lago.RegistryPage.Register("assignmentresults.AssignmentResultCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AssignmentResult]{
				Url:      lago.GetterRoutePath("assignmentresults.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Record Result",
				Subtitle: "Enter marks for an assignment",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentResultFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Result"},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentresults.AssignmentResultUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentresults.AssignmentResultDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AssignmentResult]{
				Getter: getters.GetterKey[AssignmentResult]("assignmentresult"),
				Url: lago.GetterRoutePath("assignmentresults.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Result",
				Subtitle: "Update marks or remarks",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentResultFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Result"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AssignmentResult]{
				Page:      components.Page{Key: "assignmentresults.AssignmentResultTableBody"},
				UID:       "assignment-result-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[AssignmentResult]]("assignmentresults"),
				CreateUrl: lago.GetterRoutePath("assignmentresults.CreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "assignmentresults.AssignmentResultFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Assignment",
						Name:  "Assignment.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Assignment.Name")},
						},
					},
					{
						Label: "Academic record",
						Name:  "AcademicRecord",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat(
									"%s — %s",
									getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Student.User.Name")),
									getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Semester.Name")),
								),
							},
						},
					},
					{
						Label: "Marks",
						Name:  "Marks",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Marks"))),
							},
						},
					},
					{
						Label: "Remarks",
						Name:  "Remarks",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Remarks")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentresults.AssignmentResultDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AssignmentResult]{
				Getter: getters.GetterKey[AssignmentResult]("assignmentresult"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "assignmentresults.AssignmentResultDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.GetterKey[string]("$in.Assignment.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.GetterFormat(
									"%s — %s",
									getters.GetterAny(getters.GetterKey[string]("$in.AcademicRecord.Student.User.Name")),
									getters.GetterAny(getters.GetterKey[string]("$in.AcademicRecord.Semester.Name")),
								),
							},
							&components.LabelInline{
								Title:   "Student number",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.AcademicRecord.Student.StudentNo")},
								},
							},
							&components.LabelInline{
								Title: "Semester",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.AcademicRecord.Semester.Name")},
								},
							},
							&components.LabelInline{
								Title: "Marks",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$in.Marks"))),
									},
								},
							},
							&components.LabelInline{
								Title: "Remarks",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Remarks")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentresults.AssignmentResultDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentresults.AssignmentResultDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this result?",
				CancelUrl: lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignmentresult.ID")),
				}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("assignmentresults.AssignmentResultSelectionTable", &components.Modal{
		UID:   "assignment-result-selection-modal",
		Title: "Select Assignment Result",
		Children: []components.PageInterface{
			&components.DataTable[AssignmentResult]{
				Page: components.Page{Key: "assignmentresults.AssignmentResultSelectionTableBody"},
				UID:  "assignment-result-selection-table",
				Data: getters.GetterKey[components.ObjectList[AssignmentResult]]("assignmentresults"),
				OnClick: getters.GetterSelect("AssignmentResultID", getters.GetterKey[uint]("$row.ID"), getters.GetterFormat(
					"%s / %s — %s",
					getters.GetterAny(getters.GetterKey[string]("$row.Assignment.Name")),
					getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Student.User.Name")),
					getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Semester.Name")),
				)),
				FilterComponent: lago.DynamicPage{Name: "assignmentresults.AssignmentResultSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Assignment",
						Name:  "Assignment.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Assignment.Name")},
						},
					},
					{
						Label: "Academic record",
						Name:  "AcademicRecord",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat(
									"%s — %s",
									getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Student.User.Name")),
									getters.GetterAny(getters.GetterKey[string]("$row.AcademicRecord.Semester.Name")),
								),
							},
						},
					},
					{
						Label: "Marks",
						Name:  "Marks",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Marks"))),
							},
						},
					},
				},
			},
		},
	})
}

func assignmentDetailResultsSection() components.PageInterface {
	return components.ContainerColumn{
		Page: components.Page{Key: "assignmentresults.AssignmentDetailResultsSection"},
		Children: []components.PageInterface{
			&components.DataTable[AssignmentResult]{
				Page:    components.Page{Key: "assignmentresults.AssignmentDetailResultsTable"},
				Title:   "Results",
				UID:     "assignment-detail-results-table",
				Classes: "w-full mt-2",
				Data:    getters.GetterKey[components.ObjectList[AssignmentResult]]("assignmentresults"),
				CreateUrl: getters.GetterFormat(
					"%s?AssignmentID=%d",
					getters.GetterAny(lago.GetterRoutePath("assignmentresults.CreateRoute", nil)),
					getters.GetterAny(getters.GetterKey[uint]("assignment.ID")),
				),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				Columns: []components.TableColumn{
					{
						Label: "Academic record",
						Name:  "AcademicRecord",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterKey[string]("$row.AcademicRecord.Student.User.Name"),
							},
						},
					},
					{
						Label: "Marks",
						Name:  "Marks",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Marks"))),
							},
						},
					},
					{
						Label: "Remarks",
						Name:  "Remarks",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Remarks")},
						},
					},
				},
			},
		},
	}
}

func registerAssignmentPatches() {
	lago.RegistryPage.Patch("assignments.AssignmentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			log.Panic("assignments.AssignmentMenu is not *components.SidebarMenu")
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Page:  components.Page{Key: "assignmentresults.menu_all_results"},
			Title: getters.GetterStatic("All Results"),
			Url:   lago.GetterRoutePath("assignmentresults.DefaultRoute", nil),
		})
		return menu
	})

	lago.RegistryPage.Patch("assignments.AssignmentDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("assignments.AssignmentDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "assignments.AssignmentDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, assignmentDetailResultsSection())
			return column
		})
		return scaffold
	})
}
