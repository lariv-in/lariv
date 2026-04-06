package p_nirmancampus_students

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

// studentFormUserPickURL opens the scoped user picker; on edit, allow_user_id keeps the linked user visible.
func registerDetailPages() {
	lago.RegistryPage.Register("students.StudentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Student]{
				Getter: getters.Key[Student]("student"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "students.StudentDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Key[string]("$in.User.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.Key[string]("$in.StudentNo"),
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.User.Email")},
								},
							},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldPhone{Getter: getters.Key[string]("$in.User.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Date of Birth",
								Children: []components.PageInterface{
									&components.FieldDate{
										Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB")),
									},
								},
							},
							&components.LabelInline{
								Title: "Mother's Name",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.MotherName")},
								},
							},
							&components.LabelInline{
								Title: "Father's Name",
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

	lago.RegistryPage.Register("students.StudentDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "student-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this student?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
