package p_totschool_proposals

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
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
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("All Proposals"), Url: lago.RoutePathGetter("proposals.ListRoute")},
			components.SidebarMenuItem{Title: getters.GetterStatic("Create Proposal"), Url: lago.RoutePathGetter("proposals.CreateRoute")},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Proposal: %s", getters.GetterKey("proposal.Title")),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Proposals"),
			Url:   lago.RoutePathGetter("proposals.ListRoute"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{Title: getters.GetterStatic("Proposal Detail"), Url: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("proposal.ID"))},
			components.SidebarMenuItem{Title: getters.GetterStatic("Edit Proposal"), Url: getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("proposal.ID"))},
			components.SidebarMenuItem{Title: getters.GetterStatic("Delete Proposal"), Url: getters.GetterFormat(AppUrl+"%v/delete/", getters.GetterKey("proposal.ID"))},
		},
	})
}

func registerFilter() {
	lago.RegistryPage.Register("proposals.ProposalFilter", components.FormComponent{
		Url:    getters.GetterStatic(AppUrl),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Title", Name: "title", Getter: getters.GetterKey("$get.title")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})
}

func proposalFormFields() []components.PageInterface {
	inputs := []components.PageInterface{
		components.InputText{Label: "Proposal Title", Name: "title", Required: true, Getter: getters.GetterKey("$in.Title")},
	}
	for i := 0; i < len(QUESTIONS); i++ {
		key := fmt.Sprintf("$in.answer_%d", i)
		inputs = append(inputs, components.InputTextarea{
			Label:  fmt.Sprintf("Q%d: %s", i+1, QUESTIONS[i]),
			Name:   fmt.Sprintf("answers[%d]", i),
			Getter: getters.GetterKey(key),
			Rows:   2,
		})
	}
	return inputs
}

func registerForms() {
	lago.RegistryPage.Register("proposals.ProposalFormFields", components.ContainerColumn{
		Children: []components.PageInterface{
			components.ContainerColumn{Children: append(proposalFormFields(), components.ButtonSubmit{Label: "Save Proposal"})},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:    getters.GetterStatic(AppUrl + "create/"),
				Method: http.MethodPost,
				Title:  "Create Proposal",
				Subtitle: "Fill in the questionnaire answers",
				ChildrenInput: proposalFormFields(),
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Proposal"}},
			},
		},
	})

	lago.RegistryPage.Register("proposals.ProposalUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Getter:   getters.GetterKey("proposal"),
				Url:      getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("$in.ID")),
				Method:   http.MethodPost,
				Title:    "Edit Proposal",
				Subtitle: "Update questionnaire answers",
				ChildrenInput: proposalFormFields(),
				ChildrenAction: []components.PageInterface{components.ButtonSubmit{Label: "Save Proposal"}},
			},
		},
	})
}

func registerTable() {
	lago.RegistryPage.Register("proposals.ProposalTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalMenu"}},
		Children: []components.PageInterface{
			components.DataTable{
				UID:       "proposal-table",
				Data:      getters.GetterKey("proposals"),
				Title:     "Proposals",
				Subtitle:  "List of financial proposals",
				CreateUrl: lago.RoutePathGetter("proposals.CreateRoute"),
				OnClick:   getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("$row.ID")),
				FilterComponent: lago.DynamicPage{Name: "proposals.ProposalFilter"},
				Columns: []components.TableColumn{
					{Label: "Title", Key: "Title", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.Title")}}},
					{Label: "Created At", Key: "CreatedAt", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.CreatedAt")}}},
					{Label: "Updated At", Key: "UpdatedAt", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$row.UpdatedAt")}}},
				},
			},
		},
	})
}



func registerDetail() {
	generatedSection := []components.PageInterface{
		components.ContainerColumn{Classes: "mt-2 p-4 card card-body border rounded-box border-base-300", Children: []components.PageInterface{
			components.ContainerRow{Classes: "flex flex-wrap justify-between items-center gap-4 mb-4", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Generated Proposal")},
				components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
					components.ButtonLink{Label: "Export to PDF", Link: getters.GetterFormat(AppUrl+"%v/export-pdf/", getters.GetterKey("$in.ID")), Classes: "btn-secondary btn-sm"},
					components.ButtonLink{Label: "Edit with AI", Link: getters.GetterFormat(AppUrl+"%v/ai-edit/form/", getters.GetterKey("$in.ID")), Classes: "btn-outline btn-secondary btn-sm"},
					postButton("Regenerate Proposal", getters.GetterFormat(AppUrl+"%v/generate/", getters.GetterKey("$in.ID")), "arrow-path", "btn-outline btn-primary btn-sm"),
				}},
			}},
			components.FieldMarkdown{Getter: getters.GetterKey("$in.GeneratedContent"), Classes: "bg-base-100 p-8 rounded-lg shadow border"},
		}},
	}

	pendingSection := []components.PageInterface{
		components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
			components.FieldText{Getter: getters.GetterStatic("Generating..."), Classes: "btn-primary"},
			postButton("Cancel Generation", getters.GetterFormat(AppUrl+"%v/cancel/", getters.GetterKey("$in.ID")), "x-mark", "btn-outline btn-error btn-sm"),
		}},
	}

	idleSection := []components.PageInterface{
		postButton("Generate Proposal with AI", getters.GetterFormat(AppUrl+"%v/generate/", getters.GetterKey("$in.ID")), "sparkles", "btn-primary"),
	}

	lago.RegistryPage.Register("proposals.ProposalDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "proposals.ProposalDetailMenu"}},
		Children: []components.PageInterface{
			components.Detail{
				Getter: getters.GetterKey("proposal"),
				Children: []components.PageInterface{
					components.ContainerColumn{Children: []components.PageInterface{
						components.FieldTitle{Getter: getters.GetterKey("$in.Title")},
						components.LabelInline{Title: "Created At", Children: []components.PageInterface{components.FieldText{Getter: getters.GetterKey("$in.CreatedAt")}}},
						components.ContainerColumn{Classes: "mt-6", Children: []components.PageInterface{
							components.LabelInline{Title: "Questionnaire Answers", Children: []components.PageInterface{components.FieldKeyValue{Getter: getters.GetterKey("$in.Answers"), KeyField: "Question", ValueField: "Answer", Classes: "mt-6"}}},
						}},
						components.ContainerColumn{Classes: "mt-6", Children: []components.PageInterface{
							components.ShowIf{Getter: getters.GetterKey("$in.GeneratedContent"), Children: generatedSection},
	
							components.ShowIf{Getter: getters.GetterKey("generation_pending"), Children: pendingSection},
							components.ShowIf{Getter: getterIdleGeneration(proposalGenerationIdle), Children: idleSection},
						}},
					}},
				},
			},
		},
	})
}

func getterIdleGeneration(f func(context.Context) bool) getters.Getter {
	return func(ctx context.Context) any { return f(ctx) }
}

func proposalGenerationIdle(ctx context.Context) bool {
	content := getters.IfOrGetter(getters.GetterKey("$in.GeneratedContent"), ctx, nil)
	if content != nil && content != "" {
		return false
	}
if getters.IfOrGetter(getters.GetterKey("generation_pending"), ctx, nil) != nil {
		return false
	}
	if getters.IfOrGetter(getters.GetterKey("$in.GenerationID"), ctx, nil) != nil {
		return false
	}
	return true
}

func postButton(label string, actionUrl getters.Getter, icon, classes string) components.PageInterface {
	return components.FormComponent{
		Url: actionUrl, Method: http.MethodPost,
		ChildrenAction: []components.PageInterface{
			components.ButtonSubmit{Label: label, Classes: classes},
		},
	}
}

func registerModal() {
	lago.RegistryPage.Register("proposals.AiEditModal", components.Modal{
		UID:   "ai-edit-modal",
		Title: "Edit with AI",
		Children: []components.PageInterface{
			components.FormComponent{
				Getter: getters.GetterKey("proposal"),
				Url:    getters.GetterFormat(AppUrl+"%v/ai-edit/", getters.GetterKey("proposal.ID")),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputTextarea{Name: "generated_content", Label: "Current Proposal Markdown", Getter: getters.GetterKey("$in.GeneratedContent"), Rows: 8},
					components.InputTextarea{Name: "instructions", Label: "Instructions for AI", Getter: getters.GetterKey("$in.instructions"), Rows: 4, Required: true},
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
				CancelUrl: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("proposal.ID")),
			},
		},
	})
}
