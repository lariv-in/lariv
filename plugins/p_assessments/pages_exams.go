package p_assessments

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_syllabus"
)

// Step 1: everything except topics (so step 2 topic picker URL includes CourseID from $in).
func assessmentFormStage1Fields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assessments.AssessmentFormStage1Fields"},
		Children: []components.PageInterface{
			&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 3, Getter: getters.Key[string]("$in.Description")},
			&components.InputCheckbox{Label: "Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
			&components.InputTextarea{Label: "Syllabus notes", Name: "Syllabus", Rows: 2, Getter: getters.Key[string]("$in.Syllabus")},
			&components.InputDatetime{Label: "When", Name: "WhenAt", Required: true, Getter: getters.Key[time.Time]("$in.WhenAt")},
			&components.InputText{Label: "Venue", Name: "Venue", Getter: getters.Key[string]("$in.Venue")},
			&components.InputForeignKey[p_courses.Course]{
				Label: "Course (optional)", Name: "CourseID", Required: false,
				Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course…",
				Getter: getters.Association[p_courses.Course](getters.Deref(getters.Key[*uint]("$in.CourseID"))),
			},
			&components.InputForeignKey[p_semesters.Semester]{
				Label: "Semester (optional)", Name: "SemesterID", Required: false,
				Url: lago.RoutePath("semesters.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "Select semester…",
				Getter: getters.Association[p_semesters.Semester](getters.Deref(getters.Key[*uint]("$in.SemesterID"))),
			},
			&components.InputNumber[int]{Label: "Max marks", Name: "MaxMarks", Required: false, Getter: getters.Key[int]("$in.MaxMarks")},
			&components.InputNumber[int]{Label: "Passing marks", Name: "PassingMarks", Required: false, Getter: getters.Key[int]("$in.PassingMarks")},
		},
	}
}

func assessmentFormStage2Fields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "assessments.AssessmentFormStage2Fields"},
		Children: []components.PageInterface{
			&components.InputManyToMany[p_syllabus.SyllabusTopic]{
				Label: "Topics", Name: "Topics", Required: false,
				Getter: getters.Key[[]p_syllabus.SyllabusTopic]("$in.Topics"), Display: getters.Key[string]("$in.Title"),
				Url: syllabusTopicMultiSelectURL(), Placeholder: "Select topics…", Classes: "w-full",
			},
		},
	}
}

func registerExamPages() {
	dn := getters.Static("assessments.AssessmentDeleteForm")

	lago.RegistryPage.Register("assessments.AssessmentMenu", &components.SidebarMenu{
		Title: getters.Static("Exams"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All exams"), Url: lago.RoutePath("assessments.ExamDefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create exam"), Url: lago.RoutePath("assessments.ExamCreateRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Grade entries"), Url: lago.RoutePath("assessments.DefaultRoute", nil)},
		},
	})
	lago.RegistryPage.Register("assessments.AssessmentDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Exam: %s", getters.Any(getters.Key[string]("assessment.Title"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to exams"), Url: lago.RoutePath("assessments.ExamDefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("assessments.ExamDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assessment.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("assessments.ExamUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assessment.ID"))})},
		},
	})

	lago.RegistryPage.Register("assessments.AssessmentTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.AssessmentMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Assessment]{
				Page:    components.Page{Key: "assessments.AssessmentTableBody"},
				UID:     "assessment-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Assessment]]("assessments"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("assessments.ExamCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("assessments.ExamDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "When", Name: "WhenAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.WhenAt")}}},
					{Label: "Active", Name: "IsActive", Children: []components.PageInterface{&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")}}},
				},
			},
		},
	})

	lago.RegistryPage.Register("assessments.AssessmentCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.AssessmentMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assessments.AssessmentCreateForm"),
				ActionURL: lago.RoutePath("assessments.ExamCreateRoute", nil),
				Children: []components.PageInterface{
					&components.MultiStepForm{
						Page:          components.Page{Key: "assessments.AssessmentCreateMultiStep"},
						MultiStageURL: assessmentExamFormStageURLGetter(),
						Stages: []components.FormInterface{
							&components.FormComponent[Assessment]{
								Attr:     getters.FormBubbling(getters.Static("assessments.AssessmentCreateForm")),
								Title:    "Create exam",
								Subtitle: "Details and course. Course sets which syllabus topics appear on the next step.",
								ChildrenInput: []components.PageInterface{
									assessmentFormStage1Fields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2 mt-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Continue", Classes: "btn-primary"},
										},
									},
								},
							},
							&components.FormComponent[Assessment]{
								Attr:  getters.FormBubbling(getters.Static("assessments.AssessmentCreateForm")),
								Title: "Topics", Subtitle: "Syllabus topics. Scoped by course when you chose one on step 1.",
								ChildrenInput: []components.PageInterface{
									assessmentFormStage2Fields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2 mt-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save exam", Classes: "btn-primary"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assessments.AssessmentUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.AssessmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("assessments.AssessmentUpdateForm"),
				ActionURL: lago.RoutePath("assessments.ExamUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assessment.ID"))}),
				Children: []components.PageInterface{
					&components.MultiStepForm{
						Page:          components.Page{Key: "assessments.AssessmentUpdateMultiStep"},
						MultiStageURL: assessmentExamFormStageURLGetter(),
						Stages: []components.FormInterface{
							&components.FormComponent[Assessment]{
								Getter:   getters.Key[Assessment]("assessment"),
								Attr:     getters.FormBubbling(getters.Static("assessments.AssessmentUpdateForm")),
								Title:    "Edit exam",
								Subtitle: "Details and course. Course sets which syllabus topics appear on the next step.",
								ChildrenInput: []components.PageInterface{
									assessmentFormStage1Fields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2 mt-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Continue", Classes: "btn-primary"},
										},
									},
								},
							},
							&components.FormComponent[Assessment]{
								Getter:  getters.Key[Assessment]("assessment"),
								Attr:    getters.FormBubbling(getters.Static("assessments.AssessmentUpdateForm")),
								Title:   "Topics", Subtitle: "Syllabus topics. Scoped by course when you chose one on step 1.",
								ChildrenInput: []components.PageInterface{
									assessmentFormStage2Fields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex flex-wrap gap-2 items-center justify-end mt-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save exam", Classes: "btn-primary"},
											&components.ButtonModalForm{
												Label: "Delete", Icon: "trash", Name: dn,
												Url:         lago.RoutePath("assessments.ExamDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assessment.ID"))}),
												FormPostURL: lago.RoutePath("assessments.ExamDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assessment.ID"))}),
												ModalUID:    "assessment-delete-modal", Classes: "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assessments.AssessmentDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "assessments.AssessmentDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[Assessment]{
				Getter: getters.Key[Assessment]("assessment"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "assessments.AssessmentDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "When", Children: []components.PageInterface{
								&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.WhenAt")},
							}},
							&components.LabelInline{Title: "Venue", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Venue")},
							}},
							&components.LabelInline{Title: "Active", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
							}},
							&components.LabelInline{Title: "Max / passing marks", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d / %d", getters.Any(getters.Key[int]("$in.MaxMarks")), getters.Any(getters.Key[int]("$in.PassingMarks")))},
							}},
							&components.LabelInline{Title: "Description", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Description")},
							}},
							&components.FieldManyToMany[p_syllabus.SyllabusTopic]{
								Label: "Topics", Getter: getters.Key[[]p_syllabus.SyllabusTopic]("$in.Topics"),
								Display: getters.Key[string]("$in.Title"),
								Link: lago.RoutePath("syllabus.DetailRoute", map[string]getters.Getter[any]{
									"id": getters.Any(getters.Key[uint]("$in.ID")),
								}),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assessments.AssessmentDeleteForm", &components.Modal{UID: "assessment-delete-modal", Children: []components.PageInterface{
		&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this exam?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
	}})
}
