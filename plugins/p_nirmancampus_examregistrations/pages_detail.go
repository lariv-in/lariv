package p_nirmancampus_examregistrations

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	lago.RegistryPage.Register("examregistrations.Detail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "examregistrations.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[ExamRegistration]{
				Getter: getters.Key[ExamRegistration]("examregistration"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "examregistrations.DetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.ExamTitle")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Course.Name")},
							&components.LabelInline{
								Title: "Registration status",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: registry.PairValueFromKey(
											getters.Key[string]("$in.RegistrationStatus"),
											ExamRegistrationStatusChoices,
										),
									},
								},
							},
							&components.LabelInline{
								Title: "Fee",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("₹ %d", getters.Any(getters.Key[uint]("$in.Fee")))},
								},
							},
							&components.LabelInline{
								Title: "Academic record",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%s (%s)", getters.Any(getters.Key[string]("$in.AcademicRecord.Student.Name")), getters.Any(getters.Key[string]("$in.AcademicRecord.AdmissionSession.Name")))},
								},
							},
							&components.LabelInline{
								Title: "Assets",
								Children: []components.PageInterface{
									&p_filesystem.FieldManyFile{VNode: getters.Key[[]p_filesystem.VNode]("$in.Assets")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("examregistrations.DeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "examregistration-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm deletion",
				Message: "Are you sure you want to delete this exam registration?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
