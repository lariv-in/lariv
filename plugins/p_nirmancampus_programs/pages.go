package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func universityChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "IGNOU", Value: "IGNOU"},
		{Key: "MRSPTU", Value: "MRSPTU"},
	}
}

func universityFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$get.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programUniversityPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$in.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func universityFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "University",
		Name:    "University",
		Choices: getters.GetterStatic(universityChoices()),
		Getter:  universityFilterPairGetter(),
	}
}

func universityFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "University",
		Name:     "University",
		Required: false,
		Choices:  getters.GetterStatic(universityChoices()),
		Getter:   programUniversityPairGetter(),
	}
}

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("programs.ProgramMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Programs"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Programs"),
				Url:   lago.GetterRoutePath("programs.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Program: %s", getters.GetterAny(getters.GetterKey[string]("program.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Programs"),
			Url:   lago.GetterRoutePath("programs.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Program Detail"),
				Url: lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit Program"),
				Url: lago.GetterRoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Delete Program"),
				Url: lago.GetterRoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("programs.ProgramFilter", &components.FormComponent[Program]{
		Url:    lago.GetterRoutePath("programs.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
			},
			universityFilterSelect(),
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

	lago.RegistryPage.Register("programs.ProgramSelectionFilter", &components.FormComponent[Program]{
		Url:    lago.GetterRoutePath("programs.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
			},
			universityFilterSelect(),
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

func programFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "programs.ProgramFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
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
						Error: getters.GetterKey[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:  "Code",
								Name:   "Code",
								Getter: getters.GetterKey[string]("$in.Code"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Description"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Description",
								Name:   "Description",
								Rows:   3,
								Getter: getters.GetterKey[string]("$in.Description"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.University"),
						Children: []components.PageInterface{
							universityFormSelect(),
						},
					},
				},
			},
		},
	}
}

func programCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.GetterRoutePath("programs.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("programs.ProgramFormFields", programFormFields())

	lago.RegistryPage.Register("programs.ProgramCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Program]{
				Url:      lago.GetterRoutePath("programs.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Program",
				Subtitle: "Create a new program",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					programFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Program"},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Program]{
				Getter: getters.GetterKey[Program]("program"),
				Url: lago.GetterRoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Program",
				Subtitle: "Update program details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					programFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Program"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("programs.ProgramTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:      components.Page{Key: "programs.ProgramTableBody"},
				UID:       "program-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Program]]("programs"),
				CreateUrl: programCreateUrlGetter(),
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				FilterComponent: lago.DynamicPage{Name: "programs.ProgramFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.University")},
						},
					},
					{
						Label: "Description",
						Name:  "Description",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Description")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("programs.ProgramDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Program]{
				Getter: getters.GetterKey[Program]("program"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "programs.ProgramDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Code")},
							&components.LabelInline{
								Title:   "University",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.University")},
								},
							},
							&components.LabelInline{
								Title:   "Description",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Description")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this program?",
				CancelUrl: lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("programs.ProgramSelectionTable", &components.Modal{
		UID:   "program-selection-modal",
		Title: "Select Program",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:            components.Page{Key: "programs.ProgramSelectionTableBody"},
				UID:             "program-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Program]]("programs"),
				OnClick:         getters.GetterSelect("ProgramID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "programs.ProgramSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.University")},
						},
					},
				},
			},
		},
	})
}
