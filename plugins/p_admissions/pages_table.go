package p_admissions

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("admissions.ApplicationTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "admissions.ApplicationMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[AdmissionApplication]{
				Page:    components.Page{Key: "admissions.ApplicationTableBody"},
				UID:     "application-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AdmissionApplication]]("applications"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("admissions.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("admissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Program ID", Name: "ProgramID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ProgramID")))}}},
					{Label: "Semester ID", Name: "SemesterID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.SemesterID")))}}},
					{Label: "Applicant", Name: "ApplicantName", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.ApplicantName")}}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Email")}}},
					{Label: "Status", Name: "Status", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Status")}}},
				},
			},
		},
	})
}
