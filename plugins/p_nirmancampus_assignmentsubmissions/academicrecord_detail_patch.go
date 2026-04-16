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
	"github.com/lariv-in/lago/registry"
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

func academicRecordDetailBulkCreateSubmissionsButton() *components.ButtonModalForm {
	return &components.ButtonModalForm{
		Page:  components.Page{Roles: []string{"admin", "superuser"}},
		Label: "Create Submissions for Student",
		Name:  getters.Static("assignmentsubmissions.BulkCreateFromAcademicRecordForm"),
		Url: getters.Format(
			"%s?AcademicRecordID=%d",
			getters.Any(lago.RoutePath("assignmentsubmissions.BulkCreateFromAcademicRecordRoute", nil)),
			getters.Any(getters.Key[uint]("academicrecord.ID")),
		),
		FormPostURL: getters.Format(
			"%s?AcademicRecordID=%d",
			getters.Any(lago.RoutePath("assignmentsubmissions.BulkCreateFromAcademicRecordRoute", nil)),
			getters.Any(getters.Key[uint]("academicrecord.ID")),
		),
		ModalUID: "assignmentsubmissions-bulk-create-academic-record-modal",
		Classes:  "btn-outline btn-sm",
		Attr:     getters.ModalRefreshList(getters.Static(""), getters.Static("#academic-record-assignment-submissions-table")),
	}
}

func academicRecordDetailAssignmentSubmissionsSection() components.PageInterface {
	return &components.ContainerColumn{
		Page:    components.Page{Key: "assignmentsubmissions.AcademicRecordDetailSubmissionsSection"},
		Classes: "w-full mt-4 flex flex-col gap-2",
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "flex justify-end w-full",
				Children: []components.PageInterface{
					academicRecordDetailBulkCreateSubmissionsButton(),
				},
			},
			&components.DataTable[AssignmentSubmission]{
				Page:        components.Page{Key: "assignmentsubmissions.AcademicRecordDetailTable"},
				UID:         "academic-record-assignment-submissions-table",
				Title:       "Assignment submissions",
				Classes:     "w-full",
				Data:        getters.Key[components.ObjectList[AssignmentSubmission]](academicRecordDetailSubmissionsContextKey),
				DefaultView: "Grid",
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
							&components.FieldText{
								Getter: registry.PairValueFromKey(
									getters.Key[string]("$row.SubmissionStatus"),
									AssignmentSubmissionStatusChoices,
								),
							},
						},
					},
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
