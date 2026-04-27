package p_finances

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("finances.StudentChargeTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "finances.StudentChargeMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[StudentCharge]{
				Page: components.Page{Key: "finances.StudentChargeTableBody"}, UID: "student-charge-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[StudentCharge]]("student_charges"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("finances.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("finances.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Student", Name: "StudentID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.StudentID")))}}},
					{Label: "Amount (¢)", Name: "AmountCents", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int64]("$row.AmountCents")))}}},
					{Label: "Description", Name: "Description", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Description")}}},
				},
			},
		},
	})
}
