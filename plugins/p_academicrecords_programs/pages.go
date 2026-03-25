package p_academicrecords_programs

import (
	"context"
	"errors"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_academicrecords"
	"github.com/lariv-in/lago/plugins/p_programs"
	"gorm.io/gorm"
)

func academicRecordProgramGetter(recordIDGetter getters.Getter[uint]) getters.Getter[p_programs.Program] {
	return func(ctx context.Context) (p_programs.Program, error) {
		recordID, err := recordIDGetter(ctx)
		if err != nil || recordID == 0 {
			return p_programs.Program{}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return p_programs.Program{}, errors.New("Couldn't load db connection from context")
		}

		var details AcademicRecordProgramDetails
		err = db.Preload("Program").
			Where("academic_record_id = ?", recordID).
			Take(&details).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return p_programs.Program{}, nil
			}
			return p_programs.Program{}, err
		}

		return details.Program, nil
	}
}

func academicRecordProgramNameGetter(recordIDGetter getters.Getter[uint]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		programGetter := academicRecordProgramGetter(recordIDGetter)
		p, err := programGetter(ctx)
		if err != nil {
			return "", err
		}
		return p.Name, nil
	}
}

func patchAcademicRecordFormPages() {
	const bodyKey = "academicrecords.AcademicRecordFormFieldsBody"

	patchCreate := func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordCreateForm was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, bodyKey, func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children,
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.ProgramID"),
					Children: []components.PageInterface{
						&components.InputForeignKey[p_programs.Program]{
							Label:       "Program",
							Name:        "ProgramID",
							Required:    true,
							Url:         lago.GetterRoutePath("programs.SelectRoute", nil),
							Display:     getters.GetterKey[string]("$in.Name"),
							Placeholder: "Select a program...",
							Getter: academicRecordProgramGetter(
								getters.GetterKey[uint]("$in.ID"),
							),
						},
					},
				},
			)
			return column
		})

		return scaffold
	}

	lago.RegistryPage.Patch("academicrecords.AcademicRecordCreateForm", patchCreate)

	lago.RegistryPage.Patch("academicrecords.AcademicRecordUpdateForm", func(page components.PageInterface) components.PageInterface {
		// Same scaffold shape as create; $in differs (update has $in.academicrecord).
		return patchCreate(page)
	})
}

func patchAcademicRecordTable() {
	const tableKey = "academicrecords.AcademicRecordTableBody"

	lago.RegistryPage.Patch("academicrecords.AcademicRecordTable", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordTable was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, tableKey, func(table *components.DataTable[p_academicrecords.AcademicRecord]) *components.DataTable[p_academicrecords.AcademicRecord] {
			rowID := getters.GetterKey[uint]("$row.ID")
			table.Columns = append(table.Columns,
				components.TableColumn{
					Label: "Program",
					Name:  "Program",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: academicRecordProgramNameGetter(rowID),
						},
					},
				},
			)
			return table
		})

		return scaffold
	})
}

func patchAcademicRecordDetail() {
	const detailKey = "academicrecords.AcademicRecordDetailContent"

	lago.RegistryPage.Patch("academicrecords.AcademicRecordDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordDetail was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, detailKey, func(column components.ContainerColumn) components.ContainerColumn {
			recordID := getters.GetterKey[uint]("$in.ID")
			column.Children = append(column.Children,
				&components.LabelInline{
					Title: "Program",
					Children: []components.PageInterface{
						&components.FieldText{Getter: academicRecordProgramNameGetter(recordID)},
					},
				},
			)
			return column
		})

		return scaffold
	})
}

func init() {
	patchAcademicRecordFormPages()
	patchAcademicRecordTable()
	patchAcademicRecordDetail()
}
