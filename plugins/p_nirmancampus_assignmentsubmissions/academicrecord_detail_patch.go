package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/views"
)

const academicRecordDetailSubmissionsContextKey = "academic_record_submissions_table"

func init() {
	registerAcademicRecordDetailPatch()
}

// attachAcademicRecordSubmissionsContext loads AssignmentSubmissions for the
// current academic record (from the "academicrecord" context key set by
// DetailView) and stores them as an ObjectList under
// academicRecordDetailSubmissionsContextKey.
type academicRecordSubmissionsContextLayer struct{}

func (academicRecordSubmissionsContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, ok := r.Context().Value("academicrecord").(p_nirmancampus_academicrecords.AcademicRecord)
		if !ok || record.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("attachAcademicRecordSubmissionsContext: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var rows []AssignmentSubmission
		if err := db.Model(&AssignmentSubmission{}).
			Preload("Course").
			Where("academic_record_id = ?", record.ID).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			slog.Error("attachAcademicRecordSubmissionsContext: query failed", "error", err)
			next.ServeHTTP(w, r)
			return
		}

		ol := components.ObjectList[AssignmentSubmission]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(rows)),
		}
		ctx := context.WithValue(r.Context(), academicRecordDetailSubmissionsContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func academicRecordDetailAssignmentSubmissionsSection() components.PageInterface {
	return &components.DataTable[AssignmentSubmission]{
		Page:        components.Page{Key: "assignmentsubmissions.AcademicRecordDetailTable"},
		UID:         "academic-record-assignment-submissions-table",
		Title:       "Assignment submissions",
		Classes:     "w-full mt-4",
		Data:        getters.Key[components.ObjectList[AssignmentSubmission]](academicRecordDetailSubmissionsContextKey),
		DefaultView: "Grid",
		Actions: []components.PageInterface{
			&components.ButtonModalForm{
				Page: components.Page{Roles: []string{"admin", "superuser"}},
				Name: getters.Static("assignmentsubmissions.CreateForm"),
				Url: getters.Format(
					"%s?AcademicRecordID=%d",
					getters.Any(lago.RoutePath("assignmentsubmissions.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("academicrecord.ID")),
				),
				FormPostURL: getters.Format(
					"%s?AcademicRecordID=%d",
					getters.Any(lago.RoutePath("assignmentsubmissions.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("academicrecord.ID")),
				),
				ModalUID: "assignmentsubmissions-create-modal",
				Icon:     "plus",
				Classes:  "btn-square btn-outline btn-sm",
			},
		},
		RowAttr: getters.RowAttrNavigate(
			lago.RoutePath("assignmentsubmissions.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$row.ID")),
			}),
		),
		Columns: []components.TableColumn{
			{
				Label: "Assignment",
				Name:  "AssignmentTitle",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.AssignmentTitle")},
				},
			},
			{
				Label: "Course",
				Name:  "Course.Name",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.Course.Name")},
				},
			},
			{
				Label: "Status",
				Name:  "SubmissionStatus",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.SubmissionStatus")},
				},
			},
		},
	}
}

func registerAcademicRecordDetailPatch() {
	lago.RegistryView.Patch("academicrecords.DetailView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("academicrecords.detail", "assignmentsubmissions.academic_record_detail", academicRecordSubmissionsContextLayer{})
	})

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
