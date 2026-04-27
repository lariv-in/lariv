package p_admissions

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
)

func applicationFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "admissions.ApplicationFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{Error: getters.Key[error]("$error.ProgramID"), Children: []components.PageInterface{&components.InputForeignKey[p_programs.Program]{Label: "Program", Name: "ProgramID", Required: true, Url: lago.RoutePath("programs.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select program...", Getter: getters.Association[p_programs.Program](getters.Key[uint]("$in.ProgramID"))}}},
			&components.ContainerError{Error: getters.Key[error]("$error.SemesterID"), Children: []components.PageInterface{&components.InputForeignKey[p_semesters.Semester]{Label: "Semester", Name: "SemesterID", Required: true, Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Select semester…", Getter: getters.Association[p_semesters.Semester](getters.Key[uint]("$in.SemesterID"))}}},
			&components.ContainerError{Error: getters.Key[error]("$error.ApplicantName"), Children: []components.PageInterface{&components.InputText{Label: "Applicant Name", Name: "ApplicantName", Required: true, Getter: getters.Key[string]("$in.ApplicantName")}}},
			&components.InputEmail{Label: "Email", Name: "Email", Getter: getters.Key[string]("$in.Email")},
			&components.ContainerError{Error: getters.Key[error]("$error.Status"), Children: []components.PageInterface{&components.InputSelect[string]{Label: "Status", Name: "Status", Required: true, Choices: getters.Static(ApplicationStatusChoices), Getter: registry.PairFromGetter(getters.Key[string]("$in.Status"), ApplicationStatusChoices)}}},
			&components.InputForeignKey[p_users.User]{Label: "Linked user (optional)", Name: "UserID", Required: false, Url: lago.RoutePath("users.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Select user…", Getter: getters.Association[p_users.User](getters.Deref(getters.Key[*uint]("$in.UserID")))},
			&components.InputTextarea{Label: "Remarks", Name: "Remarks", Rows: 3, Getter: getters.Key[string]("$in.Remarks")},
			&components.InputText{Label: "Aadhaar / ID", Name: "AdhaarNo", Getter: getters.Key[string]("$in.AdhaarNo")},
			&components.InputDate{Label: "Date of birth", Name: "DOB", Required: false, Getter: getters.Deref(getters.Key[*time.Time]("$in.DOB"))},
			&components.InputText{Label: "Gender", Name: "Gender", Getter: getters.Key[string]("$in.Gender")},
			&components.InputText{Label: "Nationality", Name: "Nationality", Getter: getters.Key[string]("$in.Nationality")},
			&components.InputTextarea{Label: "Address", Name: "Address", Rows: 3, Getter: getters.Key[string]("$in.Address")},
		},
	}
}

func registerFormPages() {
	deleteFormName := getters.Static("admissions.ApplicationDeleteForm")
	lago.RegistryPage.Register("admissions.ApplicationCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "admissions.ApplicationMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("admissions.ApplicationCreateForm"), ActionURL: lago.RoutePath("admissions.CreateRoute", nil), Children: []components.PageInterface{
				&components.FormComponent[AdmissionApplication]{Attr: getters.FormBubbling(getters.Static("admissions.ApplicationCreateForm")), Title: "Create Application", ChildrenInput: []components.PageInterface{applicationFormFields()}, ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save Application"}}},
			}},
		},
	})
	lago.RegistryPage.Register("admissions.ApplicationUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "admissions.ApplicationDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{Name: getters.Static("admissions.ApplicationUpdateForm"), ActionURL: lago.RoutePath("admissions.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))}), Children: []components.PageInterface{
				&components.FormComponent[AdmissionApplication]{Getter: getters.Key[AdmissionApplication]("application"), Attr: getters.FormBubbling(getters.Static("admissions.ApplicationUpdateForm")), Title: "Edit Application", ChildrenInput: []components.PageInterface{applicationFormFields()}, ChildrenAction: []components.PageInterface{
					&components.ContainerRow{Classes: "flex gap-2 items-center", Children: []components.PageInterface{
						&components.ButtonSubmit{Label: "Save Application"},
						&components.ButtonModalForm{Label: "Delete", Icon: "trash", Name: deleteFormName, Url: lago.RoutePath("admissions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))}), FormPostURL: lago.RoutePath("admissions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))}), ModalUID: "admission-delete-modal", Classes: "btn-error"},
					}},
				}},
			}},
		},
	})
}
