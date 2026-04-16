package p_nirmancampus_students

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

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
								Getter: getters.Key[string]("$in.Name"),
							},
							&components.FieldSubtitle{
								Getter: getters.Key[string]("$in.StudentNo"),
							},
							&components.LabelInline{
								Title: "Aadhar Card",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.AadharCard")},
								},
							},
							&components.LabelInline{
								Title: "ABC ID",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.ABCId")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Email")},
								},
							},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldPhone{Getter: getters.Key[string]("$in.Phone")},
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
							&components.LabelInline{
								Title: "Handicapped",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.Handicapped")},
								},
							},
							&components.LabelNewline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Address")},
								},
							},
							&components.LabelNewline{
								Title: "Remarks",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Key[string]("$in.Remarks"),
									},
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
				Attr:    getters.FormBubbling(getters.Key[string]("student.Name")),
			},
		},
	})
}
