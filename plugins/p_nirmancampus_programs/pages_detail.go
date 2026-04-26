package p_nirmancampus_programs

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
)

func registerDetailPages() {
	programStructureNonEmpty := getters.Map(
		getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits"),
		func(_ context.Context, units []ProgramStructureUnit) (bool, error) {
			return len(units) > 0, nil
		},
	)

	lago.RegistryPage.Register("programs.ProgramDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Program]{
				Getter: getters.Key[Program]("program"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "programs.ProgramDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title: "University",
								Children: []components.PageInterface{
									&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.University"), UniversityChoices)},
								},
							},
							&components.LabelInline{
								Title: "Program type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.ProgramType"), programTypeChoices)},
								},
							},
							&components.LabelInline{
								Title: "Admission sessions",
								Children: []components.PageInterface{
									&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.AdmissionSessions"), admissionSessionChoices)},
								},
							},
							&components.LabelInline{
								Title: "Term type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: registry.PairValueFromKey(getters.Key[string]("$in.TermType"), termTypeChoices)},
								},
							},
							&components.LabelInline{
								Title: "Program fee",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("₹ %d", getters.Any(getters.Key[uint]("$in.ProgramFee")))},
								},
							},
							&components.LabelNewline{
								Title: "Media languages",
								Children: []components.PageInterface{
									&components.FieldManyToMany[ProgramMedia]{
										Getter:  getters.Key[[]ProgramMedia]("$in.ProgramMedia"),
										Display: getters.Key[string]("$in.Language"),
										Classes: "w-full",
									},
								},
							},
							&components.LabelNewline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Description")},
								},
							},
							&components.LabelNewline{
								Title: "Program structure",
								Children: []components.PageInterface{
									&components.ShowIf{
										Getter: getters.Any(programStructureNonEmpty),
										Children: []components.PageInterface{
											&components.FieldList[ProgramStructureUnit]{
												Getter:  getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits"),
												Classes: "flex flex-col gap-2",
												Children: []components.PageInterface{
													&components.ContainerColumn{
														Classes: "rounded-box border border-base-300 p-2 card card-body",
														Children: []components.PageInterface{
															&components.LabelInline{
																Title: "Term",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: getters.Format(
																			"%d",
																			getters.Any(getters.Key[uint]("$row.TermNumber")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Compulsory",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListDisplayGetter(
																			getters.Key[[]courses.Course]("$row.CompulsoryCourses"),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional count",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: getters.Format(
																			"%d",
																			getters.Any(getters.Key[uint]("$row.OptionalCourseCount")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional course pool",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListDisplayGetter(
																			getters.Key[[]courses.Course]("$row.OptionalCourseSelectionPool"),
																		),
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									&components.ShowIf{
										Getter: getters.BoolNot(programStructureNonEmpty),
										Children: []components.PageInterface{
											&components.ButtonLink{
												Page:  components.Page{Roles: []string{"admin", "superuser"}},
												Label: "Add Program Structure",
												Link: lago.RoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("$in.ID")),
												}),
												Classes: "btn-primary btn-sm w-fit",
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
}

func registerDeletePages() {
	lago.RegistryPage.Register("programs.ProgramDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "program-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this program?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
