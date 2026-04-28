package p_nirmancampus_examregistrations

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

const academicRecordDetailExamRegistrationsContextKey = "academic_record_exam_registrations_table"

func init() {
	registerAcademicRecordDetailPatch()
}

type academicRecordExamRegistrationsContextLayer struct{}

func (academicRecordExamRegistrationsContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, ok := r.Context().Value("academicrecord").(p_nirmancampus_academicrecords.AcademicRecord)
		if !ok || record.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("academicRecordExamRegistrationsContextLayer: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var rows []ExamRegistration
		if err := db.Model(&ExamRegistration{}).
			Preload("Course").
			Where("academic_record_id = ?", record.ID).
			Order("created_at DESC").
			Order("id DESC").
			Find(&rows).Error; err != nil {
			slog.Error("academicRecordExamRegistrationsContextLayer: query failed", "error", err)
			next.ServeHTTP(w, r)
			return
		}

		ol := components.ObjectList[ExamRegistration]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(rows)),
		}
		ctx := context.WithValue(r.Context(), academicRecordDetailExamRegistrationsContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func academicRecordDetailBulkCreateExamRegistrationsButton() *components.ButtonModalForm {
	return &components.ButtonModalForm{
		Page:  components.Page{Roles: []string{"admin", "superuser"}},
		Label: "Create Exam Registrations for Student",
		Name:  getters.Static("examregistrations.BulkCreateFromAcademicRecordForm"),
		Url: getters.Format(
			"%s?AcademicRecordID=%d",
			getters.Any(lago.RoutePath("examregistrations.BulkCreateFromAcademicRecordRoute", nil)),
			getters.Any(getters.Key[uint]("academicrecord.ID")),
		),
		FormPostURL: getters.Format(
			"%s?AcademicRecordID=%d",
			getters.Any(lago.RoutePath("examregistrations.BulkCreateFromAcademicRecordRoute", nil)),
			getters.Any(getters.Key[uint]("academicrecord.ID")),
		),
		ModalUID: "examregistrations-bulk-create-academic-record-modal",
		Classes:  "btn-outline btn-sm",
		Attr:     getters.ModalRefreshList(getters.Static(""), getters.Static("#academic-record-exam-registrations-table")),
	}
}

func academicRecordDetailExamRegistrationsSection() components.PageInterface {
	return &components.ContainerColumn{
		Page:    components.Page{Key: "examregistrations.AcademicRecordDetailExamRegistrationsSection"},
		Classes: "mt-4 flex flex-col gap-2",
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "flex flex-wrap gap-2 items-center",
				Children: []components.PageInterface{
					academicRecordDetailBulkCreateExamRegistrationsButton(),
					&components.ButtonDownload{
						Label: "Download Receipt",
						Link: lago.RoutePath("examregistrations.AcademicRecordExamReceiptRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
						}),
						Classes: "btn-outline btn-secondary btn-sm",
					},
				},
			},
			&components.DataTable[ExamRegistration]{
				Page:        components.Page{Key: "examregistrations.AcademicRecordDetailTable"},
				UID:         "academic-record-exam-registrations-table",
				Title:       "Exam Registrations",
				Classes:     "w-full",
				Data:        getters.Key[components.ObjectList[ExamRegistration]](academicRecordDetailExamRegistrationsContextKey),
				DefaultView: "Grid",
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("examregistrations.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Exam",
						Name:  "ExamTitle",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ExamTitle")},
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
						Label: "Fee",
						Name:  "Fee",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("₹ %d", getters.Any(getters.Key[uint]("$row.Fee")))},
						},
					},
					{
						Label: "Status",
						Name:  "RegistrationStatus",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: registry.PairValueFromKey(
									getters.Key[string]("$row.RegistrationStatus"),
									ExamRegistrationStatusChoices,
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
		return v.InsertLayerAfter("academicrecords.detail", "examregistrations.academic_record_detail", academicRecordExamRegistrationsContextLayer{})
	})

	lago.RegistryPage.Patch("academicrecords.AcademicRecordDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("academicrecords.AcademicRecordDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "academicrecords.AcademicRecordDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, academicRecordDetailExamRegistrationsSection())
			return column
		})
		return scaffold
	})
}
