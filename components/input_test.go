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
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
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
	var _ InputInterface = InputDuration{}
	var _ InputInterface = InputSelect[string]{}
	var _ InputInterface = InputStringList{}
}

func TestInputStringListParse(t *testing.T) {
	input := InputStringList{Name: "Options"}

	v, err := input.Parse([]string{`["a","b"]`}, context.Background())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if v != `["a","b"]` {
		t.Fatalf("expected json string, got %#v", v)
	}

	empty, err := input.Parse([]string{""}, context.Background())
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if empty != "[]" {
		t.Fatalf("expected [] json, got %#v", empty)
	}

	_, err = input.Parse([]string{`not-json`}, context.Background())
	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestInputSelectParse(t *testing.T) {
	choices := getters.Static([]registry.Pair[string, string]{
		{Key: "a", Value: "Alpha"},
		{Key: "b", Value: "Beta"},
	})
	input := InputSelect[string]{Name: "x", Choices: choices}

	v, err := input.Parse([]string{"b"}, context.Background())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if v != "b" {
		t.Fatalf("expected b, got %#v", v)
	}

	empty, err := input.Parse([]string{""}, context.Background())
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if empty != "" {
		t.Fatalf("expected empty string, got %#v", empty)
	}

	_, err = input.Parse([]string{"nope"}, context.Background())
	if err == nil {
		t.Fatalf("expected error for invalid choice")
	}
}

func TestInputSelectBuildSelected(t *testing.T) {
	choices := getters.Static([]registry.Pair[string, string]{
		{Key: "a", Value: "Alpha"},
		{Key: "b", Value: "Beta"},
	})
	current := getters.Static(registry.Pair[string, string]{Key: "b", Value: "Beta"})
	input := InputSelect[string]{Label: "Pick", Name: "pick", Choices: choices, Getter: current}

	html := renderNode(t, input.Build(context.Background()))
	if !strings.Contains(html, `selected`) || !strings.Contains(html, "Beta") {
		t.Fatalf("expected selected Beta option in html: %s", html)
	}
	if !strings.Contains(html, `name="pick"`) {
		t.Fatalf("expected name pick: %s", html)
	}
}

func TestHiddenInputsRenderTypedValues(t *testing.T) {
	ctx := context.WithValue(context.Background(), "$tz", time.UTC)
	dateValue := time.Date(2026, time.April, 15, 10, 45, 0, 0, time.UTC)

	cases := []struct {
		name     string
		html     string
		wantType string
		wantVal  string
	}{
		{
			name:     "checkbox",
			html:     renderNode(t, InputCheckbox{Name: "Enabled", Hidden: true, Getter: getters.Static(false)}.Build(context.Background())),
			wantType: `type="hidden"`,
			wantVal:  `value="false"`,
		},
		{
			name:     "date",
			html:     renderNode(t, InputDate{Name: "Date", Hidden: true, Getter: getters.Static(dateValue)}.Build(ctx)),
			wantType: `type="hidden"`,
			wantVal:  `value="2026-04-15"`,
		},
		{
			name:     "time",
			html:     renderNode(t, InputTime{Name: "Time", Hidden: true, Getter: getters.Static(dateValue)}.Build(ctx)),
			wantType: `type="hidden"`,
			wantVal:  `value="10:45"`,
		},
		{
			name:     "datetime",
			html:     renderNode(t, InputDatetime{Name: "Datetime", Hidden: true, Getter: getters.Static(dateValue)}.Build(ctx)),
			wantType: `type="hidden"`,
			wantVal:  `value="2026-04-15T10:45"`,
		},
	}

	for _, tc := range cases {
		if !strings.Contains(tc.html, tc.wantType) || !strings.Contains(tc.html, tc.wantVal) {
			t.Fatalf("%s: expected %s and %s in html: %s", tc.name, tc.wantType, tc.wantVal, tc.html)
		}
	}
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
		ChildrenInput: []PageInterface{
			InputFile{Name: "Attachment"},
		},
	}

	html := renderNode(t, form.Build(context.Background()))
	if !strings.Contains(html, `enctype="multipart/form-data"`) {
		t.Fatalf("expected multipart enctype in rendered form, got %s", html)
	}
}

func TestFormComponentBuildPreservesContextValuesOverGetter(t *testing.T) {
	type testModel struct {
		Name string
	}

	form := FormComponent[testModel]{
		Getter: getters.Static(testModel{Name: "getter value"}),
		ChildrenInput: []PageInterface{
			InputText{Name: "Name", Getter: getters.Key[string]("$in.Name")},
		},
	}
	ctx := context.WithValue(context.Background(), getters.ContextKeyIn, map[string]any{
		"Name": "context value",
	})
	ctx = context.WithValue(ctx, getters.ContextKeyError, map[string]error{
		"Name": fmt.Errorf("required"),
	})

	html := renderNode(t, form.Build(ctx))
	if !strings.Contains(html, `value="context value"`) {
		t.Fatalf("expected rerender context value to win over getter value, got %s", html)
	}
	if strings.Contains(html, `value="getter value"`) {
		t.Fatalf("expected getter value to be overridden by context value, got %s", html)
	}
}

func TestInputManyToManyParse(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testAssociationModel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	seed := []testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	if err := gorm.G[testAssociationModel](db).CreateInBatches(context.Background(), &seed, len(seed)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	input := InputManyToMany[testAssociationModel]{Name: "Teachers"}
	value, err := input.Parse([]string{"1", "2", "2"}, context.WithValue(context.Background(), getters.ContextKeyDB, db))
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
	seed := []testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	if err := gorm.G[testAssociationModel](db).CreateInBatches(context.Background(), &seed, len(seed)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	input := InputManyToMany[testAssociationModel]{
		Label:   "Teachers",
		Name:    "Teachers",
		Display: getters.Key[string]("$in.Name"),
	}
	ctx := context.WithValue(context.Background(), getters.ContextKeyDB, db)
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
	if !strings.Contains(html, `hasItem(value)`) {
		t.Fatalf("expected multi-select state to derive from items, got %s", html)
	}
	if !strings.Contains(html, `x-init="syncStore()"`) {
		t.Fatalf("expected many-to-many input to sync selected items into Alpine store, got %s", html)
	}
	if !strings.Contains(html, `Alpine.store(&#39;m2mSelections&#39;)`) {
		t.Fatalf("expected many-to-many input to render Alpine store sync logic, got %s", html)
	}
	if strings.Contains(html, `ev.target.selected`) {
		t.Fatalf("expected multi-select handler not to use row-local selected state, got %s", html)
	}
}

func TestInputManyToManyBuildEmptyStateUsesArray(t *testing.T) {
	input := InputManyToMany[testAssociationModel]{
		Label:       "Teachers",
		Name:        "Teachers",
		Placeholder: "Select teachers...",
	}

	html := renderNode(t, input.Build(context.Background()))
	if !strings.Contains(html, `items: []`) {
		t.Fatalf("expected empty items array in x-data, got %s", html)
	}
	if strings.Contains(html, `items: null`) {
		t.Fatalf("expected not to render null items, got %s", html)
	}
}

func TestFormComponentParseFormUsesRepeatedValuesForManyToMany(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testAssociationModel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	seed := []testAssociationModel{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	if err := gorm.G[testAssociationModel](db).CreateInBatches(context.Background(), &seed, len(seed)); err != nil {
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
	req = req.WithContext(context.WithValue(req.Context(), getters.ContextKeyDB, db))

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
