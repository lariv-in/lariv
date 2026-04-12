package p_nirmancampus_studentapplications

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
)

func registerDetailPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "studentapplications.ApplicationDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentApplication]{
				Getter: getters.Key[StudentApplication]("studentapplication"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "studentapplications.ApplicationDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Key[string]("$in.StudentName"),
							},
							&components.FieldSubtitle{
								Getter: getters.Key[string]("$in.Email"),
							},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: p_nirmancampus_programs.ProgramDisplayLabel(
										getters.Key[string]("$in.Program.Name"),
										getters.Key[string]("$in.Program.University"),
									)},
								},
							},
							&components.LabelInline{
								Title: "Date of birth",
								Children: []components.PageInterface{
									&components.FieldDate{Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB"))},
								},
							},
							&components.LabelInline{
								Title: "Mother name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.MotherName")},
								},
							},
							&components.LabelInline{
								Title: "Father name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.FatherName")},
								},
							},
							&components.LabelInline{
								Title: "Category",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Category")},
								},
							},
							&components.LabelInline{
								Title: "Mobile",
								Children: []components.PageInterface{
									&components.FieldPhone{Getter: getters.Key[string]("$in.Mobile")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Email")},
								},
							},
							&components.LabelNewline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Address")},
								},
							},
							&components.LabelNewline{
								Title: "Photo",
								Children: []components.PageInterface{
									&p_filesystem.FieldPhoto{
										VNode:   getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$in.PhotoID"))),
										Classes: "w-42 rounded",
									},
								},
							},
							&components.LabelNewline{
								Title: "Documents",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{
										VNode: getters.Key[[]p_filesystem.VNode]("$in.Documents"),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "studentapplication-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this application?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
