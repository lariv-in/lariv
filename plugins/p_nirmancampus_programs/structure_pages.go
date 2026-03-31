package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"net/http"
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
	units, err := getters.GetterKey[[]ProgramStructureUnit]("program.ProgramStructureUnits")(ctx)
	if err != nil || len(units) == 0 {
		return P(Class("text-base-content/70"), Text("No structure units yet. Use “Add new unit” to create one."))
	}
	var nodes []Node
	for i := range units {
		u := units[i]
		editURL, err := lago.GetterRoutePath("programs.StructureUnitEditModalRoute", map[string]getters.Getter[any]{
			"id":     getters.GetterAny(getters.GetterKey[uint]("program.ID")),
			"unitId": getters.GetterAny(getters.GetterStatic[uint](u.ID)),
		})(ctx)
		if err != nil {
			editURL = "#"
		}
		pool := joinCourseCodes(u.OptionalCourseSelectionPool)
		comp := joinCourseCodes(u.CompulsoryCourses)
		nodes = append(nodes,
			Div(Class("rounded-box border border-base-300 p-4 flex flex-col gap-2 @md:flex-row @md:items-start @md:justify-between"),
				Div(Class("flex flex-col gap-1 min-w-0"),
					Div(Class("font-semibold"), Text(fmt.Sprintf("Term %d", u.TermNumber))),
					Div(Class("text-sm text-base-content/80"), Text("Compulsory: "+comp)),
					Div(Class("text-sm text-base-content/80"), Text(fmt.Sprintf("Optional count: %d", u.OptionalCourseCount))),
					Div(Class("text-sm text-base-content/80"), Text("Optional pool: "+pool)),
				),
				components.Render(&components.ButtonModal{
					Label:   "Edit",
					Url:     getters.GetterStatic(editURL),
					Classes: "btn-outline btn-sm",
				}, ctx),
			),
		)
	}
	return Div(Class("flex flex-col gap-3"), Group(nodes))
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
						Getter: getters.GetterStatic("Edit program structure"),
					},
					&components.FieldSubtitle{
						Getter: getters.GetterKey[string]("program.Name"),
					},
					programStructureUnitCards{},
					&components.ButtonModal{
						Label:   "Add new unit",
						Url:     lago.GetterRoutePath("programs.StructureUnitCreateModalRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("program.ID"))}),
						Classes: "btn-primary",
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
					Error: getters.GetterKey[error]("$error.TermNumber"),
					Children: []components.PageInterface{
						&components.InputNumber{
							Label:    "Term number",
							Name:     "TermNumber",
							Required: true,
							Getter:   getters.GetterKey[int]("$in.TermNumber"),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.CompulsoryCourses"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Compulsory courses",
							Name:        "CompulsoryCourses",
							Getter:      getters.GetterKey[[]courses.Course]("$in.CompulsoryCourses"),
							Url:         lago.GetterRoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.GetterKey[string]("$in.Name"),
							Placeholder: "Select compulsory courses…",
							Classes:     "w-full",
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.OptionalCourseCount"),
					Children: []components.PageInterface{
						&components.InputNumber{
							Label:    "Optional course count",
							Name:     "OptionalCourseCount",
							Required: false,
							Getter:   getters.GetterKey[int]("$in.OptionalCourseCount"),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.OptionalCourseSelectionPool"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Optional course pool",
							Name:        "OptionalCourseSelectionPool",
							Getter:      getters.GetterKey[[]courses.Course]("$in.OptionalCourseSelectionPool"),
							Url:         lago.GetterRoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.GetterKey[string]("$in.Name"),
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
					Error: getters.GetterKey[error]("$error.TermNumber"),
					Children: []components.PageInterface{
						&components.InputNumber{
							Label:    "Term number",
							Name:     "TermNumber",
							Required: true,
							Getter:   getters.GetterStatic(0),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.CompulsoryCourses"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Compulsory courses",
							Name:        "CompulsoryCourses",
							Getter:      getters.GetterKey[[]courses.Course]("$in.CompulsoryCourses"),
							Url:         lago.GetterRoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.GetterKey[string]("$in.Name"),
							Placeholder: "Select compulsory courses…",
							Classes:     "w-full",
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.OptionalCourseCount"),
					Children: []components.PageInterface{
						&components.InputNumber{
							Label:    "Optional course count",
							Name:     "OptionalCourseCount",
							Required: false,
							Getter:   getters.GetterStatic(0),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.OptionalCourseSelectionPool"),
					Children: []components.PageInterface{
						&components.InputManyToMany[courses.Course]{
							Label:       "Optional course pool",
							Name:        "OptionalCourseSelectionPool",
							Getter:      getters.GetterKey[[]courses.Course]("$in.OptionalCourseSelectionPool"),
							Url:         lago.GetterRoutePath("courses.MultiSelectRoute", nil),
							Display:     getters.GetterKey[string]("$in.Name"),
							Placeholder: "Select optional pool courses…",
							Classes:     "w-full",
						},
					},
				},
			},
		}
	}

	lago.RegistryPage.Register("programs.StructureUnitCreateModal", components.Modal{
		UID:   "structure-unit-create-modal",
		Title: "Add structure unit",
		Children: []components.PageInterface{
			&components.FormComponent[ProgramStructureUnit]{
				Url:    lago.GetterRoutePath("programs.StructureUnitCreateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("program.ID"))}),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					&components.InputText{
						Hidden: true,
						Name:   "ProgramID",
						Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("program.ID"))),
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

	lago.RegistryPage.Register("programs.StructureUnitEditModal", components.Modal{
		UID:   "structure-unit-edit-modal",
		Title: "Edit structure unit",
		Children: []components.PageInterface{
			&components.FormComponent[ProgramStructureUnit]{
				Getter: getters.GetterKey[ProgramStructureUnit]("unit"),
				Url: lago.GetterRoutePath("programs.StructureUnitUpdateRoute", map[string]getters.Getter[any]{
					"id":     getters.GetterAny(getters.GetterKey[uint]("program.ID")),
					"unitId": getters.GetterAny(getters.GetterKey[uint]("unit.ID")),
				}),
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					&components.InputText{
						Hidden: true,
						Name:   "ProgramID",
						Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("program.ID"))),
					},
					structureUnitFormFieldsEdit(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save changes", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})
}
