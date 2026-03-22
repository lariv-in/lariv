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
		Title: getters.GetterStatic("Proposals"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("All Proposals"), Url: lago.GetterRoutePath("proposals.ListRoute", nil)},
			components.SidebarMenuItem{Title: getters.GetterStatic("Create Proposal"), Url: lago.GetterRoutePath("proposals.CreateRoute", nil)},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Proposal: %s", getters.GetterAny(getters.GetterKey[string]("proposal.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Proposals"),
			Url:   lago.GetterRoutePath("proposals.ListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("Proposal Detail"), Url: lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("proposal.ID"))})},
			components.SidebarMenuItem{Title: getters.GetterStatic("Edit Proposal"), Url: lago.GetterRoutePath("proposals.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("proposal.ID"))})},
			components.SidebarMenuItem{Title: getters.GetterStatic("Delete Proposal"), Url: lago.GetterRoutePath("proposals.DeleteRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("proposal.ID"))})},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("proposals.ProposalFilter", components.FormComponent[Proposal]{
		Url:    lago.GetterRoutePath("proposals.ListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Title", Name: "Title", Getter: getters.GetterKey[string]("$get.Title")},
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
			components.ContainerColumn{Children: []components.PageInterface{components.InputText{Label: "Proposal Title", Name: "Title", Required: true, Getter: getters.GetterKey[string]("$in.Title")}, components.InputKeyValue{Getter: getters.GetterKey[datatypes.JSON]("$in.Answers"), Keys: getters.GetterStatic(QUESTIONS), Name: "Answers"}}},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Url:            lago.GetterRoutePath("proposals.CreateRoute", nil),
				Method:         http.MethodPost,
				Title:          "Create Proposal",
				Subtitle:       "Fill in the questionnaire answers",
				ChildrenInput:  []components.PageInterface{components.InputText{Label: "Proposal Title", Name: "Title", Required: true, Getter: getters.GetterKey[string]("$in.Title")}, components.InputKeyValue{Getter: getters.GetterKey[datatypes.JSON]("$in.Answers"), Keys: getters.GetterStatic(QUESTIONS), Name: "Answers"}},
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Proposal"}},
			},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Getter:   getters.GetterKey[Proposal]("proposal"),
				Url:      lago.GetterRoutePath("proposals.UpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Edit Proposal",
				Subtitle: "Update questionnaire answers",
				ChildrenInput: []components.PageInterface{
					components.InputText{Label: "Title", Name: "Title", Getter: getters.GetterKey[string]("$in.Title")},
					components.InputKeyValue{
						Getter: getters.GetterKey[datatypes.JSON]("$in.Answers"),
						Keys:   getters.GetterStatic(QUESTIONS),
						Name:   "Answers",
					},
				},
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Proposal"}},
			},
		},
	})
}

func registerTable() {
	lago.RegistryPage.Register("proposals.ProposalTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.DataTable[Proposal]{
				UID:             "proposal-table",
				Data:            getters.GetterKey[components.ObjectList[Proposal]]("proposals"),
				Title:           "Proposals",
				Subtitle:        "List of financial proposals",
				CreateUrl:       lago.GetterRoutePath("proposals.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "proposals.ProposalFilter"},
				Columns: []components.TableColumn{
					{Label: "Title", Key: "Title", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey[string]("$row.Title")}}},
					{Label: "Created At", Key: "CreatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.CreatedAt")}}},
					{Label: "Updated At", Key: "UpdatedAt", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.UpdatedAt")}}},
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
					Title: components.FieldText{Classes: "font-semibold", Getter: getters.GetterStatic("Generated Proposal")},
					Children: []components.PageInterface{
						components.ContainerColumn{Classes: "my-2", Children: []components.PageInterface{
							components.ContainerRow{Classes: "flex flex-wrap justify-between items-center gap-4 mb-4", Children: []components.PageInterface{
								components.ContainerColumn{Classes: "flex gap-2", Children: []components.PageInterface{
									components.ButtonDownload{Label: "Export to PDF", Link: lago.GetterRoutePath("proposals.ExportPdfRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
									components.ButtonModal{Label: "Edit with AI", Url: lago.GetterRoutePath("proposals.AiEditFormRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-secondary btn-sm"},
									components.ButtonPost{Label: "Regenerate Proposal", URL: lago.GetterRoutePath("proposals.GenerateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-outline btn-primary btn-sm"},
								}},
							}},
							components.FieldMarkdown{Classes: "ml-2", Getter: getters.GetterKey[string]("$in.GeneratedContent")},
						}},
					},
				},
			},
		},
	}

	pendingSection := []components.PageInterface{
		components.HTMXPolling{
			URL: lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
			}),
			Children: []components.PageInterface{
				components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
					components.FieldText{Getter: getters.GetterStatic("Generating...")},
					components.ButtonPost{
						Label:   "Cancel Generation",
						URL:     lago.GetterRoutePath("proposals.CancelRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
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
			URL:   lago.GetterRoutePath("proposals.GenerateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}), Classes: "btn-primary",
		},
	}

	lago.RegistryPage.Register("proposals.ProposalDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail[Proposal]{
				Getter: getters.GetterKey[Proposal]("proposal"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.GetterKey[string]("$in.Title")},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.CreatedAt")}}},
						components.Accordion{Classes: "mt-4", Items: []components.AccordionItem{
							{
								Title: components.FieldText{Classes: "font-semibold", Getter: getters.GetterStatic("Questionnaire Answers")},
								Children: []components.PageInterface{
									components.FieldKeyValue{Getter: getters.GetterKey[datatypes.JSON]("$in.Answers")},
								},
							},
						}},
						components.ContainerColumn{Children: []components.PageInterface{
							components.ShowIf{Getter: getters.GetterAny(getterGenerated()), Children: generatedSection},
							components.ShowIf{Getter: getters.GetterAny(getterGenerationPending()), Children: pendingSection},
							components.ShowIf{Getter: getters.GetterAny(getterIdleGeneration()), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterGenerated() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.GetterKey[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.GetterKey[string]("$in.GeneratedContent")(ctx)
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
		id, err := getters.GetterKey[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.GetterKey[string]("$in.GeneratedContent")(ctx)
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
		id, err := getters.GetterKey[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.GetterKey[string]("$in.GeneratedContent")(ctx)
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
		UID:   "ai-edit-modal",
		Title: "Edit with AI",
		Children: []components.PageInterface{
			components.FormComponent[Proposal]{
				Getter: getters.GetterKey[Proposal]("proposal"),
				Url:    lago.GetterRoutePath("proposals.AiEditRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("proposal.ID"))}),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputTextarea{Name: "GeneratedContent", Label: "Current Proposal Markdown", Getter: getters.GetterKey[string]("$in.GeneratedContent"), Rows: 8},
					components.InputTextarea{Name: "instructions", Label: "Instructions for AI", Getter: getters.GetterKey[string]("$in.instructions"), Rows: 4, Required: true},
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
	lago.RegistryPage.Register("proposals.ProposalDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this proposal?",
				CancelUrl: lago.GetterRoutePath("proposals.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("proposal.ID"))}),
			},
		},
	})
}
