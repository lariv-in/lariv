package p_llm_assistant

import (
	"context"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
)

func skillFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Description", Name: "Description", Getter: getters.Key[string]("$in.Description")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Content"),
				Children: []components.PageInterface{
					&components.InputTextarea{Label: "Content", Name: "Content", Required: true, Getter: getters.Key[string]("$in.Content")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Files"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_filesystem.VNode]{
						Label:       "Files",
						Name:        "Files",
						Url:         lariv.RoutePath("filesystem.MultiSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select files...",
						Required:    false,
						Getter:      getters.Key[[]p_filesystem.VNode]("$in.Files"),
					},
				},
			},
		},
	}
}

func registerSkillsPages() {
	// Sidebar menu for a specific Skill Detail
	registerPluginPage("llm_assistant.SkillsDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Skill: %s", getters.Any(getters.Key[string]("skill.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Skills"),
			Url:   lariv.RoutePath("llm_assistant.SkillsListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Skill Details"),
				Url: lariv.RoutePath("llm_assistant.SkillsDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("skill.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Skill"),
				Url: lariv.RoutePath("llm_assistant.SkillsUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("skill.ID")),
				}),
			},
		},
	})

	// List Page
	registerPluginPage("llm_assistant.SkillsListPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lariv.DynamicPage{Name: "llm_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Skill]{
				UID:     "skills-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Skill]]("skills"),
				Actions: []components.PageInterface{
					&components.ButtonLink{
						Link:    lariv.RoutePath("llm_assistant.SkillsCreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
					&components.ButtonModalForm{
						Icon:        "arrow-up-on-square",
						Name:        getters.Static("llm_assistant.SkillsImportPage"),
						Url:         lariv.RoutePath("llm_assistant.SkillsImportRoute", nil),
						FormPostURL: lariv.RoutePath("llm_assistant.SkillsImportRoute", nil),
						ModalUID:    "skill-import-modal",
						Classes:     "btn-square btn-outline btn-sm",
					},
				},
				RowAttr: getters.RowAttrNavigate(lariv.RoutePath("llm_assistant.SkillsDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Description", Name: "Description", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Description")},
					}},
				},
			},
		},
	})

	// Detail Page
	registerPluginPage("llm_assistant.SkillsDetailPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lariv.DynamicPage{Name: "llm_assistant.SkillsDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Skill]{
				Page:   components.Page{Key: "llm_assistant.SkillsDetailContent"},
				Getter: getters.Key[Skill]("skill"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Description")},
							&components.ButtonDownload{
								Label: "Export Skill",
								Icon:  "arrow-down-tray",
								Link: lariv.RoutePath("llm_assistant.SkillsExportRoute", map[string]getters.Getter[any]{
									"id": getters.Any(getters.Key[uint]("$in.ID")),
								}),
								Classes: "btn-outline btn-sm w-fit mt-2",
							},
							&components.LabelInline{
								Title:   "Content",
								Classes: "mt-4 block",
								Children: []components.PageInterface{
									&components.FieldMarkdown{Getter: getters.Key[string]("$in.Content")},
								},
							},
							&components.LabelInline{
								Title:   "Files",
								Classes: "mt-4 block",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{VNode: getters.Key[[]p_filesystem.VNode]("$in.Files")},
								},
							},
						},
					},
				},
			},
		},
	})

	// Create Page
	registerPluginPage("llm_assistant.SkillsCreatePage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lariv.DynamicPage{Name: "llm_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("llm_assistant.SkillsCreatePage"),
				ActionURL: lariv.RoutePath("llm_assistant.SkillsCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Skill]{
						Getter:   func(ctx context.Context) (Skill, error) { return Skill{}, nil },
						Attr:     getters.FormBubbling(getters.Static("llm_assistant.SkillsCreatePage")),
						Title:    "Create Skill",
						Subtitle: "Define a new assistant skill",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							skillFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex justify-end gap-2 mt-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Skill", Classes: "btn-primary"},
								},
							},
						},
					},
				},
			},
		},
	})

	// Update Page
	registerPluginPage("llm_assistant.SkillsUpdatePage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lariv.DynamicPage{Name: "llm_assistant.SkillsDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("llm_assistant.SkillsUpdatePage"),
				ActionURL: lariv.RoutePath("llm_assistant.SkillsUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("skill.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Skill]{
						Getter:   getters.Key[Skill]("skill"),
						Attr:     getters.FormBubbling(getters.Static("llm_assistant.SkillsUpdatePage")),
						Title:    "Edit Skill",
						Subtitle: "Update skill definition",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							skillFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Skill"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("llm_assistant.SkillsDeletePage"),
												Url:         lariv.RoutePath("llm_assistant.SkillsDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("skill.ID"))}),
												FormPostURL: lariv.RoutePath("llm_assistant.SkillsDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("skill.ID"))}),
												ModalUID:    "skill-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	// Delete Confirmation Page
	registerPluginPage("llm_assistant.SkillsDeletePage", &components.Modal{
		UID: "skill-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this skill?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})

	// Import Page Modal
	registerPluginPage("llm_assistant.SkillsImportPage", &components.Modal{
		UID: "skill-import-modal",
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("llm_assistant.SkillsImportPage"),
				ActionURL: lariv.RoutePath("llm_assistant.SkillsImportRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[any]{
						Attr:     getters.FormBubbling(getters.Static("llm_assistant.SkillsImportPage")),
						Title:    "Import Skill",
						Subtitle: "Upload a skill zip file to import it",
						ChildrenInput: []components.PageInterface{
							&components.InputFile{
								Label:    "Skill Zip File",
								Name:     "File",
								Required: true,
								Accept:   ".zip",
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex justify-end gap-2 mt-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Import", Classes: "btn-primary"},
								},
							},
						},
					},
				},
			},
		},
	})
}

func init() {
	registerSkillsPages()
}
