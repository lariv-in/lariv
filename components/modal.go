package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
	UID      string
	// Children represents the slice of sub-components rendered in the modal viewport box.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes  string
}

// Build compiles the Modal component into a Dialog layout wrapper.
func (e Modal) Build(ctx context.Context) Node {
	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, ctx))
	}

	uid := e.UID

	modalContent := Div(Class("mt-8"), Group(childNodes))

	return El("dialog",
		ID(uid), Class("modal modal-open fk-modal-container"),
		Attr("hx-push-url", "false"),
		Attr("hx-target", "this"),
		Attr("hx-swap", "outerHTML"),
		Div(Class("modal-box max-w-4xl bg-base-200 border border-base-300 "+e.Classes),
			FormEl(Method("dialog"),
				Button(Type("button"), Class("btn btn-sm btn-circle btn-outline btn-error absolute right-3 top-3"),
					Attr("onclick", "document.getElementById('"+uid+"').remove()"), Render(Icon{Name: "x-mark"}, ctx),
				),
			),
			modalContent,
		),
		FormEl(Method("dialog"), Class("modal-backdrop"),
			Button(Attr("onclick", "document.getElementById('"+uid+"').remove()"), Text("close")),
		),
	)
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
