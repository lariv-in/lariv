package p_nirmancampus_examregistrations

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alnah/go-md2pdf"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const pdfExamReceiptCSS = `
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
	margin: 8pt 0 10pt;
	font-size: 9.5pt;
}
td, th {
	border: 1px solid #bbb;
	padding: 4pt 6pt;
	vertical-align: top;
}
thead th {
	background: #f4f4f4;
	font-weight: 600;
	text-align: left;
}
.footer { font-size: 9pt; color: #555; margin-top: 16pt; }
`

func mdTblExam(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	return s
}

func receiptIssuedExam(r *http.Request) string {
	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.Local
	}
	return time.Now().In(tz).Format("January 2, 2006")
}

func registrationStatusLabel(k string) string {
	if p, ok := registry.PairFromPairs(k, ExamRegistrationStatusChoices); ok {
		return p.Value
	}
	return k
}

func academicRecordExamPDFMarkdown(r *http.Request, ar *p_nirmancampus_academicrecords.AcademicRecord, rows []ExamRegistration) string {
	st := ar.Student
	prog := ar.Program.Name
	if u := strings.TrimSpace(ar.Program.University); u != "" {
		if p, ok := registry.PairFromPairs(u, p_nirmancampus_programs.UniversityChoices); ok {
			prog = fmt.Sprintf("%s (%s)", ar.Program.Name, p.Value)
		} else {
			prog = fmt.Sprintf("%s (%s)", ar.Program.Name, u)
		}
	}

	var table strings.Builder
	fmt.Fprintf(&table, "| Exam | Course | Status | Fee | Recorded |\n|---|---|---|---|---|\n")
	for _, s := range rows {
		course := s.Course.Name
		recAt := "—"
		if !s.CreatedAt.IsZero() {
			recAt = s.CreatedAt.Format("2006-01-02 15:04")
		}
		fee := fmt.Sprintf("₹ %d", s.Fee)
		fmt.Fprintf(&table, "| %s | %s | %s | %s | %s |\n",
			mdTblExam(s.ExamTitle),
			mdTblExam(course),
			mdTblExam(registrationStatusLabel(s.RegistrationStatus)),
			mdTblExam(fee),
			mdTblExam(recAt))
	}

	if len(rows) == 0 {
		table.Reset()
		table.WriteString("_No exam registration rows on file for this academic record._\n")
	}

	return fmt.Sprintf(`# Exam registration acknowledgement

**Issued:** %s

## Student and academic record

| Field | Value |
|---|---|
| Name | %s |
| Enrolment no. | %s |
| Academic record id | %d |
| Program | %s |
| Session | %s |
| Term | %d |

## Registrations on file

%s

---

The institution acknowledges that the exam registration entries listed above are **on record** for this academic period. File attachments, when applicable, remain stored per institutional policy.

<div class="footer">Reference: academic record id %d · %d registration(s) listed · generated electronically</div>
`,
		mdTblExam(receiptIssuedExam(r)),
		mdTblExam(st.Name),
		mdTblExam(st.StudentNo),
		ar.ID,
		mdTblExam(prog),
		mdTblExam(ar.AdmissionSession.Name),
		ar.ProgramStructureUnit.TermNumber,
		table.String(),
		ar.ID,
		len(rows),
	)
}

func sanitizeReceiptBaseExam(s string) string {
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

func downloadAcademicRecordExamReceiptHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		idU, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil || idU == 0 {
			http.NotFound(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		q := views.QueryPatchers[p_nirmancampus_academicrecords.AcademicRecord]{
			{
				Key:   "academicrecords.preload",
				Value: views.QueryPatcherPreload[p_nirmancampus_academicrecords.AcademicRecord]{Fields: []string{"Student", "Program", "AdmissionSession", "ProgramStructureUnit"}},
			},
			{
				Key:   "academicrecords.scope_by_role",
				Value: p_nirmancampus_academicrecords.AcademicRecordScopeByRole,
			},
		}.Apply(views.View{}, r, gorm.G[p_nirmancampus_academicrecords.AcademicRecord](db).Scopes())

		ar, err := q.Where("ID = ?", uint(idU)).First(r.Context())
		if err != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: load academic record", "error", err, "id", idU)
			http.NotFound(w, r)
			return
		}

		var registrations []ExamRegistration
		if err := db.Where("academic_record_id = ?", ar.ID).
			Preload("Course").
			Order("created_at DESC").
			Order("id DESC").
			Find(&registrations).Error; err != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: list registrations", "error", err, "academicRecordID", ar.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: PDF converter", "error", err)
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		md := academicRecordExamPDFMarkdown(r, &ar, registrations)
		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: md,
			CSS:      pdfExamReceiptCSS,
		})
		if err != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: convert", "error", err)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		base := fmt.Sprintf("exam-receipt-record-%d", ar.ID)
		if strings.TrimSpace(ar.Student.StudentNo) != "" {
			base = sanitizeReceiptBaseExam(ar.Student.StudentNo) + "-exams-" + strconv.FormatUint(uint64(ar.ID), 10)
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, base))
		if _, werr := w.Write(result.PDF); werr != nil {
			slog.Error("downloadAcademicRecordExamReceiptHandler: write", "error", werr)
		}
	})
}
