package components

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lariv-in/getters"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"maragu.dev/gomponents"
)

func TestInputImplementations(t *testing.T) {
	// Compile-time checks to ensure that the input components implement the InputInterface
	var _ InputInterface = InputCheckbox{}
	var _ InputInterface = InputEmail{}
	var _ InputInterface = InputFile{}
	var _ MultipartInputInterface = InputFile{}
	var _ InputInterface = InputManyToMany[testAssociationModel]{}
	var _ InputInterface = InputPassword{}
	var _ InputInterface = InputPhone{}
	var _ InputInterface = InputText{}
}

type testAssociationModel struct {
	ID   uint
	Name string
}

func TestInputFileParseSingle(t *testing.T) {
	input := InputFile{Name: "file"}
	file := &multipart.FileHeader{Filename: "hello.txt"}

	value, err := input.ParseMultipart([]*multipart.FileHeader{file}, context.Background())
	if err != nil {
		t.Fatalf("ParseMultipart returned error: %v", err)
	}

	got, ok := value.(*multipart.FileHeader)
	if !ok {
		t.Fatalf("expected *multipart.FileHeader, got %T", value)
	}
	if got.Filename != "hello.txt" {
		t.Fatalf("expected filename hello.txt, got %q", got.Filename)
	}
}

func TestInputFileParseMultiple(t *testing.T) {
	input := InputFile{Name: "files", Multiple: true}
	files := []*multipart.FileHeader{
		{Filename: "a.txt"},
		{Filename: "b.txt"},
	}

	value, err := input.ParseMultipart(files, context.Background())
	if err != nil {
		t.Fatalf("ParseMultipart returned error: %v", err)
	}

	got, ok := value.([]*multipart.FileHeader)
	if !ok {
		t.Fatalf("expected []*multipart.FileHeader, got %T", value)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d", len(got))
	}
}

func TestFormComponentParseFormMultipartUsesFileHeaders(t *testing.T) {
	form := FormComponent[struct{}]{
		ChildrenInput: []PageInterface{
			InputText{Name: "Title"},
			InputFile{Name: "Attachment"},
			InputFile{Name: "Files", Multiple: true},
		},
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("Title", "Report"); err != nil {
		t.Fatalf("WriteField Title failed: %v", err)
	}
	fileWriter, err := writer.CreateFormFile("Attachment", "report.txt")
	if err != nil {
		t.Fatalf("CreateFormFile Attachment failed: %v", err)
	}
	if _, err := fileWriter.Write([]byte("hello")); err != nil {
		t.Fatalf("write attachment failed: %v", err)
	}
	fileWriter, err = writer.CreateFormFile("Files", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile Files[0] failed: %v", err)
	}
	if _, err := fileWriter.Write([]byte("a")); err != nil {
		t.Fatalf("write files[0] failed: %v", err)
	}
	fileWriter, err = writer.CreateFormFile("Files", "b.txt")
	if err != nil {
		t.Fatalf("CreateFormFile Files[1] failed: %v", err)
	}
	if _, err := fileWriter.Write([]byte("b")); err != nil {
		t.Fatalf("write files[1] failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	values, errs, err := form.ParseForm(req)
	if err != nil {
		t.Fatalf("ParseForm returned error: %v", err)
	}
	if errs["Title"] != nil || errs["Attachment"] != nil || errs["Files"] != nil {
		t.Fatalf("unexpected field errors: %#v", errs)
	}
	if title, _ := values["Title"].(string); title != "Report" {
		t.Fatalf("expected Title to be Report, got %#v", values["Title"])
	}
	attachment, ok := values["Attachment"].(*multipart.FileHeader)
	if !ok || attachment == nil {
		t.Fatalf("expected Attachment to be *multipart.FileHeader, got %#v", values["Attachment"])
	}
	if attachment.Filename != "report.txt" {
		t.Fatalf("expected report.txt, got %q", attachment.Filename)
	}
	files, ok := values["Files"].([]*multipart.FileHeader)
	if !ok {
		t.Fatalf("expected Files to be []*multipart.FileHeader, got %#v", values["Files"])
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestFormComponentBuildAddsMultipartEnctype(t *testing.T) {
	form := FormComponent[struct{}]{
		Method: http.MethodPost,
		ChildrenInput: []PageInterface{
			InputFile{Name: "Attachment"},
		},
	}

	html := renderNode(t, form.Build(context.Background()))
	if !strings.Contains(html, `enctype="multipart/form-data"`) {
		t.Fatalf("expected multipart enctype in rendered form, got %s", html)
	}
}

func TestInputManyToManyParse(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testAssociationModel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Create([]testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	input := InputManyToMany[testAssociationModel]{Name: "Teachers"}
	value, err := input.Parse([]string{"1", "2", "2"}, context.WithValue(context.Background(), "$db", db))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	got, ok := value.(AssociationIDs)
	if !ok {
		t.Fatalf("expected AssociationIDs, got %T", value)
	}
	if got.Field != "Teachers" {
		t.Fatalf("expected field Teachers, got %q", got.Field)
	}
	if len(got.IDs) != 2 || got.IDs[0] != 1 || got.IDs[1] != 2 {
		t.Fatalf("unexpected ids: %#v", got.IDs)
	}
}

func TestInputManyToManyBuildUsesAssociationIDsContext(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testAssociationModel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Create([]testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	input := InputManyToMany[testAssociationModel]{
		Label:   "Teachers",
		Name:    "Teachers",
		Display: getters.GetterKey[string]("$in.Name"),
	}
	ctx := context.WithValue(context.Background(), "$db", db)
	ctx = context.WithValue(ctx, getters.ContextKeyIn, map[string]any{
		"Teachers": AssociationIDs{Field: "Teachers", IDs: []uint{2, 1}},
	})

	html := renderNode(t, input.Build(ctx))
	if !strings.Contains(html, "Alpha") || !strings.Contains(html, "Beta") {
		t.Fatalf("expected selected names in rendered html, got %s", html)
	}
	if !strings.Contains(html, `@fk-multi-select.window`) {
		t.Fatalf("expected multi-select event handler, got %s", html)
	}
}

func TestFormComponentParseFormUsesRepeatedValuesForManyToMany(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testAssociationModel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Create([]testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	form := FormComponent[struct{}]{
		ChildrenInput: []PageInterface{
			InputManyToMany[testAssociationModel]{Name: "Teachers"},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/?ignored=1", strings.NewReader(url.Values{
		"Teachers": {"1", "2"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), "$db", db))

	values, errs, err := form.ParseForm(req)
	if err != nil {
		t.Fatalf("ParseForm returned error: %v", err)
	}
	if errs["Teachers"] != nil {
		t.Fatalf("unexpected field error: %v", errs["Teachers"])
	}
	got, ok := values["Teachers"].(AssociationIDs)
	if !ok {
		t.Fatalf("expected AssociationIDs, got %#v", values["Teachers"])
	}
	if len(got.IDs) != 2 || got.IDs[0] != 1 || got.IDs[1] != 2 {
		t.Fatalf("unexpected ids: %#v", got.IDs)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	return db
}

func renderNode(t *testing.T, node gomponents.Node) string {
	t.Helper()
	var out bytes.Buffer
	if err := node.Render(&out); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out.String()
}
