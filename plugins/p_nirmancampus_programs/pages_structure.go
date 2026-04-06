package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func joinCourseCodes(list []courses.Course) string {
	if len(list) == 0 {
		return "—"
	}
	codes := make([]string, 0, len(list))
	for _, c := range list {
		codes = append(codes, c.Code)
	}
	return strings.Join(codes, ", ")
}

// programStructureUnitCards lists units as cards with an edit-in-modal control per row.
type programStructureUnitCards struct {
	components.Page
}

func (programStructureUnitCards) GetRoles() []string { return nil }

func (e programStructureUnitCards) GetKey() string { return e.Key }

func (programStructureUnitCards) GetChildren() []components.PageInterface { return nil }

func (e programStructureUnitCards) Build(ctx context.Context) Node {
	units, err := getters.Key[[]ProgramStructureUnit]("program.ProgramStructureUnits")(ctx)
	if err != nil || len(units) == 0 {
		return P(Class("text-base-content/70"), Text("No structure units yet. Use “Add new unit” to create one."))
	}
	var nodes []Node
	for i := range units {
		u := units[i]
		editURL, err := lago.RoutePath("programs.StructureUnitEditModalRoute", map[string]getters.Getter[any]{
			"id":     getters.Any(getters.Key[uint]("program.ID")),
			"unitId": getters.Any(getters.Static[uint](u.ID)),
		})(ctx)
		if err != nil {
			editURL = "#"
		}
		pool := joinCourseCodes(u.OptionalCourseSelectionPool)
		comp := joinCourseCodes(u.CompulsoryCourses)
		nodes = append(nodes,
			Div(Class("rounded-box border border-base-300 p-2 flex flex-col gap-2 @md:flex-row @md:items-start @md:justify-between"),
				Div(Class("flex flex-col gap-1 min-w-0"),
					Div(Class("font-semibold"), Text(fmt.Sprintf("Term %d", u.TermNumber))),
					Div(Class("text-sm text-base-content/80"), Text("Compulsory: "+comp)),
					Div(Class("text-sm text-base-content/80"), Text(fmt.Sprintf("Optional count: %d", u.OptionalCourseCount))),
					Div(Class("text-sm text-base-content/80"), Text("Optional pool: "+pool)),
				),
				Div(Class("flex flex-wrap gap-2 shrink-0"),
					components.Render(&components.ButtonModalForm{
						Label: "Edit",
						Name:  getters.Static("programs.StructureUnitEditModal"),
						Url:   getters.Static(editURL),
						FormPostURL: lago.RoutePath("programs.StructureUnitUpdateRoute", map[string]getters.Getter[any]{
							"id":     getters.Any(getters.Key[uint]("program.ID")),
							"unitId": getters.Any(getters.Static[uint](u.ID)),
						}),
						ModalUID: "structure-unit-edit-modal",
						Classes:  "btn-outline btn-sm",
					}, ctx),
					components.Render(&components.ButtonModalForm{
						Label: "Remove",
						Name:  getters.Static("programs.StructureUnitDeleteForm"),
						Url: lago.RoutePath("programs.StructureUnitDeleteRoute", map[string]getters.Getter[any]{
							"id":     getters.Any(getters.Key[uint]("program.ID")),
							"unitId": getters.Any(getters.Static[uint](u.ID)),
						}),
						FormPostURL: lago.RoutePath("programs.StructureUnitDeleteRoute", map[string]getters.Getter[any]{
							"id":     getters.Any(getters.Key[uint]("program.ID")),
							"unitId": getters.Any(getters.Static[uint](u.ID)),
						}),
						ModalUID: "structure-unit-delete-modal",
						Classes:  "btn-outline btn-error btn-sm",
					}, ctx),
				),
			),
		)
	}
	return Div(Class("flex flex-col gap-2 my-4"), Group(nodes))
}

func registerStructurePages() {
	lago.RegistryPage.Register("programs.ProgramStructureEditPage", &components.ShellScaffold{
		Page: components.Page{
			Key:   "programs.ProgramStructureEditPage",
			Roles: []string{"admin", "superuser"},
		},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Classes: "max-w-3xl flex flex-col",
				Children: []components.PageInterface{
					&components.FieldTitle{
						Getter: getters.Static("Edit program structure"),
					},
					&components.FieldSubtitle{
						Getter: getters.Key[string]("program.Name"),
					},
					programStructureUnitCards{},
					&components.ButtonModalForm{
						Label:       "Add new unit",
						Url:         lago.RoutePath("programs.StructureUnitCreateModalRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}),
						FormPostURL: lago.RoutePath("programs.StructureUnitCreateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}),
						ModalUID:    "structure-unit-create-modal",
						Classes:     "btn-primary",
					},
				},
			},
		},
	})

	structureUnitFormFieldsEdit := func() components.ContainerColumn {
		return components.ContainerColumn{
			Page: components.Page{Key: "programs.StructureUnitFormFieldsBody"},
			Children: []components.PageInterface{
				&components.ContainerError{
					Error: getters.Key[error]("$error.TermNumber"),
					Children: []components.PageInterface{
						&components.InputNumber[uint]{
							Label:    "Term number",
							Name:     "TermNumber",
							Required: true,
							Getter:   getters.Key[uint]("$in.TermNumber"),
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.CompulsoryCourses"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Compulsory courses",
							Name:        "CompulsoryCourses",
							Getter:      getters.Key[[]courses.Course]("$in.CompulsoryCourses"),
							Url:         lago.RoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.Key[string]("$in.Name"),
							Placeholder: "Select compulsory courses…",
							Classes:     "w-full",
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.OptionalCourseCount"),
					Children: []components.PageInterface{
						&components.InputNumber[uint]{
							Label:    "Optional course count",
							Name:     "OptionalCourseCount",
							Required: false,
							Getter:   getters.Key[uint]("$in.OptionalCourseCount"),
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.OptionalCourseSelectionPool"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Optional course pool",
							Name:        "OptionalCourseSelectionPool",
							Getter:      getters.Key[[]courses.Course]("$in.OptionalCourseSelectionPool"),
							Url:         lago.RoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.Key[string]("$in.Name"),
							Placeholder: "Select optional pool courses…",
							Classes:     "w-full",
						},
					},
				},
			},
		}
	}

	structureUnitFormFieldsCreate := func() components.ContainerColumn {
		return components.ContainerColumn{
			Page: components.Page{Key: "programs.StructureUnitFormFieldsCreateBody"},
			Children: []components.PageInterface{
				&components.ContainerError{
					Error: getters.Key[error]("$error.TermNumber"),
					Children: []components.PageInterface{
						&components.InputNumber[uint]{
							Label:    "Term number",
							Name:     "TermNumber",
							Required: true,
							Getter:   getters.Static[uint](0),
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.CompulsoryCourses"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Compulsory courses",
							Name:        "CompulsoryCourses",
							Getter:      getters.Key[[]courses.Course]("$in.CompulsoryCourses"),
							Url:         lago.RoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.Key[string]("$in.Name"),
							Placeholder: "Select compulsory courses…",
							Classes:     "w-full",
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.OptionalCourseCount"),
					Children: []components.PageInterface{
						&components.InputNumber[uint]{
							Label:    "Optional course count",
							Name:     "OptionalCourseCount",
							Required: false,
							Getter:   getters.Static[uint](0),
						},
					},
				},
				&components.ContainerError{
					Error: getters.Key[error]("$error.OptionalCourseSelectionPool"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Optional course pool",
							Name:        "OptionalCourseSelectionPool",
							Getter:      getters.Key[[]courses.Course]("$in.OptionalCourseSelectionPool"),
							Url:         lago.RoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.Key[string]("$in.Name"),
							Placeholder: "Select optional pool courses…",
							Classes:     "w-full",
						},
					},
				},
			},
		}
	}

	lago.RegistryPage.Register("programs.StructureUnitCreateModal", components.Modal{
		UID: "structure-unit-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[ProgramStructureUnit]{
				Attr: getters.FormBubbling(getters.Static("programs.StructureUnitCreateModal")),

				Title: "Add structure unit",
				ChildrenInput: []components.PageInterface{
					&components.InputText{
						Hidden: true,
						Name:   "ProgramID",
						Getter: getters.Format("%d", getters.Any(getters.Key[uint]("program.ID"))),
					},
					structureUnitFormFieldsCreate(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save unit", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.StructureUnitDeleteForm", &components.Modal{
		Page: components.Page{
			Key:   "programs.StructureUnitDeleteForm",
			Roles: []string{"admin", "superuser"},
		},
		UID: "structure-unit-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Remove structure unit",
				Message: "This removes the term from the program structure. Course links for this unit will be cleared.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})

	lago.RegistryPage.Register("programs.StructureUnitEditModal", components.Modal{
		UID: "structure-unit-edit-modal",
		Children: []components.PageInterface{
			&components.FormComponent[ProgramStructureUnit]{
				Getter: getters.Key[ProgramStructureUnit]("unit"),
				Attr:   getters.FormBubbling(getters.Static("programs.StructureUnitEditModal")),

				Title: "Edit structure unit",
				ChildrenInput: []components.PageInterface{
					&components.InputText{
						Hidden: true,
						Name:   "ProgramID",
						Getter: getters.Format("%d", getters.Any(getters.Key[uint]("program.ID"))),
					},
					structureUnitFormFieldsEdit(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save changes", Classes: "btn-primary"},
									&components.ButtonModalForm{
										Page:  components.Page{Roles: []string{"admin", "superuser"}},
										Label: "Remove unit",
										Icon:  "trash",
										Name:  getters.Static("programs.StructureUnitDeleteForm"),
										Url: lago.RoutePath("programs.StructureUnitDeleteRoute", map[string]getters.Getter[any]{
											"id":     getters.Any(getters.Key[uint]("program.ID")),
											"unitId": getters.Any(getters.Key[uint]("unit.ID")),
										}),
										FormPostURL: lago.RoutePath("programs.StructureUnitDeleteRoute", map[string]getters.Getter[any]{
											"id":     getters.Any(getters.Key[uint]("program.ID")),
											"unitId": getters.Any(getters.Key[uint]("unit.ID")),
										}),
										ModalUID: "structure-unit-delete-modal",
										Classes:  "btn-error",
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
