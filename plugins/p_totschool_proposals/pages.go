package p_totschool_proposals

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
)

func init() {
	registerMenus()
	registerFilter()
	registerForms()
	registerTable()
	registerDetail()
	registerModal()
	registerDelete()
}

func registerMenus() {
	lago.RegistryPage.Register("proposals.ProposalMenu", components.SidebarMenu{
		Title: getters.Static("Proposals"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.Static("All Proposals"), Url: lago.RoutePath("proposals.ListRoute", nil)},
			components.SidebarMenuItem{Title: getters.Static("Create Proposal"), Url: lago.RoutePath("proposals.CreateRoute", nil)},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalDetailMenu", components.SidebarMenu{
		Title: getters.Format("Proposal: %s", getters.Any(getters.Key[string]("proposal.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Proposals"),
			Url:   lago.RoutePath("proposals.ListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.Static("Proposal Detail"), Url: lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("proposal.ID"))})},
			components.SidebarMenuItem{Title: getters.Static("Edit Proposal"), Url: lago.RoutePath("proposals.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("proposal.ID"))})},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("proposals.ProposalFilter", components.FormComponent[Proposal]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(lago.RoutePath("proposals.ListRoute", nil))),

		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Title", Name: "Title", Getter: getters.Key[string]("$get.Title")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func registerForms() {
	lago.RegistryPage.Register("proposals.ProposalFormFields", components.ContainerColumn{
		Children: []components.PageInterface{
			components.ContainerColumn{Children: []components.PageInterface{components.InputText{Label: "Proposal Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")}, components.InputKeyValue{Getter: getters.Key[datatypes.JSON]("$in.Answers"), Keys: getters.Static(QUESTIONS), Name: "Answers"}}},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("proposals.CreateRoute", nil))),

				Title:          "Create Proposal",
				Subtitle:       "Fill in the questionnaire answers",
				ChildrenInput:  []components.PageInterface{components.InputText{Label: "Proposal Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")}, components.InputKeyValue{Getter: getters.Key[datatypes.JSON]("$in.Answers"), Keys: getters.Static(QUESTIONS), Name: "Answers"}},
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Proposal"}},
			},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Getter: getters.Key[Proposal]("proposal"),
				Attr:   getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("proposals.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}))),

				Title:    "Edit Proposal",
				Subtitle: "Update questionnaire answers",
				ChildrenInput: []components.PageInterface{
					components.InputText{Label: "Title", Name: "Title", Getter: getters.Key[string]("$in.Title")},
					components.InputKeyValue{
						Getter: getters.Key[datatypes.JSON]("$in.Answers"),
						Keys:   getters.Static(QUESTIONS),
						Name:   "Answers",
					},
				},
				ChildrenAction: []components.PageInterface{
					components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							components.ButtonModal{
								Label:   "Delete",
								Icon:    "trash",
								Url:     lago.RoutePath("proposals.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Classes: "btn-outline btn-error btn-sm",
							},
							components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									components.ButtonSubmit{Label: "Save Proposal"},
								},
							},
						},
					},
				},
			},
		},
	})
}

func registerTable() {
	lago.RegistryPage.Register("proposals.ProposalTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.DataTable[Proposal]{
				UID:      "proposal-table",
				Data:     getters.Key[components.ObjectList[Proposal]]("proposals"),
				Title:    "Proposals",
				Subtitle: "List of financial proposals",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "proposals.ProposalFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("proposals.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Created At", Name: "CreatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")}}},
					{Label: "Updated At", Name: "UpdatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")}}},
				},
			},
		},
	})
}

func registerDetail() {
	generatedSection := []components.PageInterface{
		components.Accordion{
			Classes: "mt-4",
			Items: []components.AccordionItem{
				{
					Title: components.FieldText{Classes: "font-semibold", Getter: getters.Static("Generated Proposal")},
					Children: []components.PageInterface{
						components.ContainerColumn{Classes: "my-2", Children: []components.PageInterface{
							components.ContainerRow{Classes: "flex flex-wrap justify-between items-center gap-4 mb-4", Children: []components.PageInterface{
								components.ContainerColumn{Classes: "flex flex-wrap gap-2 items-center", Children: []components.PageInterface{
									components.ButtonDownload{Label: "Export to PDF", Link: lago.RoutePath("proposals.ExportPdfRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
									components.ButtonDownload{Label: "Export to Word", Link: lago.RoutePath("proposals.ExportDocxRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
									components.ButtonModal{Label: "Edit with AI", Url: lago.RoutePath("proposals.AiEditFormRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
									components.ButtonPost{Label: "Regenerate Proposal", URL: lago.RoutePath("proposals.GenerateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-outline btn-primary btn-sm"},
								}},
							}},
							components.FieldMarkdown{Classes: "ml-2", Getter: getters.Key[string]("$in.GeneratedContent")},
						}},
					},
				},
			},
		},
	}

	pendingSection := []components.PageInterface{
		components.HTMXPolling{
			URL: lago.RoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$in.ID")),
			}),
			Children: []components.PageInterface{
				components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
					components.FieldText{Getter: getters.Static("Generating...")},
					components.ButtonPost{
						Label:   "Cancel Generation",
						URL:     lago.RoutePath("proposals.CancelRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
						Classes: "btn-outline btn-error btn-sm",
					},
				}},
			},
		},
	}

	idleSection := []components.PageInterface{
		components.ButtonPost{
			Page: components.Page{
				Key: "proposals.GenerateProposalWithAi",
			},
			Label: "Generate Proposal with AI",
			URL:   lago.RoutePath("proposals.GenerateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}), Classes: "btn-primary",
		},
	}

	lago.RegistryPage.Register("proposals.ProposalDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail[Proposal]{
				Getter: getters.Key[Proposal]("proposal"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.Key[time.Time]("$in.CreatedAt")}}},
						components.Accordion{Classes: "mt-4", Items: []components.AccordionItem{
							{
								Title: components.FieldText{Classes: "font-semibold", Getter: getters.Static("Questionnaire Answers")},
								Children: []components.PageInterface{
									components.FieldKeyValue{Getter: getters.Key[datatypes.JSON]("$in.Answers")},
								},
							},
						}},
						components.ContainerColumn{Children: []components.PageInterface{
							components.ShowIf{Getter: getters.Any(getterGenerated()), Children: generatedSection},
							components.ShowIf{Getter: getters.Any(getterGenerationPending()), Children: pendingSection},
							components.ShowIf{Getter: getters.Any(getterIdleGeneration()), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterGenerated() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content != "" {
			return true, nil
		}
		return false, nil
	}
}

func getterGenerationPending() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id != nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}

func getterIdleGeneration() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}

func registerModal() {
	lago.RegistryPage.Register("proposals.AiEditModal", components.Modal{
		UID: "ai-edit-modal",
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Getter: getters.Key[Proposal]("proposal"),
				Attr:   getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("proposals.AiEditRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("proposal.ID"))}))),

				Title: "Edit with AI",
				ChildrenInput: []components.PageInterface{
					components.InputTextarea{Name: "GeneratedContent", Label: "Current Proposal Markdown", Getter: getters.Key[string]("$in.GeneratedContent"), Rows: 8},
					components.InputTextarea{Name: "instructions", Label: "Instructions for AI", Getter: getters.Key[string]("$in.instructions"), Rows: 4, Required: true},
				},
				ChildrenAction: []components.PageInterface{
					components.ContainerRow{Classes: "flex justify-end gap-2", Children: []components.PageInterface{
						components.ButtonSubmit{Label: "Generate", Classes: "btn-secondary"},
					}},
				},
			},
		},
	})
}

func registerDelete() {
	lago.RegistryPage.Register("proposals.ProposalDeleteForm", components.Modal{
		UID: "proposal-delete-modal",
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this proposal?",
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("proposals.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("proposal.ID")),
				}))),
			},
		},
	})
}
