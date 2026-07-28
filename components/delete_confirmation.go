package components

import (
	"context"
	"io"
	"net/http"

	"github.com/lariv-in/lariv/getters"
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

// Build compiles deleteConfirmSubmitBtn into a red destructive submit button.
func (deleteConfirmSubmitBtn) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, `<button type="submit" class="btn btn-error my-2">Confirm Delete</button>`)
	return err
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
	// Attr is a Getter yielding additional attributes to apply to the form (e.g., FormBubbling).
	Attr getters.Getter[HTMLAttributes]
}

// GetKey returns the unique key identifier for this DeleteConfirmation component.
func (e DeleteConfirmation) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this DeleteConfirmation.
func (e DeleteConfirmation) GetRoles() []string {
	return e.Roles
}

// Build compiles the DeleteConfirmation component into an HTML warning section with confirm/cancel submit actions.
func (e DeleteConfirmation) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	attrs, err := ResolveAttrs(ctx, e.Attr)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	var errorMessage string
	if errVal, lookupErr := getters.Key[error]("$error._global")(ctx); lookupErr == nil && errVal != nil {
		errorMessage = errVal.Error()
	}

	var formError string
	if errMap, ok := ctx.Value(getters.ContextKeyError).(map[string]error); ok {
		if formErr := errMap["_form"]; formErr != nil {
			formError = formErr.Error()
		}
	} else if errorMap, ok := ctx.Value(getters.ContextKeyError).(map[string]any); ok {
		if formErr, exists := errorMap["_form"]; exists && formErr != nil {
			if err, ok := formErr.(error); ok {
				formError = err.Error()
			}
		}
	}

	return Execute(w, "delete_confirmation", struct {
		Classes      string
		Title        string
		Message      string
		ErrorMessage string
		FormClasses  string
		Attrs        HTMLAttributes
		FormError    string
	}{
		Classes:      e.Classes,
		Title:        e.Title,
		Message:      e.Message,
		ErrorMessage: errorMessage,
		FormClasses:  "gap-2 my-4",
		Attrs:        attrs,
		FormError:    formError,
	})
}

// ParseForm parses the submitted deletion form parameters.
func (e DeleteConfirmation) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	inner := FormComponent[struct{}]{
		ChildrenAction: []PageInterface{deleteConfirmSubmitBtn{}},
	}
	return inner.ParseForm(r)
}
