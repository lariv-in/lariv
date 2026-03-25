package p_nirmancampus_students

import (
	"context"
	"errors"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_students"
	"gorm.io/gorm"
)

func studentDetailsFieldGetter(field string, studentIDGetter getters.Getter[uint]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		studentID, err := studentIDGetter(ctx)
		// For "create" forms or empty selections, $in/$row may not exist yet.
		// Treat any resolution failure as "no value" (empty string).
		if err != nil || studentID == 0 {
			return "", nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return "", errors.New("Couldn't load db connection from context")
		}

		var details NirmancampusStudentDetails
		err = db.Where("student_id = ?", studentID).Take(&details).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil
			}
			return "", err
		}

		switch field {
		case "FathersName":
			return details.FathersName, nil
		case "Category":
			return details.Category, nil
		case "Address":
			return details.Address, nil
		default:
			return "", errors.New("unknown student detail field: " + field)
		}
	}
}

func patchStudentFormPages() {
	const bodyKey = "students.StudentFormFieldsBody"

	patchCreate := func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for students.StudentCreateForm was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, bodyKey, func(column components.ContainerColumn) components.ContainerColumn {
			studentID := getters.GetterKey[uint]("$in.ID")

			column.Children = append(column.Children,
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.FathersName"),
					Children: []components.PageInterface{
						&components.InputText{
							Label: "Father's Name",
							Name:  "FathersName",
							Getter: studentDetailsFieldGetter(
								"FathersName",
								studentID,
							),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.Category"),
					Children: []components.PageInterface{
						&components.InputText{
							Label:  "Category",
							Name:   "Category",
							Getter: studentDetailsFieldGetter("Category", studentID),
						},
					},
				},
				&components.ContainerError{
					Error: getters.GetterKey[error]("$error.Address"),
					Children: []components.PageInterface{
						&components.InputTextarea{
							Label:  "Address",
							Name:   "Address",
							Rows:   3,
							Getter: studentDetailsFieldGetter("Address", studentID),
						},
					},
				},
			)

			return column
		})

		return scaffold
	}

	lago.RegistryPage.Patch("students.StudentCreateForm", patchCreate)

	lago.RegistryPage.Patch("students.StudentUpdateForm", func(page components.PageInterface) components.PageInterface {
		// Same page shape as create, only $in differs (update has getters.GetterKey[Student]("student")).
		return patchCreate(page)
	})
}

func patchStudentTable() {
	const tableKey = "students.StudentTableBody"

	lago.RegistryPage.Patch("students.StudentTable", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for students.StudentTable was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, tableKey, func(table *components.DataTable[p_students.Student]) *components.DataTable[p_students.Student] {
			rowStudentID := getters.GetterKey[uint]("$row.ID")

			table.Columns = append(table.Columns,
				components.TableColumn{
					Label: "Father's Name",
					Name:  "FathersName",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("FathersName", rowStudentID),
						},
					},
				},
				components.TableColumn{
					Label: "Category",
					Name:  "Category",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("Category", rowStudentID),
						},
					},
				},
				components.TableColumn{
					Label: "Address",
					Name:  "Address",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("Address", rowStudentID),
						},
					},
				},
			)

			return table
		})

		return scaffold
	})
}

func patchStudentDetail() {
	const detailKey = "students.StudentDetailContent"

	lago.RegistryPage.Patch("students.StudentDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for students.StudentDetail was not ShellScaffold")
		}

		components.ReplaceChild(scaffold, detailKey, func(column components.ContainerColumn) components.ContainerColumn {
			studentID := getters.GetterKey[uint]("$in.ID")

			column.Children = append(column.Children,
				&components.LabelInline{
					Title: "Address",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("Address", studentID),
						},
					},
				},
				&components.LabelInline{
					Title: "Category",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("Category", studentID),
						},
					},
				},
				&components.LabelInline{
					Title: "Father's Name",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("FathersName", studentID),
						},
					},
				},
			)

			return column
		})

		return scaffold
	})
}

func patchStudentFilter() {
	lago.RegistryPage.Patch("students.StudentFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_students.Student])
		if !ok {
			panic("Base page for students.StudentFilter was not FormComponent[Student]")
		}

		form.ChildrenInput = append(form.ChildrenInput,
			&components.InputText{
				Label:  "Father's Name",
				Name:   "FathersName",
				Getter: getters.GetterKey[string]("$get.FathersName"),
			},
			&components.InputText{
				Label:  "Category",
				Name:   "Category",
				Getter: getters.GetterKey[string]("$get.Category"),
			},
		)

		return form
	})
}

func patchStudentSelectionTable() {
	const selectionTableKey = "students.StudentSelectionTableBody"

	lago.RegistryPage.Patch("students.StudentSelectionTable", func(page components.PageInterface) components.PageInterface {
		modal, ok := page.(*components.Modal)
		if !ok {
			panic("Base page for students.StudentSelectionTable was not Modal")
		}

		components.ReplaceChild(modal, selectionTableKey, func(table *components.DataTable[p_students.Student]) *components.DataTable[p_students.Student] {
			rowStudentID := getters.GetterKey[uint]("$row.ID")

			table.Columns = append(table.Columns,
				components.TableColumn{
					Label: "Father's Name",
					Name:  "FathersName",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("FathersName", rowStudentID),
						},
					},
				},
				components.TableColumn{
					Label: "Category",
					Name:  "Category",
					Children: []components.PageInterface{
						&components.FieldText{
							Getter: studentDetailsFieldGetter("Category", rowStudentID),
						},
					},
				},
			)

			return table
		})

		return modal
	})
}

func init() {
	patchStudentFormPages()
	patchStudentTable()
	patchStudentDetail()
	patchStudentFilter()
	patchStudentSelectionTable()
}
