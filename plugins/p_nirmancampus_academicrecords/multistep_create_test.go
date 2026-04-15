package p_nirmancampus_academicrecords

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/views"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"maragu.dev/gomponents"
)

func TestAcademicRecordCreateFormFirstStageOnlyShowsInitialFields(t *testing.T) {
	page, ok := lago.RegistryPage.Get("academicrecords.AcademicRecordCreateForm")
	if !ok {
		t.Fatal("expected create form page in registry")
	}

	req := httptest.NewRequest(http.MethodGet, "/academicrecords/create/?name=academicrecords.AcademicRecordCreateForm", nil)
	ctx := context.WithValue(context.Background(), "$request", req)
	ctx = context.WithValue(ctx, "$get", map[string]any{"name": "academicrecords.AcademicRecordCreateForm"})
	ctx = context.WithValue(ctx, "$stage", 0)

	html := renderAcademicRecordNode(t, components.Render(page, ctx))
	if !strings.Contains(html, "Create Academic Record") {
		t.Fatalf("expected first-stage title, got %s", html)
	}
	if !strings.Contains(html, "Continue") {
		t.Fatalf("expected continue button on first stage, got %s", html)
	}
	if strings.Contains(html, "Optional courses") {
		t.Fatalf("expected optional-course stage hidden on first stage, got %s", html)
	}
}

func TestAcademicRecordCreateFormSecondStageShowsProgramStructureUnitChoices(t *testing.T) {
	page, ok := lago.RegistryPage.Get("academicrecords.AcademicRecordCreateForm")
	if !ok {
		t.Fatal("expected create form page in registry")
	}

	req := httptest.NewRequest(http.MethodGet, "/academicrecords/create/?name=academicrecords.AcademicRecordCreateForm", nil)
	ctx := context.WithValue(context.Background(), "$request", req)
	ctx = context.WithValue(ctx, "$get", map[string]any{"name": "academicrecords.AcademicRecordCreateForm"})
	ctx = context.WithValue(ctx, "$stage", 1)
	ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{
		"SessionID": uint(1),
		"StudentID": uint(2),
		"ProgramID": uint(3),
		"Status":    "Enrolled",
		"Date":      time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC),
	}, nil)
	ctx = context.WithValue(ctx, academicRecordProgramStructureUnitsContextKey, []p_nirmancampus_programs.ProgramStructureUnit{
		{Model: gorm.Model{ID: 31}, TermNumber: 1},
		{Model: gorm.Model{ID: 32}, TermNumber: 2},
	})

	html := renderAcademicRecordNode(t, components.Render(page, ctx))
	if !strings.Contains(html, "Select Term") {
		t.Fatalf("expected second-stage title, got %s", html)
	}
	if !strings.Contains(html, `name="ProgramStructureUnitID"`) || !strings.Contains(html, `program-structure-units/select/?ProgramID=3`) {
		t.Fatalf("expected foreign-key program structure unit field in second stage, got %s", html)
	}
	if !strings.Contains(html, "Select...") {
		t.Fatalf("expected FK placeholder in second stage, got %s", html)
	}
	if !strings.Contains(html, "Continue") {
		t.Fatalf("expected continue button on second stage, got %s", html)
	}
	if !strings.Contains(html, `name="$stage"`) || !strings.Contains(html, `value="1"`) {
		t.Fatalf("expected stage marker for second stage, got %s", html)
	}
}

func TestAcademicRecordCreateFormThirdStageShowsProgramCourses(t *testing.T) {
	page, ok := lago.RegistryPage.Get("academicrecords.AcademicRecordCreateForm")
	if !ok {
		t.Fatal("expected create form page in registry")
	}

	req := httptest.NewRequest(http.MethodGet, "/academicrecords/create/?name=academicrecords.AcademicRecordCreateForm", nil)
	ctx := context.WithValue(context.Background(), "$request", req)
	ctx = context.WithValue(ctx, "$get", map[string]any{"name": "academicrecords.AcademicRecordCreateForm"})
	ctx = context.WithValue(ctx, "$stage", 2)
	ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{
		"SessionID":              uint(1),
		"StudentID":              uint(2),
		"ProgramID":              uint(3),
		"ProgramStructureUnitID": uint(99),
		"Status":                 "Enrolled",
		"Date":                   time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC),
		"OptionalCourses": []p_nirmancampus_courses.Course{
			{Model: gorm.Model{ID: 21}, Name: "Elective A"},
		},
	}, nil)
	ctx = context.WithValue(ctx, academicRecordProgramStructureUnitContextKey, p_nirmancampus_programs.ProgramStructureUnit{
		Model:               gorm.Model{ID: 99},
		TermNumber:          2,
		OptionalCourseCount: 1,
		CompulsoryCourses: []p_nirmancampus_courses.Course{
			{Model: gorm.Model{ID: 11}, Name: "Core Course"},
		},
		OptionalCourseSelectionPool: []p_nirmancampus_courses.Course{
			{Model: gorm.Model{ID: 21}, Name: "Elective A"},
		},
	})

	html := renderAcademicRecordNode(t, components.Render(page, ctx))
	if !strings.Contains(html, "Select Courses") {
		t.Fatalf("expected third-stage title, got %s", html)
	}
	if !strings.Contains(html, "Core Course") {
		t.Fatalf("expected compulsory course in third stage, got %s", html)
	}
	if !strings.Contains(html, "Optional courses") || !strings.Contains(html, "pool_course_ids=21") {
		t.Fatalf("expected optional course selector limited to pool, got %s", html)
	}
	if !strings.Contains(html, "Save Academic Record") {
		t.Fatalf("expected final submit button, got %s", html)
	}
	if !strings.Contains(html, `name="$stage"`) || !strings.Contains(html, `value="2"`) {
		t.Fatalf("expected stage marker for third stage, got %s", html)
	}
}

func TestAcademicRecordProgramStructureUnitContextLayerUsesCreateValues(t *testing.T) {
	db := openAcademicRecordTestDB(t)
	program := p_nirmancampus_programs.Program{Name: "BCA", Code: "BCA"}
	if err := db.Create(&program).Error; err != nil {
		t.Fatalf("create program: %v", err)
	}
	core := p_nirmancampus_courses.Course{Name: "Core Course", Code: "CORE-1"}
	elective := p_nirmancampus_courses.Course{Name: "Elective A", Code: "EL-1"}
	if err := db.Create(&core).Error; err != nil {
		t.Fatalf("create core course: %v", err)
	}
	if err := db.Create(&elective).Error; err != nil {
		t.Fatalf("create elective course: %v", err)
	}
	psu := p_nirmancampus_programs.ProgramStructureUnit{
		ProgramID:           program.ID,
		TermNumber:          2,
		OptionalCourseCount: 1,
	}
	if err := db.Create(&psu).Error; err != nil {
		t.Fatalf("create psu: %v", err)
	}
	if err := db.Model(&psu).Association("CompulsoryCourses").Append(&core); err != nil {
		t.Fatalf("append compulsory: %v", err)
	}
	if err := db.Model(&psu).Association("OptionalCourseSelectionPool").Append(&elective); err != nil {
		t.Fatalf("append optional pool: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/academicrecords/create/", nil)
	ctx := context.WithValue(req.Context(), getters.ContextKeyDB, db)
	ctx = views.ContextWithErrorsAndValues(ctx, map[string]any{
		"ProgramID":              program.ID,
		"ProgramStructureUnitID": psu.ID,
	}, nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	academicRecordProgramStructureUnitContextLayer{}.Next(views.View{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, ok := r.Context().Value(academicRecordProgramStructureUnitContextKey).(p_nirmancampus_programs.ProgramStructureUnit)
		if !ok {
			t.Fatal("expected psu in context")
		}
		if len(value.CompulsoryCourses) != 1 || value.CompulsoryCourses[0].Name != "Core Course" {
			t.Fatalf("unexpected compulsory courses: %#v", value.CompulsoryCourses)
		}
		if len(value.OptionalCourseSelectionPool) != 1 || value.OptionalCourseSelectionPool[0].Name != "Elective A" {
			t.Fatalf("unexpected optional pool: %#v", value.OptionalCourseSelectionPool)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected layer to continue, got %d", rec.Code)
	}
}

func openAcademicRecordTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	if err := db.AutoMigrate(
		&p_nirmancampus_programs.Program{},
		&p_nirmancampus_courses.Course{},
		&p_nirmancampus_programs.ProgramStructureUnit{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	return db
}

func renderAcademicRecordNode(t *testing.T, node gomponents.Node) string {
	t.Helper()
	var out bytes.Buffer
	if err := node.Render(io.Writer(&out)); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return out.String()
}
