package components

import (
	"context"
	"io"
	"mime/multipart"
)

// InputFile represents a file upload form input component.
// It implements [MultipartInputInterface], rendering an HTML file selector input that supports single or multiple uploads and file type constraints (Accept filter).
//
// Use Cases:
//   - Uploading user avatars, resume attachments, documents, or data spreadsheets.
//
// Example:
//
//	&components.InputFile{
//	    Label:  "Attach Resume (PDF)",
//	    Name:   "resume",
//	    Accept: ".pdf",
//	}
type InputFile struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the file input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Required specifies if uploading a file is mandatory.
	Required bool
	// Multiple indicates if selecting multiple files is allowed.
	Multiple bool
	// Accept specifies file extensions or MIME-types allowed for selection (e.g. ".pdf", "image/*").
	Accept string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputFile component.
func (e InputFile) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputFile.
func (e InputFile) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputFile component into a Div wrapping a file selection Input.
func (e InputFile) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	return Execute(w, "input_file", struct {
		Classes  string
		Label    string
		Name     string
		Required bool
		Multiple bool
		Accept   string
	}{
		Classes:  e.Classes,
		Label:    e.Label,
		Name:     e.Name,
		Required: e.Required,
		Multiple: e.Multiple,
		Accept:   e.Accept,
	})
}

// Parse processes input parameter interfaces containing slice file headers and returns appropriate items.
func (e InputFile) Parse(v any, _ context.Context) (any, error) {
	files, _ := v.([]*multipart.FileHeader)
	if e.Multiple {
		if len(files) == 0 {
			return []*multipart.FileHeader(nil), nil
		}
		return files, nil
	}
	if len(files) == 0 {
		return (*multipart.FileHeader)(nil), nil
	}
	return files[0], nil
}

// ParseMultipart routes file upload processing to the standard parsing implementation.
func (e InputFile) ParseMultipart(files []*multipart.FileHeader, ctx context.Context) (any, error) {
	return e.Parse(files, ctx)
}

// GetName returns the HTML form element's name attribute value.
func (e InputFile) GetName() string {
	return e.Name
}
