package p_nirmancampus_examregistrations

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/registry"
)

func registerFilterPages() {
	lago.RegistryPage.Register("examregistrations.Filter", &components.FormComponent[ExamRegistration]{
		Attr: getters.FormBoostedGet(lago.RoutePath("examregistrations.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Exam title",
				Name:   "ExamTitle",
				Getter: getters.Key[string]("$get.ExamTitle"),
			},
			&components.InputSelect[string]{
				Label:   "Registration status",
				Name:    "RegistrationStatus",
				Choices: getters.Static(ExamRegistrationStatusChoices),
				Getter:  registry.PairFromGetter(getters.Key[string]("$get.RegistrationStatus"), ExamRegistrationStatusChoices),
			},
			&components.InputForeignKey[p_nirmancampus_academicrecords.AcademicRecord]{
				Label:       "Academic record",
				Name:        "AcademicRecordID",
				Url:         lago.RoutePath("academicrecords.SelectRoute", nil),
				Display:     getters.Format("%s (%s)", getters.Any(getters.Key[string]("$in.Student.Name")), getters.Any(getters.Key[string]("$in.AdmissionSession.Name"))),
				Placeholder: "Filter by academic record...",
				Getter:      academicRecordForInputForeignKey(),
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func registerTablePages() {
	examRegistrationsSessionEnvironment := &components.Environment[uint]{
		Label:   "Admission session",
		Key:     getters.Static(examRegistrationsEnvironmentSessionKey),
		Options: p_nirmancampus_academicrecords.AcademicSessionsListGetter,
		Default: examRegistrationsSessionEnvironmentDefault,
	}
	lago.RegistryPage.Register("examregistrations.Table", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			examRegistrationsSessionEnvironment,
			&components.DataTable[ExamRegistration]{
				Page:    components.Page{Key: "examregistrations.TableBody"},
				UID:     "exam-registrations-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ExamRegistration]]("examregistrations"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "examregistrations.Filter"}},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("examregistrations.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Exam",
						Name:  "ExamTitle",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ExamTitle")},
						},
					},
					{
						Label: "Course",
						Name:  "Course.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Course.Name")},
						},
					},
					{
						Label: "Academic record",
						Name:  "AcademicRecord.Student.Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format(
									"%s (%s)",
									getters.Any(getters.Key[string]("$row.AcademicRecord.Student.Name")),
									getters.Any(getters.Key[string]("$row.AcademicRecord.AdmissionSession.Name")),
								),
							},
						},
					},
					{
						Label: "Status",
						Name:  "RegistrationStatus",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: registry.PairValueFromKey(
									getters.Key[string]("$row.RegistrationStatus"),
									ExamRegistrationStatusChoices,
								),
							},
						},
					},
				},
			},
		},
	})
}
