package p_nirmancampus_academicrecords

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/views"
)

const studentDetailAcademicRecordsContextKey = "student_academic_records_table"

func init() {
	registerStudentDetailAcademicRecordsPatch()
}

// attachStudentAcademicRecordsContext loads AcademicRecords for the current
// student (from the "student" context key set by DetailView) and stores
// them as an ObjectList under studentDetailAcademicRecordsContextKey.
type studentAcademicRecordsContextLayer struct{}

func (studentAcademicRecordsContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		student, ok := r.Context().Value("student").(p_nirmancampus_students.Student)
		if !ok || student.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("attachStudentAcademicRecordsContext: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var rows []AcademicRecord
		if err := db.Model(&AcademicRecord{}).
			Preload("Program").
			Preload("Session").
			Preload("ProgramStructureUnit").
			Preload("CompulsoryCourses").
			Preload("OptionalCourses").
			Where("student_id = ?", student.ID).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			slog.Error("attachStudentAcademicRecordsContext: query failed", "error", err)
			next.ServeHTTP(w, r)
			return
		}

		ol := components.ObjectList[AcademicRecord]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(rows)),
		}
		ctx := context.WithValue(r.Context(), studentDetailAcademicRecordsContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func studentDetailAcademicRecordsSection() components.PageInterface {
	return &components.DataTable[AcademicRecord]{
		Page:        components.Page{Key: "academicrecords.StudentDetailAcademicRecordsTable"},
		UID:         "student-detail-academic-records-table",
		Title:       "Academic records",
		Classes:     "w-full mt-4",
		Data:        getters.Key[components.ObjectList[AcademicRecord]](studentDetailAcademicRecordsContextKey),
		DefaultView: "Grid",
		Actions: []components.PageInterface{
			&components.ButtonModalForm{
				Name: getters.Static("academicrecords.AcademicRecordCreateForm"),
				Url: getters.Format(
					"%s?StudentID=%d",
					getters.Any(lago.RoutePath("academicrecords.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("student.ID")),
				),
				FormPostURL: getters.Format(
					"%s?StudentID=%d",
					getters.Any(lago.RoutePath("academicrecords.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("student.ID")),
				),
				ModalUID: "academicrecords-create-modal",
				Icon:     "plus",
				Classes:  "btn-square btn-outline btn-sm",
			},
		},
		RowAttr: getters.RowAttrNavigate(lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("$row.ID")),
		})),
		Columns: []components.TableColumn{
			{
				Label: "Program",
				Name:  "Program.Name",
				Children: []components.PageInterface{
					&components.FieldText{Getter: p_nirmancampus_programs.ProgramDisplayLabel(
						getters.Key[string]("$row.Program.Name"),
						getters.Key[string]("$row.Program.University"),
					)},
				},
			},
			{
				Label: "Session",
				Name:  "Session.Name",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.Session.Name")},
				},
			},
			{
				Label: "Status",
				Name:  "Status",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.Status")},
				},
			},
			{
				Label: "Term",
				Name:  "Term",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ProgramStructureUnit.TermNumber"))),
					},
				},
			},
		},
	}
}

func registerStudentDetailAcademicRecordsPatch() {
	lago.RegistryView.Patch("students.DetailView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("students.detail", "academicrecords.student_detail", studentAcademicRecordsContextLayer{})
	})

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
