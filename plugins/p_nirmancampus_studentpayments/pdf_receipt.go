package p_nirmancampus_studentpayments

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alnah/go-md2pdf"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const pdfReceiptCSS = `
@import url('https://fonts.googleapis.com/css2?family=Noto+Serif:wght@400;700&family=Noto+Serif+Devanagari:wght@400;700&display=swap');
html, body {
	font-family: "Noto Serif Devanagari", "Noto Serif", serif;
	font-size: 11pt;
	line-height: 1.35;
	color: #111;
}
h1 {
	font-size: 15pt;
	font-weight: 700;
	margin: 0 0 8pt;
	border-bottom: 1px solid #ccc;
	padding-bottom: 4pt;
}
h2 {
	font-size: 11pt;
	font-weight: 700;
	margin: 12pt 0 6pt;
	color: #222;
}
table {
	border-collapse: collapse;
	width: 100%;
	margin: 6pt 0 10pt;
}
td, th {
	border: 1px solid #bbb;
	padding: 5pt 8pt;
	font-size: 10pt;
	vertical-align: top;
}
thead th {
	background: #f4f4f4;
	font-weight: 600;
	width: 28%;
}
tbody td:last-child { width: 72%; }
hr { border: none; border-top: 1px solid #ddd; margin: 12pt 0; }
.footer { font-size: 9pt; color: #555; margin-top: 16pt; }
`

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	return s
}

func issuedDateLocal(r *http.Request) string {
	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.Local
	}
	return time.Now().In(tz).Format("January 2, 2006")
}

func paymentReceiptMarkdown(r *http.Request, p *Payment) string {
	method := p.PaymentMethod
	if pair, ok := registry.PairFromPairs(p.PaymentMethod, PaymentMethodChoices); ok {
		method = pair.Value
	}
	paidOn := "—"
	if p.PaidAt != nil && !p.PaidAt.IsZero() {
		paidOn = p.PaidAt.Format("January 2, 2006")
	}
	remarks := strings.TrimSpace(p.Remarks)
	if remarks == "" {
		remarks = "—"
	} else {
		remarks = mdCell(remarks)
	}
	st := &p.Student
	return fmt.Sprintf(`# Payment receipt

**Issued:** %s

## Student

| Field | Value |
|---|---|
| Name | %s |
| Enrolment no. | %s |
| Email | %s |
| Phone | %s |

## Payment

| Field | Value |
|---|---|
| Receipt no. | %d |
| Amount | ₹ %.2f |
| Method | %s |
| Transaction ID | %s |
| Paid on | %s |
| Remarks | %s |

---

This certifies payment of **₹ %.2f** recorded as receipt **#%d**, received from the student named above.

<div class="footer">Reference: payment id %d · generated electronically</div>
`,
		mdCell(issuedDateLocal(r)),
		mdCell(st.Name),
		mdCell(st.StudentNo),
		mdCell(st.Email),
		mdCell(st.Phone),
		p.ID,
		p.Amount,
		mdCell(method),
		mdCell(p.TransactionID),
		mdCell(paidOn),
		remarks,
		p.Amount,
		p.ID,
		p.ID,
	)
}

func sanitizePDFBaseName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "document"
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "document"
	}
	return out
}

func downloadReceiptHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || id == 0 {
			http.NotFound(w, r)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("downloadReceiptHandler: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		q := views.QueryPatchers[Payment]{
			{Key: "studentpayments.preload", Value: views.QueryPatcherPreload[Payment]{Fields: []string{"Student"}}},
			{Key: "studentpayments.scope_by_role", Value: PaymentScopeByRole},
		}.Apply(views.View{}, r, gorm.G[Payment](db).Scopes())

		payment, err := q.Where("ID = ?", uint(id)).First(r.Context())
		if err != nil {
			slog.Error("downloadReceiptHandler: load payment", "error", err, "id", id)
			http.NotFound(w, r)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			slog.Error("downloadReceiptHandler: PDF converter", "error", err)
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		md := paymentReceiptMarkdown(r, &payment)
		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: md,
			CSS:      pdfReceiptCSS,
		})
		if err != nil {
			slog.Error("downloadReceiptHandler: convert", "error", err, "paymentID", payment.ID)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		base := fmt.Sprintf("payment-receipt-%d", payment.ID)
		if payment.Student.StudentNo != "" {
			base = sanitizePDFBaseName(payment.Student.StudentNo) + "-receipt-" + strconv.FormatUint(uint64(payment.ID), 10)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, base))
		if _, err := w.Write(result.PDF); err != nil {
			slog.Error("downloadReceiptHandler: write response", "error", err)
		}
	})
}
