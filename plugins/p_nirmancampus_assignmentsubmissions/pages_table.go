package p_nirmancampus_assignmentsubmissions

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
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
				Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
					s, err := getters.Key[string]("$get.SubmissionStatus")(ctx)
					if err != nil || s == "" {
						return registry.Pair[string, string]{}, nil
					}
					if p, ok := registry.PairFromPairs(s, AssignmentSubmissionStatusChoices); ok {
						return p, nil
					}
					return registry.Pair[string, string]{Key: s, Value: s}, nil
				},
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
	lago.RegistryPage.Register("assignmentsubmissions.Table", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AssignmentSubmission]{
				Page:    components.Page{Key: "assignmentsubmissions.TableBody"},
				UID:     "assignment-submissions-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AssignmentSubmission]]("assignmentsubmissions"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "assignmentsubmissions.Filter"}},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Name:        getters.Static("assignmentsubmissions.CreateForm"),
						Url:         lago.RoutePath("assignmentsubmissions.CreateRoute", nil),
						FormPostURL: lago.RoutePath("assignmentsubmissions.CreateRoute", nil),
						ModalUID:    "assignmentsubmissions-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
					},
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
