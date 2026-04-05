package components

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ FormInterface = DeleteConfirmation{}

// deleteConfirmSubmitBtn is the destructive submit action for delete flows.
type deleteConfirmSubmitBtn struct{}

func (deleteConfirmSubmitBtn) GetKey() string       { return "" }
func (deleteConfirmSubmitBtn) GetRoles() []string   { return nil }
func (deleteConfirmSubmitBtn) Build(context.Context) Node {
	return Button(Type("submit"), Class("btn btn-error my-2"), Text("Confirm Delete"))
}

type DeleteConfirmation struct {
	Page
	Title   string
	Message string
	Classes string
	// Attr is merged onto the form (method, @submit.prevent, etc.); use getters.FormAttr with getters.FormSubmit pointing at the resource DeleteRoute.
	Attr getters.Getter[Node]
}

func (e DeleteConfirmation) GetKey() string {
	return e.Key
}

func (e DeleteConfirmation) GetRoles() []string {
	return e.Roles
}

func (e DeleteConfirmation) Build(ctx context.Context) Node {
	form := FormComponent[struct{}]{
		Classes:        "gap-2 my-4",
		Attr:           e.Attr,
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}

	return Div(Class("container mx-auto "+e.Classes),
		H2(Class("text-xl font-bold text-error"), Text(e.Title)),
		P(Class("my-2"), Text(e.Message)),
		form.Build(ctx),
	)
}

func (e DeleteConfirmation) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	inner := FormComponent[struct{}]{
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}
	return inner.ParseForm(r)
}
