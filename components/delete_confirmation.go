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
type deleteConfirmSubmitBtn struct {
	Page
}

func (e deleteConfirmSubmitBtn) GetKey() string     { return e.Key }
func (e deleteConfirmSubmitBtn) GetRoles() []string { return e.Roles }
func (deleteConfirmSubmitBtn) Build(context.Context) Node {
	return Button(Type("submit"), Class("btn btn-error my-2"), Text("Confirm Delete"))
}

type DeleteConfirmation struct {
	Page
	Title   string
	Message string
	Classes string
	// Attr is merged onto the form; use [getters.FormBubbling] and wire [ButtonModalForm] (opening this modal) with FormPostURL set to the resource DeleteRoute.
	// For row-scoped deletes with a shared modal id, use [getters.FormBubblingWithDataPostURL] here instead of [getters.FormBubbling] alone.
	// Modal HTML is a fragment (no [ShellBase]), so Build also shows "$error._global" for [views.LayerDelete] failures.
	Attr getters.Getter[Node]
}

func (e DeleteConfirmation) GetKey() string {
	return e.Key
}

func (e DeleteConfirmation) GetRoles() []string {
	return e.Roles
}

func deleteConfirmationGlobalError(ctx context.Context) Node {
	err, lookupErr := getters.Key[error]("$error._global")(ctx)
	if lookupErr != nil || err == nil {
		return nil
	}
	return Div(Class("alert alert-error my-2 text-sm"), Text(err.Error()))
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
		deleteConfirmationGlobalError(ctx),
		form.Build(ctx),
	)
}

func (e DeleteConfirmation) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	inner := FormComponent[struct{}]{
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}
	return inner.ParseForm(r)
}
