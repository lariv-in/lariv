package p_nirmancampus_academicrecords

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/registry"
)

func tableColumns() []components.TableColumn {
	return []components.TableColumn{
		{Label: "Student", Name: "Student.User.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Student.User.Name")},
		}},
		{Label: "Program", Name: "Program.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Program.Name")},
		}},
		{Label: "Session", Name: "Session.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Session.Name")},
		}},
		{Label: "Status", Name: "Status", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Status")},
		}},
		{Label: "Term", Name: "Term", Children: []components.PageInterface{
			&components.FieldText{
				Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.Term"))),
			},
		}},
	}
}

// --- Form Field Getters ---

// programStructureUnitForIn loads the ProgramStructureUnit for $in.ProgramID
// and $in.Term. When preloadOptionalPool is true, OptionalCourseSelectionPool
// is preloaded (for multi-select URLs).
func registerFilterPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFilter", &components.FormComponent[AcademicRecord]{
		Attr: getters.FormBoostedGet(lago.RoutePath("academicrecords.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputSelect[string]{
				Label:   "Status",
				Name:    "Status",
				Choices: getters.Static(registry.PairsFromMap(AcademicRecordStatusChoices)),
				Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
					s, err := getters.Key[string]("$get.Status")(ctx)
					if err != nil || s == "" {
						return registry.Pair[string, string]{}, nil
					}
					if p, ok := registry.PairFromMap(s, AcademicRecordStatusChoices); ok {
						return p, nil
					}
					return registry.Pair[string, string]{Key: s, Value: s}, nil
				},
			},
			&components.InputText{
				Label:  "Term",
				Name:   "Term",
				Getter: getters.Key[string]("$get.Term"),
			},
			&components.InputForeignKey[p_nirmancampus_programs.Program]{
				Label:       "Program",
				Name:        "ProgramID",
				Url:         lago.RoutePath("programs.SelectRoute", nil),
				Placeholder: "Filter by program...",
				Display:     getters.Key[string]("$in.Name"),
				Getter: getters.Association[p_nirmancampus_programs.Program](
					getters.Key[uint]("$get.ProgramID"),
				),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

// --- Form Fields ---

func registerTablePages() {
	academicRecordsSessionEnvironment := &components.Environment[uint]{
		Label:   "Session",
		Key:     getters.Static(academicRecordsEnvironmentSessionKey),
		Options: AcademicSessionsListGetter,
		Default: academicRecordsSessionEnvironmentDefault,
	}
	lago.RegistryPage.Register("academicrecords.AcademicRecordTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			academicRecordsSessionEnvironment,
			&components.DataTable[AcademicRecord]{
				Page:    components.Page{Key: "academicrecords.AcademicRecordTableBody"},
				UID:     "academicrecords-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AcademicRecord]]("academicrecords"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "academicrecords.AcademicRecordFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
					&components.ButtonModalForm{
						Page:        components.Page{Roles: []string{"admin", "superuser"}},
						Url:         lago.RoutePath("academicrecords.CreateRoute", nil),
						FormPostURL: lago.RoutePath("academicrecords.CreateRoute", nil),
						ModalUID:    "academicrecords-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
					},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: tableColumns(),
			},
		},
	})
}

// --- Detail & Delete ---

func registerSelectionPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordSelectionTable", &components.Modal{
		UID: "academicrecords-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page:  components.Page{Key: "academicrecords.AcademicRecordSelectionTableBody"},
				UID:   "academicrecords-selection-table",
				Title: "Select Academic Record",
				Data:  getters.Key[components.ObjectList[AcademicRecord]]("academicrecords"),
				RowAttr: getters.RowAttrSelect("AcademicRecordID", getters.Key[uint]("$row.ID"), getters.Format(
					"%s (%s) · term %s",
					getters.Any(getters.Key[string]("$row.Program.Name")),
					getters.Any(getters.Key[string]("$row.Status")),
					getters.Any(getters.Format("%d", getters.Any(getters.Key[uint]("$row.Term")))),
				)),
				Columns: tableColumns(),
			},
		},
	})
}
