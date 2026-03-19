package views

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type crudTestTeacher struct {
	ID   uint
	Name string
}

type crudTestUser struct {
	ID   uint
	Name string
}

type crudTestCourse struct {
	ID       uint
	Name     string
	Teachers []crudTestTeacher `gorm:"many2many:crud_test_course_teachers;"`
}

type crudTestStudent struct {
	ID     uint
	Name   string
	UserID uint
	User   crudTestUser
}

func TestCreateViewPersistsManyToManyAssociations(t *testing.T) {
	db := openCrudTestDB(t)
	if err := db.AutoMigrate(&crudTestTeacher{}, &crudTestCourse{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Create([]crudTestTeacher{
		{ID: 1, Name: "Ada"},
		{ID: 2, Name: "Grace"},
	}).Error; err != nil {
		t.Fatalf("Create teachers failed: %v", err)
	}

	view := newCrudTestView(createCoursePage())
	CreateView[crudTestCourse](getters.GetterStatic("/courses/1/"))(view)

	req := httptest.NewRequest(http.MethodPost, "/", stringsReader(url.Values{
		"Name":     {"Systems"},
		"Teachers": {"1", "2"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), "$db", db))
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}

	var course crudTestCourse
	if err := db.Preload("Teachers").First(&course).Error; err != nil {
		t.Fatalf("loading course failed: %v", err)
	}
	if course.Name != "Systems" {
		t.Fatalf("expected Systems, got %q", course.Name)
	}
	if len(course.Teachers) != 2 {
		t.Fatalf("expected 2 teachers, got %d", len(course.Teachers))
	}
}

func TestUpdateViewReplacesManyToManyAssociations(t *testing.T) {
	db := openCrudTestDB(t)
	if err := db.AutoMigrate(&crudTestTeacher{}, &crudTestCourse{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Create([]crudTestTeacher{
		{ID: 1, Name: "Ada"},
		{ID: 2, Name: "Grace"},
	}).Error; err != nil {
		t.Fatalf("Create teachers failed: %v", err)
	}

	course := crudTestCourse{Name: "Original", Teachers: []crudTestTeacher{{ID: 1}}}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("Create course failed: %v", err)
	}

	view := newCrudTestView(createCoursePage())
	UpdateView[crudTestCourse](getters.GetterStatic("/courses/1/"))(view)

	req := httptest.NewRequest(http.MethodPost, "/", stringsReader(url.Values{
		"Name":     {"Updated"},
		"Teachers": {"2"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), "$db", db))
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}

	var updated crudTestCourse
	if err := db.Preload("Teachers").First(&updated, course.ID).Error; err != nil {
		t.Fatalf("loading updated course failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Fatalf("expected Updated, got %q", updated.Name)
	}
	if len(updated.Teachers) != 1 || updated.Teachers[0].ID != 2 {
		t.Fatalf("expected only teacher 2, got %#v", updated.Teachers)
	}
}

func TestUpdateViewWithPreloadedBelongsToDoesNotDuplicateForeignKeyAssignments(t *testing.T) {
	db := openCrudTestDB(t)
	if err := db.AutoMigrate(&crudTestUser{}, &crudTestStudent{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	user := crudTestUser{Name: "User One"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Create user failed: %v", err)
	}
	student := crudTestStudent{Name: "Before", UserID: user.ID}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("Create student failed: %v", err)
	}

	view := newCrudTestView(&components.ContainerColumn{
		Children: []components.PageInterface{
			&components.FormComponent[crudTestStudent]{
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputText{Name: "Name"},
					components.InputForeignKey[crudTestUser]{Name: "UserID"},
				},
			},
		},
	})
	view.WithQueryPatcher("students.preload_user", QueryPatcherPreload("User"))
	UpdateView[crudTestStudent](getters.GetterStatic("/students/1/"))(view)

	req := httptest.NewRequest(http.MethodPost, "/", stringsReader(url.Values{
		"Name":   {"After"},
		"UserID": {fmt.Sprintf("%d", user.ID)},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", fmt.Sprintf("%d", student.ID))
	req = req.WithContext(context.WithValue(req.Context(), "$db", db))
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}

	var updated crudTestStudent
	if err := db.First(&updated, student.ID).Error; err != nil {
		t.Fatalf("loading updated student failed: %v", err)
	}
	if updated.Name != "After" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
	if updated.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, updated.UserID)
	}
}

func createCoursePage() *components.ContainerColumn {
	return &components.ContainerColumn{
		Children: []components.PageInterface{
			&components.FormComponent[crudTestCourse]{
				Method: http.MethodPost,
				ChildrenInput: []components.PageInterface{
					components.InputText{Name: "Name"},
					components.InputManyToMany[crudTestTeacher]{
						Name:    "Teachers",
						Display: getters.GetterKey[string]("$in.Name"),
					},
				},
			},
		},
	}
}

func newCrudTestView(page components.PageInterface) *View {
	return &View{
		PageName: "test.form",
		Registry: map[string]components.PageInterface{
			"test.form": page,
		},
		Handlers:      map[string]func(*View) http.Handler{},
		FormPatchers:  nil,
		QueryPatchers: nil,
		Middlewares:   nil,
	}
}

func openCrudTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	return db
}

func stringsReader(value string) *strings.Reader {
	return strings.NewReader(value)
}
