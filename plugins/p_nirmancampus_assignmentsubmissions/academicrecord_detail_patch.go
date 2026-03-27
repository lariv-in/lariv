package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"log"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	registerAcademicRecordDetailPatch()
}

func submissionsForCurrentAcademicRecordGetter() getters.Getter[components.ObjectList[AssignmentSubmission]] {
	return func(ctx context.Context) (components.ObjectList[AssignmentSubmission], error) {
		academicRecordID, err := getters.GetterKey[uint]("$in.ID")(ctx)
		if err != nil || academicRecordID == 0 {
			return components.ObjectList[AssignmentSubmission]{Number: 1, NumPages: 1}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return components.ObjectList[AssignmentSubmission]{Number: 1, NumPages: 1}, nil
		}

		var rows []AssignmentSubmission
		if err := db.Model(&AssignmentSubmission{}).
			Preload("Course").
			Where("academic_record_id = ?", academicRecordID).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			return components.ObjectList[AssignmentSubmission]{}, err
		}

		return components.ObjectList[AssignmentSubmission]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    int64(len(rows)),
		}, nil
	}
}

func academicRecordDetailAssignmentSubmissionsSection() components.PageInterface {
	return &components.DataTable[AssignmentSubmission]{
		Page:        components.Page{Key: "assignmentsubmissions.AcademicRecordDetailTable"},
		UID:         "academic-record-assignment-submissions-table",
		Title:       "Assignment submissions",
		Classes:     "w-full mt-4",
		Data:        submissionsForCurrentAcademicRecordGetter(),
		DefaultView: "Grid",
		CreateUrl: getters.GetterFormat(
			"%s?AcademicRecordID=%d",
			getters.GetterAny(lago.GetterRoutePath("assignmentsubmissions.CreateRoute", nil)),
			getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
		),
		OnClick: getters.GetterNavigateGetter(
			lago.GetterRoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
			}),
		),
		Columns: []components.TableColumn{
			{
				Label: "Assignment",
				Name:  "AssignmentTitle",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.AssignmentTitle")},
				},
			},
			{
				Label: "Course",
				Name:  "Course.Name",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Name")},
				},
			},
			{
				Label: "Status",
				Name:  "SubmissionStatus",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.SubmissionStatus")},
				},
			},
		},
	}
}

func registerAcademicRecordDetailPatch() {
	lago.RegistryPage.Patch("academicrecords.AcademicRecordDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("academicrecords.AcademicRecordDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "academicrecords.AcademicRecordDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, academicRecordDetailAssignmentSubmissionsSection())
			return column
		})
		return scaffold
	})
}
