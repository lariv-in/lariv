package p_nirmancampus_academicrecords

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
)

func registerDetailPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AcademicRecord]{
				Getter: getters.Key[AcademicRecord]("academicrecord"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "academicrecords.AcademicRecordDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Student.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Student.StudentNo")},
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
								Title: "Admission session",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.AdmissionSession.Name")},
								},
							},
							&components.LabelInline{
								Title: "Status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Status")},
								},
							},
							&components.LabelInline{
								Title: "Admission date",
								Children: []components.PageInterface{
									&components.FieldDate{Getter: getters.Key[time.Time]("$in.Date")},
								},
							},
							&components.LabelInline{
								Title: "Term",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.ProgramStructureUnit.TermNumber"))),
									},
								},
							},
							&components.LabelNewline{
								Title: "Compulsory courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.Key[[]p_nirmancampus_courses.Course]("$in.CompulsoryCourses"),
										Display: getters.Key[string]("$in.Name"),
										Link:    courseDetailLink,
										Classes: "w-full",
									},
								},
							},
							&components.LabelNewline{
								Title: "Optional courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.Key[[]p_nirmancampus_courses.Course]("$in.OptionalCourses"),
										Display: getters.Key[string]("$in.Name"),
										Link:    courseDetailLink,
										Classes: "w-full",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "academicrecord-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this academic record?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}

// --- Selection Tables ---
