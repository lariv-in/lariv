package components

import (
	"context"
	"fmt"
	"mime/multipart"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputFile struct {
	Page
	Label    string
	Name     string
	Required bool
	Multiple bool
	Accept   string
	Classes  string
}

func (e InputFile) GetKey() string {
	return e.Key
}

func (e InputFile) GetRoles() []string {
	return e.Roles
}

func (e InputFile) Build(_ context.Context) Node {
	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(
				Type("file"),
				Name(e.Name),
				Class(fmt.Sprintf("file-input file-input-bordered w-full %s", e.Classes)),
				If(e.Required, Required()),
				If(e.Multiple, Multiple()),
				If(e.Accept != "", Accept(e.Accept)),
			),
		),
	)
}

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

func (e InputFile) ParseMultipart(files []*multipart.FileHeader, ctx context.Context) (any, error) {
	return e.Parse(files, ctx)
}

func (e InputFile) GetName() string {
	return e.Name
}
