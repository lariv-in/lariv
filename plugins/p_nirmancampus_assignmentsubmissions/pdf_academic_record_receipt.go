package p_nirmancampus_assignmentsubmissions

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

const pdfAssignmentReceiptCSS = `
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

func mdTbl(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	return s
}

func receiptIssued(r *http.Request) string {
	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.Local
	}
	return time.Now().In(tz).Format("January 2, 2006")
}

func submissionStatusLabel(k string) string {
	if p, ok := registry.PairFromPairs(k, AssignmentSubmissionStatusChoices); ok {
		return p.Value
	}
	return k
}

func academicRecordAssignmentPDFMarkdown(r *http.Request, ar *p_nirmancampus_academicrecords.AcademicRecord, rows []AssignmentSubmission) string {
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
	fmt.Fprintf(&table, "| Assignment | Course | Status | Marks | Recorded |\n|---|---|---|---|---|\n")
	for _, s := range rows {
		marks := fmt.Sprintf("%d / %d", s.Marks, s.MaxMarks)
		course := s.Course.Name
		recAt := "—"
		if !s.CreatedAt.IsZero() {
			recAt = s.CreatedAt.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(&table, "| %s | %s | %s | %s | %s |\n",
			mdTbl(s.AssignmentTitle),
			mdTbl(course),
			mdTbl(submissionStatusLabel(s.SubmissionStatus)),
			mdTbl(marks),
			mdTbl(recAt))
	}

	if len(rows) == 0 {
		table.Reset()
		table.WriteString("_No assignment submission rows on file for this academic record._\n")
	}

	return fmt.Sprintf(`# Assignment submission acknowledgement

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

## Submissions on file

%s

---

The institution acknowledges that the assignment submission entries listed above are **on record** for this academic period. File attachments, when applicable, remain stored per institutional policy.

<div class="footer">Reference: academic record id %d · %d submission(s) listed · generated electronically</div>
`,
		mdTbl(receiptIssued(r)),
		mdTbl(st.Name),
		mdTbl(st.StudentNo),
		ar.ID,
		mdTbl(prog),
		mdTbl(ar.AdmissionSession.Name),
		ar.ProgramStructureUnit.TermNumber,
		table.String(),
		ar.ID,
		len(rows),
	)
}

func sanitizeReceiptBase(s string) string {
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

func downloadAcademicRecordAssignmentReceiptHandler(_ *views.View) http.Handler {
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
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: db from context", "error", dberr)
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
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: load academic record", "error", err, "id", idU)
			http.NotFound(w, r)
			return
		}

		var submissions []AssignmentSubmission
		if err := db.Where("academic_record_id = ?", ar.ID).
			Preload("Course").
			Order("created_at DESC").
			Order("id DESC").
			Find(&submissions).Error; err != nil {
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: list submissions", "error", err, "academicRecordID", ar.ID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: PDF converter", "error", err)
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		md := academicRecordAssignmentPDFMarkdown(r, &ar, submissions)
		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: md,
			CSS:      pdfAssignmentReceiptCSS,
		})
		if err != nil {
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: convert", "error", err)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		base := fmt.Sprintf("assignment-receipt-record-%d", ar.ID)
		if strings.TrimSpace(ar.Student.StudentNo) != "" {
			base = sanitizeReceiptBase(ar.Student.StudentNo) + "-assignments-" + strconv.FormatUint(uint64(ar.ID), 10)
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, base))
		if _, werr := w.Write(result.PDF); werr != nil {
			slog.Error("downloadAcademicRecordAssignmentReceiptHandler: write", "error", werr)
		}
	})
}
