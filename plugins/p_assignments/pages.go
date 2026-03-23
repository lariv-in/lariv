package p_assignments

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
	"github.com/lariv-in/lago/p_semesters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("assignments.AssignmentMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Assignments"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Assignments"),
				Url:   lago.GetterRoutePath("assignments.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("assignments.AssignmentDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Assignment: %s", getters.GetterAny(getters.GetterKey[string]("assignment.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Assignments"),
			Url:   lago.GetterRoutePath("assignments.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Assignment Detail"),
				Url: lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignment.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Assignment"),
				Url: lago.GetterRoutePath("assignments.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignment.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Assignment"),
				Url: lago.GetterRoutePath("assignments.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignment.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("assignments.AssignmentFilter", &components.FormComponent[Assignment]{
		Url:    lago.GetterRoutePath("assignments.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Description",
				Name:   "Description",
				Getter: getters.GetterKey[string]("$get.Description"),
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

	lago.RegistryPage.Register("assignments.AssignmentSelectionFilter", &components.FormComponent[Assignment]{
		Url:    lago.GetterRoutePath("assignments.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
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

func assignmentFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assignments.AssignmentFormFieldsBody"},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Name",
								Name:     "Name",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Due"),
						Children: []components.PageInterface{
							&components.InputDatetime{
								Label:    "Due",
								Name:     "Due",
								Required: true,
								Getter:   getters.GetterKey[time.Time]("$in.Due"),
							},
						},
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.MaxMarks"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Max Marks",
								Name:     "MaxMarks",
								Required: true,
								Getter:   getters.GetterKey[int]("$in.MaxMarks"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.SemesterID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_semesters.Semester]{
						Label:       "Semester",
						Name:        "SemesterID",
						Required:    true,
						Getter:      getters.GetterAssociation[p_semesters.Semester](getters.GetterKey[uint]("$in.SemesterID")),
						Url:         lago.GetterRoutePath("semesters.SelectRoute", nil),
						Display:     getters.GetterKey[string]("$in.Name"),
						Placeholder: "Select a semester...",
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Description",
						Name:   "Description",
						Rows:   4,
						Getter: getters.GetterKey[string]("$in.Description"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Assets"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_filesystem.VNode]{
						Label:       "Assets",
						Name:        "Assets",
						Getter:      getters.GetterKey[[]p_filesystem.VNode]("$in.Assets"),
						Url:         lago.GetterRoutePath("filesystem.MultiSelectRoute", nil),
						Display:     getters.GetterKey[string]("$in.Name"),
						Placeholder: "Select assets...",
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("assignments.AssignmentFormFields", assignmentFormFields())

	lago.RegistryPage.Register("assignments.AssignmentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Assignment]{
				Url:      lago.GetterRoutePath("assignments.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Assignment",
				Subtitle: "Create a new assignment",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Assignment"},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignments.AssignmentUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Assignment]{
				Getter: getters.GetterKey[Assignment]("assignment"),
				Url: lago.GetterRoutePath("assignments.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Assignment",
				Subtitle: "Update assignment details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					assignmentFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Assignment"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("assignments.AssignmentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentMenu"},
		},
		Children: []components.PageInterface{
			&components.Environment[uint]{
				Label:   "Semester",
				Key:     getters.GetterStatic("semester"),
				Options: semestersEnvOptionsGetterForEnvironment,
				Default: semesterEnvironmentDefaultGetter,
			},
			&components.DataTable[Assignment]{
				Page:      components.Page{Key: "assignments.AssignmentTableBody"},
				UID:       "assignment-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Assignment]]("assignments"),
				CreateUrl: lago.GetterRoutePath("assignments.CreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "assignments.AssignmentFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Semester",
						Name:  "Semester",
						Children: []components.PageInterface{
							&components.FieldText{Getter: semesterNameFromRow()},
						},
					},
					{
						Label: "Due",
						Name:  "Due",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Due")},
						},
					},
					{
						Label: "Max Marks",
						Name:  "MaxMarks",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.MaxMarks"))),
							},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("assignments.AssignmentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Assignment]{
				Getter: getters.GetterKey[Assignment]("assignment"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "assignments.AssignmentDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Semester.Name")},
							&components.LabelInline{
								Title: "Due",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.Due")},
								},
							},
							&components.LabelInline{
								Title:   "Description",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Description")},
								},
							},
							&components.LabelInline{
								Title: "Max Marks",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$in.MaxMarks"))),
									},
								},
							},
							&components.LabelInline{
								Title: "Assets",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_filesystem.VNode]{
										Getter:  getters.GetterKey[[]p_filesystem.VNode]("$in.Assets"),
										Display: getters.GetterKey[string]("$in.Name"),
										Link: lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
										}),
										Classes: "w-full",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignments.AssignmentDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignments.AssignmentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this assignment?",
				CancelUrl: lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("assignment.ID")),
				}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("assignments.AssignmentSelectionTable", &components.Modal{
		UID:   "assignment-selection-modal",
		Title: "Select Assignment",
		Children: []components.PageInterface{
			&components.DataTable[Assignment]{
				Page:            components.Page{Key: "assignments.AssignmentSelectionTableBody"},
				UID:             "assignment-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Assignment]]("assignments"),
				OnClick:         getters.GetterSelect("AssignmentID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "assignments.AssignmentSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Semester",
						Name:  "Semester",
						Children: []components.PageInterface{
							&components.FieldText{Getter: semesterNameFromRow()},
						},
					},
					{
						Label: "Due",
						Name:  "Due",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Due")},
						},
					},
					{
						Label: "Max Marks",
						Name:  "MaxMarks",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.MaxMarks"))),
							},
						},
					},
				},
			},
		},
	})
}

func semesterNameFromRow() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		name, err := getters.GetterKey[string]("$row.Semester.Name")(ctx)
		if err != nil {
			return "", nil
		}
		return name, nil
	}
}

// semesterEnvironmentDefaultGetter selects the semester whose [Start, End] contains time.Now(),
// matching assignmentsListSemesterEnvQueryPatcher when the environment cookie has no semester.
func semesterEnvironmentDefaultGetter(ctx context.Context) (uint, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return 0, nil
	}
	id, ok := semesterEnvironmentDefault(db, time.Now())
	if !ok {
		return 0, nil
	}
	return id, nil
}

func semestersEnvOptionsGetterForEnvironment(ctx context.Context) ([]registry.Pair[uint, string], error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("semestersEnvOptionsGetterForEnvironment: missing $db in context")
	}

	var semesters []p_semesters.Semester
	if err := db.Order(`"start" ASC`).Find(&semesters).Error; err != nil {
		return nil, err
	}

	options := make([]registry.Pair[uint, string], 0, len(semesters))
	for _, s := range semesters {
		options = append(options, registry.Pair[uint, string]{
			Key:   s.ID,
			Value: s.Name,
		})
	}
	return options, nil
}
