package components

import (
	"context"
	"net/http"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ FormInterface = DeleteConfirmation{}

// deleteConfirmSubmitBtn represents the internal destructive submit button rendered inside DeleteConfirmation forms.
type deleteConfirmSubmitBtn struct {
	Page
}

// GetKey returns the unique key identifier for deleteConfirmSubmitBtn.
func (e deleteConfirmSubmitBtn) GetKey() string { return e.Key }

// GetRoles returns the authorized roles required to view deleteConfirmSubmitBtn.
func (e deleteConfirmSubmitBtn) GetRoles() []string { return e.Roles }

// Build compiles deleteConfirmSubmitBtn into a red destructive submit button Node.
func (deleteConfirmSubmitBtn) Build(context.Context) Node {
	return Button(Type("submit"), Class("btn btn-error my-2"), Text("Confirm Delete"))
}

// DeleteConfirmation represents a destructive form/modal content that warns users before performing delete operations.
// It integrates with [FormComponent] and displays any global error message under "$error._global".
//
// Use Cases:
//   - Showing a verification popup/modal before deleting critical items (e.g., invoices, records, users).
//
// Example:
//
//	&components.DeleteConfirmation{
//	    Title:   "Confirm deletion",
//	    Message: "Are you sure you want to delete this invoice? This action cannot be undone.",
//	    Attr:    getters.FormBubbling(),
//	}
type DeleteConfirmation struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title is the modal heading text (usually styled in red to indicate a destructive action).
	Title string
	// Message is the descriptive warning text indicating the consequences of deletion.
	Message string
	// Classes represents additional CSS classes applied to the outer div wrapper.
	Classes string
	// Attr is a Getter yielding additional attributes (Node) to apply to the form (e.g., FormBubbling).
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this DeleteConfirmation component.
func (e DeleteConfirmation) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this DeleteConfirmation.
func (e DeleteConfirmation) GetRoles() []string {
	return e.Roles
}

// deleteConfirmationGlobalError resolves and returns a global error banner node if $error._global contains an error.
func deleteConfirmationGlobalError(ctx context.Context) Node {
	err, lookupErr := getters.Key[error]("$error._global")(ctx)
	if lookupErr != nil || err == nil {
		return nil
	}
	return Div(Class("alert alert-error my-2 text-sm"), Text(err.Error()))
}

// Build compiles the DeleteConfirmation component into an HTML warning section with confirm/cancel submit actions.
func (e DeleteConfirmation) Build(ctx context.Context) Node {
	form := FormComponent[struct{}]{
		Classes:        "gap-2 my-4",
		Attr:           e.Attr,
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}

	return Div(
		Class("container mx-auto "+e.Classes),
		H2(Class("text-xl font-bold text-error"), Text(e.Title)),
		P(Class("my-2"), Text(e.Message)),
		deleteConfirmationGlobalError(ctx),
		form.Build(ctx),
	)
}

// ParseForm parses the submitted deletion form parameters.
func (e DeleteConfirmation) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	inner := FormComponent[struct{}]{
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}
	return inner.ParseForm(r)
}
