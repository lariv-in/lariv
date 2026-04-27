package p_admissions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("admissions.ApplicationDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "admissions.ApplicationDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[AdmissionApplication]{
				Getter: getters.Key[AdmissionApplication]("application"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "admissions.ApplicationDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.ApplicantName")},
							&components.LabelInline{Title: "Program", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Program.Code")}}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")}}},
							&components.LabelInline{Title: "Email", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Email")}}},
							&components.LabelInline{Title: "Status", Children: []components.PageInterface{&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.Status"), ApplicationStatusChoices)}}},
							&components.LabelInline{Title: "Linked user", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.LinkedUser.Name")}}},
							&components.LabelInline{Title: "Remarks", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.Remarks")}}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("admissions.ApplicationDeleteForm", &components.Modal{
		UID: "admission-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm Deletion", Message: "Delete this application?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
