package p_nirmancampus_programs

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
)

func programAdmissionSessionsDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.AdmissionSessions")(ctx)
		if err != nil || s == "" {
			return "—", nil
		}
		switch s {
		case AdmissionSessionJan:
			return "January", nil
		case AdmissionSessionJuly:
			return "July", nil
		case AdmissionSessionBoth:
			return "January and July", nil
		default:
			return s, nil
		}
	}
}

func programTermTypeDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.TermType")(ctx)
		if err != nil || s == "" {
			return "—", nil
		}
		switch s {
		case TermTypeYear:
			return "Year", nil
		case TermTypeSession:
			return "Session", nil
		default:
			return s, nil
		}
	}
}

func courseListDisplayGetter(g getters.Getter[[]courses.Course]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		list, err := g(ctx)
		if err != nil {
			return "", err
		}
		if len(list) == 0 {
			return "—", nil
		}
		codes := make([]string, 0, len(list))
		for _, c := range list {
			codes = append(codes, c.Code)
		}
		return strings.Join(codes, ", "), nil
	}
}

func programStructureRowsGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		units, err := getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
		if err != nil {
			return nil, err
		}
		if units == nil {
			return []ProgramStructureUnit{}, nil
		}
		return units, nil
	}
}

func programStructureNonEmptyGetter() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		units, err := getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
		if err != nil {
			return false, err
		}
		return len(units) > 0, nil
	}
}

func registerDetailPages() {
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
									&components.FieldText{Getter: getters.Key[string]("$in.University")},
								},
							},
							&components.LabelInline{
								Title: "Program type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.ProgramType")},
								},
							},
							&components.LabelInline{
								Title: "Admission sessions",
								Children: []components.PageInterface{
									&components.FieldText{Getter: programAdmissionSessionsDisplayGetter()},
								},
							},
							&components.LabelInline{
								Title: "Term type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: programTermTypeDisplayGetter()},
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
										Getter: getters.Any(programStructureNonEmptyGetter()),
										Children: []components.PageInterface{
											&components.FieldList{
												Getter:  programStructureRowsGetter(),
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
										Getter: getters.BoolNot(programStructureNonEmptyGetter()),
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
				Attr:    getters.FormBubbling(nil),
			},
		},
	})
}
