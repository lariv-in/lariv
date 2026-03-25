package p_nirmancampus_programs

import (
	"context"
	"errors"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func universityChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "IGNOU", Value: "IGNOU"},
		{Key: "MRSPTU", Value: "MRSPTU"},
	}
}

func programUniversityStringGetter(programIDGetter getters.Getter[uint]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		programID, err := programIDGetter(ctx)
		if err != nil || programID == 0 {
			return "", nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return "", errors.New("Couldn't load db connection from context")
		}

		var details NirmancampusProgramDetails
		err = db.Where("program_id = ?", programID).Take(&details).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}

		return details.University, nil
	}
}

func programUniversityPairGetter(programIDGetter getters.Getter[uint]) getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := programUniversityStringGetter(programIDGetter)(ctx)
		if err != nil {
			return registry.Pair[string, string]{}, err
		}
		if s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func universityFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$get.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func universitySelectInput(programIDGetter getters.Getter[uint]) *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "University",
		Name:     "University",
		Required: false,
		Choices:  getters.GetterStatic(universityChoices()),
		Getter:   programUniversityPairGetter(programIDGetter),
	}
}

func patchProgramFormPages() {
	const bodyKey = "programs.ProgramFormFieldsBody"

	patchForm := func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for programs program form was not ShellScaffold")
		}

		programID := getters.GetterKey[uint]("$in.ID")

		components.ReplaceChild(scaffold, bodyKey, func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children,
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.University"),
					Children: []components.PageInterface{
						universitySelectInput(programID),
					},
				},
			)
			return column
		})

		return scaffold
	}

	lago.RegistryPage.Patch("programs.ProgramCreateForm", patchForm)
	lago.RegistryPage.Patch("programs.ProgramUpdateForm", patchForm)
}

func patchProgramTable() {
	const tableKey = "programs.ProgramTableBody"

	lago.RegistryPage.Patch("programs.ProgramTable", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for programs.ProgramTable was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, tableKey, func(table *components.DataTable[p_programs.Program]) *components.DataTable[p_programs.Program] {
			rowProgramID := getters.GetterKey[uint]("$row.ID")

			table.Columns = append(table.Columns, components.TableColumn{
				Label: "University",
				Name:  "University",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: programUniversityStringGetter(rowProgramID),
					},
				},
			})

			return table
		})

		return scaffold
	})
}

func patchProgramDetail() {
	const detailKey = "programs.ProgramDetailContent"

	lago.RegistryPage.Patch("programs.ProgramDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for programs.ProgramDetail was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, detailKey, func(column components.ContainerColumn) components.ContainerColumn {
			programID := getters.GetterKey[uint]("$in.ID")

			column.Children = append(column.Children,
				&components.LabelInline{
					Title: "University",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: programUniversityStringGetter(programID),
						},
					},
				},
			)

			return column
		})

		return scaffold
	})
}

func universityFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "University",
		Name:    "University",
		Choices: getters.GetterStatic(universityChoices()),
		Getter:  universityFilterPairGetter(),
	}
}

func patchProgramFilters() {
	lago.RegistryPage.Patch("programs.ProgramFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_programs.Program])
		if !ok {
			panic("Base page for programs.ProgramFilter was not FormComponent[Program]")
		}
		form.ChildrenInput = append(form.ChildrenInput, universityFilterSelect())
		return form
	})

	lago.RegistryPage.Patch("programs.ProgramSelectionFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_programs.Program])
		if !ok {
			panic("Base page for programs.ProgramSelectionFilter was not FormComponent[Program]")
		}
		form.ChildrenInput = append(form.ChildrenInput, universityFilterSelect())
		return form
	})
}

func patchProgramSelectionTable() {
	const selectionTableKey = "programs.ProgramSelectionTableBody"

	lago.RegistryPage.Patch("programs.ProgramSelectionTable", func(page components.PageInterface) components.PageInterface {
		modal, ok := page.(*components.Modal)
		if !ok {
			panic("Base page for programs.ProgramSelectionTable was not Modal")
		}

		components.ReplaceChild(modal, selectionTableKey, func(table *components.DataTable[p_programs.Program]) *components.DataTable[p_programs.Program] {
			rowProgramID := getters.GetterKey[uint]("$row.ID")

			table.Columns = append(table.Columns, components.TableColumn{
				Label: "University",
				Name:  "University",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: programUniversityStringGetter(rowProgramID),
					},
				},
			})

			return table
		})

		return modal
	})
}

func init() {
	patchProgramFormPages()
	patchProgramTable()
	patchProgramDetail()
	patchProgramFilters()
	patchProgramSelectionTable()
}
