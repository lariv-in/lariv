package p_nirmancampus_assignmentsubmissions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/registry"
)

func registerFilterPages() {
	lago.RegistryPage.Register("assignmentsubmissions.Filter", &components.FormComponent[AssignmentSubmission]{
		Attr: getters.FormBoostedGet(lago.RoutePath("assignmentsubmissions.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Assignment title",
				Name:   "AssignmentTitle",
				Getter: getters.Key[string]("$get.AssignmentTitle"),
			},
			&components.InputSelect[string]{
				Label:   "Submission status",
				Name:    "SubmissionStatus",
				Choices: getters.Static(AssignmentSubmissionStatusChoices),
				Getter:  registry.PairFromGetter(getters.Key[string]("$get.SubmissionStatus"), AssignmentSubmissionStatusChoices),
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
	assignmentSubmissionsSessionEnvironment := &components.Environment[uint]{
		Label:   "Admission session",
		Key:     getters.Static(assignmentSubmissionsEnvironmentSessionKey),
		Options: p_nirmancampus_academicrecords.AcademicSessionsListGetter,
		Default: assignmentSubmissionsSessionEnvironmentDefault,
	}
	lago.RegistryPage.Register("assignmentsubmissions.Table", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			assignmentSubmissionsSessionEnvironment,
			&components.DataTable[AssignmentSubmission]{
				Page:    components.Page{Key: "assignmentsubmissions.TableBody"},
				UID:     "assignment-submissions-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AssignmentSubmission]]("assignmentsubmissions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "assignmentsubmissions.Filter"}},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{
						Label: "Assignment",
						Name:  "AssignmentTitle",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.AssignmentTitle")},
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
						Name:  "SubmissionStatus",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: registry.PairValueFromKey(
									getters.Key[string]("$row.SubmissionStatus"),
									AssignmentSubmissionStatusChoices,
								),
							},
						},
					},
				},
			},
		},
	})
}
