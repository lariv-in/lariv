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
		{Label: "Student", Name: "Student.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Student.Name")},
		}},
		{Label: "Program", Name: "Program.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: p_nirmancampus_programs.ProgramDisplayLabel(
				getters.Key[string]("$row.Program.Name"),
				getters.Key[string]("$row.Program.University"),
			)},
		}},
		{Label: "Term", Name: "Term", Children: []components.PageInterface{
			&components.FieldText{
				Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ProgramStructureUnit.TermNumber"))),
			},
		}},
		{Label: "Session", Name: "Session.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Session.Name")},
		}},
		{Label: "Status", Name: "Status", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Status")},
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
				Choices: getters.Static(AcademicRecordStatusChoices),
				Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
					s, err := getters.Key[string]("$get.Status")(ctx)
					if err != nil || s == "" {
						return registry.Pair[string, string]{}, nil
					}
					if p, ok := registry.PairFromPairs(s, AcademicRecordStatusChoices); ok {
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
				Display: p_nirmancampus_programs.ProgramDisplayLabel(
					getters.Key[string]("$in.Name"),
					getters.Key[string]("$in.University"),
				),
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
						Name:        getters.Static("academicrecords.AcademicRecordCreateForm"),
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
					getters.Any(p_nirmancampus_programs.ProgramDisplayLabel(
						getters.Key[string]("$row.Program.Name"),
						getters.Key[string]("$row.Program.University"),
					)),
					getters.Any(getters.Key[string]("$row.Status")),
					getters.Any(getters.Format("%d", getters.Any(getters.Key[uint]("$row.ProgramStructureUnit.TermNumber")))),
				)),
				Columns: tableColumns(),
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.ProgramStructureUnitSelectionTable", &components.Modal{
		UID: "academicrecords-program-structure-unit-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[p_nirmancampus_programs.ProgramStructureUnit]{
				Page:  components.Page{Key: "academicrecords.ProgramStructureUnitSelectionTableBody"},
				UID:   "academicrecords-program-structure-unit-selection-table",
				Title: "Select Term",
				Data:  getters.Key[components.ObjectList[p_nirmancampus_programs.ProgramStructureUnit]]("academicrecord_program_structure_units_select"),
				RowAttr: getters.RowAttrSelect(
					"ProgramStructureUnitID",
					getters.Key[uint]("$row.ID"),
					getters.Format("Term %d", getters.Any(getters.Key[uint]("$row.TermNumber"))),
				),
				Columns: []components.TableColumn{
					{
						Label: "Term",
						Name:  "TermNumber",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.TermNumber"))),
							},
						},
					},
					{
						Label: "Optional count",
						Name:  "OptionalCourseCount",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.OptionalCourseCount"))),
							},
						},
					},
				},
			},
		},
	})
}
