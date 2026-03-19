package components

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"maragu.dev/gomponents"
)

func TestInputImplementations(t *testing.T) {
	// Compile-time checks to ensure that the input components implement the InputInterface
	var _ InputInterface = InputCheckbox{}
	var _ InputInterface = InputEmail{}
	var _ InputInterface = InputFile{}
	var _ MultipartInputInterface = InputFile{}
	var _ InputInterface = InputPassword{}
	var _ InputInterface = InputPhone{}
	var _ InputInterface = InputText{}
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

func renderNode(t *testing.T, node gomponents.Node) string {
	t.Helper()
	var out bytes.Buffer
	if err := node.Render(&out); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out.String()
}
