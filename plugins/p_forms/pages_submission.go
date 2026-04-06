package forms

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerSubmissionPages() {
	lago.RegistryPage.Register("forms.SubmissionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[FormSubmission]{
				Page:    components.Page{Key: "forms.SubmissionTableBody"},
				UID:     "form-submissions-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[FormSubmission]]("form_submissions"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("forms.SubmissionDetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.Any(getters.Key[uint]("$row.FormID")),
						"id":      getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Name:  "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ID")))},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.SubmissionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[FormSubmission]{
				Getter: getters.Key[FormSubmission]("form_submission"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.LabelInline{
								Title: "Submitted at",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%v", getters.Any(getters.Key[any]("$in.CreatedAt")))},
								},
							},
							&SubmissionAnswersDisplay{
								Page: components.Page{Key: "forms.SubmissionAnswersDisplay"},
							},
						},
					},
				},
			},
		},
	})
}
