package p_nirmancampus_academicrecords

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alnah/go-md2pdf"
	"github.com/lariv-in/lago/getters"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const pdfAcademicRecordCSS = `
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

func mdPipe(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	return s
}

func issuedLocal(r *http.Request) string {
	tz, _ := r.Context().Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.Local
	}
	return time.Now().In(tz).Format("January 2, 2006")
}

func programDisplay(ar *AcademicRecord) string {
	name := ar.Program.Name
	uk := strings.TrimSpace(ar.Program.University)
	if uk == "" {
		return name
	}
	if p, ok := registry.PairFromPairs(uk, p_nirmancampus_programs.UniversityChoices); ok {
		return fmt.Sprintf("%s (%s)", name, p.Value)
	}
	return fmt.Sprintf("%s (%s)", name, uk)
}

func statusLabel(ar *AcademicRecord) string {
	if p, ok := registry.PairFromPairs(ar.Status, AcademicRecordStatusChoices); ok {
		return p.Value
	}
	return ar.Status
}

func courseLines(list []courses.Course) string {
	if len(list) == 0 {
		return "_—_"
	}
	var b strings.Builder
	for _, c := range list {
		line := "- " + mdPipe(c.Name)
		if t := strings.TrimSpace(c.Code); t != "" {
			line += fmt.Sprintf(" (%s)", mdPipe(t))
		}
		fmt.Fprintf(&b, "%s\n", line)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func academicRecordPDFMarkdown(r *http.Request, ar *AcademicRecord) string {
	st := &ar.Student
	admissionDate := "—"
	if !ar.Date.IsZero() {
		admissionDate = ar.Date.Format("January 2, 2006")
	}

	return fmt.Sprintf(`# Academic record

**Issued:** %s

## Student

| Field | Value |
|---|---|
| Name | %s |
| Enrolment no. | %s |
| Email | %s |
| Phone | %s |

## Program and session

| Field | Value |
|---|---|
| Record id | %d |
| Program | %s |
| Admission session | %s |
| Term | %d |
| Status | %s |
| Admission / record date | %s |

## Courses

### Compulsory

%s

### Optional

%s

---

This certifies that the student named above is **recorded** for the program, session, and term shown, with the course selections listed.

<div class="footer">Reference: academic record id %d · generated electronically</div>
`,
		mdPipe(issuedLocal(r)),
		mdPipe(st.Name),
		mdPipe(st.StudentNo),
		mdPipe(st.Email),
		mdPipe(st.Phone),
		ar.ID,
		mdPipe(programDisplay(ar)),
		mdPipe(ar.AdmissionSession.Name),
		ar.ProgramStructureUnit.TermNumber,
		mdPipe(statusLabel(ar)),
		mdPipe(admissionDate),
		courseLines(ar.CompulsoryCourses),
		courseLines(ar.OptionalCourses),
		ar.ID,
	)
}

func sanitizeBase(s string) string {
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

func downloadAcademicRecordPDFHandler(_ *views.View) http.Handler {
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
			slog.Error("downloadAcademicRecordPDFHandler: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		q := views.QueryPatchers[AcademicRecord]{
			{Key: "academicrecords.preload", Value: views.QueryPatcherPreload[AcademicRecord]{Fields: []string{
				"Student", "Program", "AdmissionSession", "ProgramStructureUnit", "CompulsoryCourses", "OptionalCourses",
			}}},
			{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
		}.Apply(views.View{}, r, gorm.G[AcademicRecord](db).Scopes())

		rec, err := q.Where("ID = ?", uint(id)).First(r.Context())
		if err != nil {
			slog.Error("downloadAcademicRecordPDFHandler: load record", "error", err, "id", id)
			http.NotFound(w, r)
			return
		}

		conv, err := md2pdf.NewConverter()
		if err != nil {
			slog.Error("downloadAcademicRecordPDFHandler: PDF converter", "error", err)
			http.Error(w, "PDF converter unavailable", http.StatusInternalServerError)
			return
		}
		defer conv.Close()

		md := academicRecordPDFMarkdown(r, &rec)
		result, err := conv.Convert(r.Context(), md2pdf.Input{
			Markdown: md,
			CSS:      pdfAcademicRecordCSS,
		})
		if err != nil {
			slog.Error("downloadAcademicRecordPDFHandler: convert", "error", err, "recordID", rec.ID)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		base := fmt.Sprintf("academic-record-%d", rec.ID)
		if rec.Student.StudentNo != "" {
			base = sanitizeBase(rec.Student.StudentNo) + "-record-" + strconv.FormatUint(uint64(rec.ID), 10)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, base))
		if _, err := w.Write(result.PDF); err != nil {
			slog.Error("downloadAcademicRecordPDFHandler: write", "error", err)
		}
	})
}
