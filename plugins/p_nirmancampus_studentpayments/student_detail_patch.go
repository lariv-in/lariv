package p_nirmancampus_studentpayments

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

const studentDetailPaymentsContextKey = "student_payments_table"

func init() {
	registerStudentDetailPaymentsPatch()
}

type studentPaymentsContextLayer struct{}

func (studentPaymentsContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		student, ok := r.Context().Value("student").(p_nirmancampus_students.Student)
		if !ok || student.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("studentPaymentsContextLayer: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		var rows []Payment
		if err := db.Model(&Payment{}).
			Preload("Student").
			Where("student_id = ?", student.ID).
			Order(`"paid_at" DESC NULLS LAST`).
			Order("id DESC").
			Find(&rows).Error; err != nil {
			slog.Error("studentPaymentsContextLayer: query failed", "error", err)
			next.ServeHTTP(w, r)
			return
		}

		ol := components.ObjectList[Payment]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(rows)),
		}
		ctx := context.WithValue(r.Context(), studentDetailPaymentsContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func studentDetailPaymentColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Amount",
			Name:  "Amount",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: getters.Format("%.2f", getters.Any(getters.Key[float64]("$row.Amount"))),
				},
			},
		},
		{
			Label: "Method",
			Name:  "PaymentMethod",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: registry.PairValueFromKey(getters.Key[string]("$row.PaymentMethod"), PaymentMethodChoices),
				},
			},
		},
		{
			Label: "Paid on",
			Name:  "PaidAt",
			Children: []components.PageInterface{
				&components.FieldText{Getter: paidAtDisplay(getters.Key[*time.Time]("$row.PaidAt"))},
			},
		},
		{
			Label: "Transaction ID",
			Name:  "TransactionID",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.TransactionID")},
			},
		},
	}
}

func studentDetailPaymentsSection() components.PageInterface {
	createFormName := getters.Static("studentpayments.PaymentCreateForm")
	return &components.DataTable[Payment]{
		Page:        components.Page{Key: "studentpayments.StudentDetailPaymentsTable"},
		UID:         "student-detail-payments-table",
		Title:       "Payments",
		Classes:     "w-full mt-4",
		Data:        getters.Key[components.ObjectList[Payment]](studentDetailPaymentsContextKey),
		DefaultView: "Grid",
		Actions: []components.PageInterface{
			&components.ButtonModalForm{
				Page:        components.Page{Roles: []string{"admin", "superuser"}},
				Name:        createFormName,
				Url: getters.Format(
					"%s?StudentID=%d",
					getters.Any(lago.RoutePath("studentpayments.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("student.ID")),
				),
				FormPostURL: getters.Format(
					"%s?StudentID=%d",
					getters.Any(lago.RoutePath("studentpayments.CreateRoute", nil)),
					getters.Any(getters.Key[uint]("student.ID")),
				),
				ModalUID: "studentpayments-create-modal",
				Icon:     "plus",
				Classes:  "btn-square btn-outline btn-sm",
				Attr:     getters.ModalRefreshList(getters.Static(""), getters.Static("#student-detail-payments-table")),
			},
		},
		RowAttr: getters.RowAttrNavigate(lago.RoutePath("studentpayments.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("$row.ID")),
		})),
		Columns: studentDetailPaymentColumns(),
	}
}

func registerStudentDetailPaymentsPatch() {
	lago.RegistryView.Patch("students.DetailView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("students.detail", "studentpayments.student_detail", studentPaymentsContextLayer{})
	})

	lago.RegistryPage.Patch("students.StudentDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("students.StudentDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "students.StudentDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, studentDetailPaymentsSection())
			return column
		})
		return scaffold
	})
}
