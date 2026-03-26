package p_nirmancampus_academicrecords

import (
	"context"
	"log"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	registerStudentDetailAcademicRecordsPatch()
}

func academicRecordsForCurrentStudentGetter() getters.Getter[components.ObjectList[AcademicRecord]] {
	return func(ctx context.Context) (components.ObjectList[AcademicRecord], error) {
		studentID, err := getters.GetterKey[uint]("$in.ID")(ctx)
		if err != nil || studentID == 0 {
			return components.ObjectList[AcademicRecord]{Number: 1, NumPages: 1}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return components.ObjectList[AcademicRecord]{Number: 1, NumPages: 1}, nil
		}

		var rows []AcademicRecord
		if err := db.Model(&AcademicRecord{}).
			Preload("Program").
			Preload("Courses").
			Where("student_id = ?", studentID).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			return components.ObjectList[AcademicRecord]{}, err
		}

		return components.ObjectList[AcademicRecord]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    int64(len(rows)),
		}, nil
	}
}

func studentDetailAcademicRecordsSection() components.PageInterface {
	return &components.DataTable[AcademicRecord]{
		Page:        components.Page{Key: "academicrecords.StudentDetailAcademicRecordsTable"},
		UID:         "student-detail-academic-records-table",
		Title:       "Academic records",
		Classes:     "w-full mt-4",
		Data:        academicRecordsForCurrentStudentGetter(),
		DefaultView: "Grid",
		CreateUrl: getters.GetterFormat(
			"%s?StudentID=%d",
			getters.GetterAny(lago.GetterRoutePath("academicrecords.CreateRoute", nil)),
			getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
		),
		OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
		})),
		Columns: []components.TableColumn{
			{
				Label: "Program",
				Name:  "Program.Name",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
				},
			},
			{
				Label: "Status",
				Name:  "Status",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Status")},
				},
			},
			{
				Label: "Semester / year",
				Name:  "SemesterOrYear",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.SemesterOrYear")},
				},
			},
		},
	}
}

func registerStudentDetailAcademicRecordsPatch() {
	lago.RegistryPage.Patch("students.StudentDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("students.StudentDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "students.StudentDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, studentDetailAcademicRecordsSection())
			return column
		})
		return scaffold
	})
}
