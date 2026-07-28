package components

import (
	"context"
	"html/template"
	"io"
)

// HTMXTargetBodyModal / HTMXSwapBodyModal append modal markup as a direct child of document.body.
// Target configuration keys mapping HTMX target swapping actions to specific layout hosts.
const (
	// HTMXTargetBodyModal / HTMXSwapBodyModal appends modal markup as a direct child of document.body.
	HTMXTargetBodyModal = "body"
	HTMXSwapBodyModal   = "beforeend"
	// HTMXTargetLocalModal / HTMXSwapLocalModal appends modal markup into the closest
	// local modal host container (used by ButtonModal and ButtonModalForm).
	HTMXTargetLocalModal = "closest .fk-modal-host"
	HTMXSwapLocalModal   = "beforeend"
)

// Modal represents a responsive overlay dialog pop-up component.
// It renders an HTML dialog element styled with open classes, including absolute top-right close buttons and overlay backdrop actions that remove the element from the DOM.
//
// Use Cases:
//   - Showing confirmation alerts, detail profiles, editor overlay forms, or dynamic option selectors without full-page navigation.
//
// Example:
//
//	 &components.Modal{
//	     UID: "warning-dialog",
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Proceed?"},
//	         &components.ButtonLink{Label: getters.Static("Cancel"), Link: getters.Static("#")},
//	     },
//	 }
type Modal struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// UID represents the unique HTML element ID used by DOM scripts to close or query the dialog.
	UID string
	// Children represents the slice of sub-components rendered in the modal viewport box.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// Build compiles the Modal component into a Dialog layout wrapper.
func (e Modal) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	closeIcon, err := RenderHTML(Icon{Name: "x-mark"}, cat, ctx)
	if err != nil {
		return err
	}
	onclick := "document.getElementById('" + e.UID + "').remove()"
	return Execute(w, "modal", struct {
		UID           string
		Classes       string
		CloseIcon     template.HTML
		Children      template.HTML
		CloseAttrs    HTMLAttributes
		BackdropAttrs HTMLAttributes
	}{
		UID:           e.UID,
		Classes:       e.Classes,
		CloseIcon:     closeIcon,
		Children:      children,
		CloseAttrs:    HTMLAttributes{"onclick": onclick},
		BackdropAttrs: HTMLAttributes{"onclick": onclick},
	})
}

// GetKey returns the unique key identifier for this Modal component.
func (e Modal) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this Modal.
func (e Modal) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e Modal) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *Modal) SetChildren(children []PageInterface) {
	e.Children = children
}
