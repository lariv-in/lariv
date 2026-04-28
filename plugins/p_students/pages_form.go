package p_students

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func studentFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "students.StudentFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.StudentNo"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Student No",
						Name:     "StudentNo",
						Required: true,
						Getter:   getters.Key[string]("$in.StudentNo"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
					},
				},
			},
			&components.InputEmail{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$in.Email"),
			},
			&components.InputPhone{
				Label:  "Phone",
				Name:   "Phone",
				Getter: getters.Key[string]("$in.Phone"),
			},
			&components.InputText{Label: "Aadhaar / ID", Name: "AdhaarNo", Getter: getters.Key[string]("$in.AdhaarNo")},
			&components.InputDate{Label: "Date of birth", Name: "DOB", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB"))},
			&components.InputText{Label: "Gender", Name: "Gender", Getter: getters.Key[string]("$in.Gender")},
			&components.InputText{Label: "Nationality", Name: "Nationality", Getter: getters.Key[string]("$in.Nationality")},
			&components.InputText{Label: "Mother tongue", Name: "MotherTongue", Getter: getters.Key[string]("$in.MotherTongue")},
			&components.InputText{Label: "Religion", Name: "Religion", Getter: getters.Key[string]("$in.Religion")},
			&components.InputText{Label: "Caste", Name: "Caste", Getter: getters.Key[string]("$in.Caste")},
			&components.InputText{Label: "Category", Name: "Category", Getter: getters.Key[string]("$in.Category")},
			&components.InputTextarea{Label: "Special needs", Name: "SpecialNeeds", Rows: 2, Getter: getters.Key[string]("$in.SpecialNeeds")},
			&components.InputTextarea{Label: "Address", Name: "Address", Rows: 3, Getter: getters.Key[string]("$in.Address")},
			&components.InputText{Label: "Parent 1 name", Name: "Guardian1Name", Getter: getters.Key[string]("$in.Guardian1Name")},
			&components.InputEmail{Label: "Parent 1 email", Name: "Guardian1Email", Getter: getters.Key[string]("$in.Guardian1Email")},
			&components.InputPhone{Label: "Parent 1 phone", Name: "Guardian1Phone", Getter: getters.Key[string]("$in.Guardian1Phone")},
			&components.InputText{Label: "Parent 2 name", Name: "Guardian2Name", Getter: getters.Key[string]("$in.Guardian2Name")},
			&components.InputEmail{Label: "Parent 2 email", Name: "Guardian2Email", Getter: getters.Key[string]("$in.Guardian2Email")},
			&components.InputPhone{Label: "Parent 2 phone", Name: "Guardian2Phone", Getter: getters.Key[string]("$in.Guardian2Phone")},
			&components.InputText{Label: "Previous school name", Name: "PrevSchoolName", Getter: getters.Key[string]("$in.PrevSchoolName")},
			&components.InputTextarea{Label: "Previous school address", Name: "PrevSchoolAddress", Rows: 2, Getter: getters.Key[string]("$in.PrevSchoolAddress")},
			&components.InputText{Label: "Previous school class", Name: "PrevSchoolClass", Getter: getters.Key[string]("$in.PrevSchoolClass")},
			&components.InputDate{Label: "Previous school pass date", Name: "PrevSchoolPassDate", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.PrevSchoolPassDate"))},
			&components.InputText{Label: "Previous school UDISE", Name: "PrevSchoolUDISECode", Getter: getters.Key[string]("$in.PrevSchoolUDISECode")},
			&components.InputManyToMany[p_filesystem.VNode]{
				Label: "Documents", Name: "Documents", Required: false,
				Getter: getters.Key[[]p_filesystem.VNode]("$in.Documents"), Display: getters.Key[string]("$in.Name"),
				Url: lago.RoutePath("filesystem.MultiSelectRoute", nil), Placeholder: "Select files…", Classes: "w-full",
			},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("students.StudentDeleteForm")

	lago.RegistryPage.Register("students.StudentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("students.StudentCreateForm"),
				ActionURL: lago.RoutePath("students.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Attr:     getters.FormBubbling(getters.Static("students.StudentCreateForm")),
						Title:    "Create Student",
						Subtitle: "Add a student record",
						ChildrenInput: []components.PageInterface{
							studentFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Student"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("students.StudentUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("students.StudentUpdateForm"),
				ActionURL: lago.RoutePath("students.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Student]{
						Getter: getters.Key[Student]("student"),
						Attr:   getters.FormBubbling(getters.Static("students.StudentUpdateForm")),
						Title:  "Edit Student",
						ChildrenInput: []components.PageInterface{
							studentFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Student"},
									&components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Name:        deleteFormName,
										Url:         lago.RoutePath("students.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
										FormPostURL: lago.RoutePath("students.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student.ID"))}),
										ModalUID:    "student-delete-modal",
										Classes:     "btn-error",
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
