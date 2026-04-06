package forms

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFormListPages() {
	lago.RegistryPage.Register("forms.FormTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Form]{
				Page:    components.Page{Key: "forms.FormTableBody"},
				UID:     "forms-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Form]]("forms"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{
						Link: lago.RoutePath("forms.CreateRoute", nil),
					},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Slug",
						Name:  "Slug",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Slug")},
						},
					},
				},
			},
		},
	})
}
