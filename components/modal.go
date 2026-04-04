package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// HTMXTargetBodyModal / HTMXSwapBodyModal append modal markup as a direct child of document.body so
// overlays stack above parent stacking contexts (nested modals, selector inside a dialog, etc.).
const (
	HTMXTargetBodyModal = "body"
	HTMXSwapBodyModal   = "beforeend"
)

type Modal struct {
	Page
	UID      string
	Children []PageInterface
	Classes  string
}

func (e Modal) Build(ctx context.Context) Node {
	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, ctx))
	}

	uid := e.UID

	return El("dialog",
		ID(uid), Class("modal modal-open fk-modal-container"),
		Attr("hx-push-url", "false"),
		Attr("hx-target", "this"),
		Attr("hx-swap", "outerHTML"),
		Div(Class("modal-box max-w-4xl "+e.Classes),
			FormEl(Method("dialog"),
				Button(Type("button"), Class("btn btn-sm btn-circle btn-outline btn-error absolute right-3 top-3"),
					Attr("onclick", "document.getElementById('"+uid+"').remove()"), Render(Icon{Name: "x-mark"}, ctx),
				),
			),
			Div(Class("mt-8"), Group(childNodes)),
		),
		FormEl(Method("dialog"), Class("modal-backdrop"),
			Button(Attr("onclick", "document.getElementById('"+uid+"').remove()"), Text("close")),
		),
	)
}

func (e Modal) GetKey() string {
	return e.Key
}

func (e Modal) GetRoles() []string {
	return e.Roles
}

func (e Modal) GetChildren() []PageInterface {
	return e.Children
}

func (e *Modal) SetChildren(children []PageInterface) {
	e.Children = children
}
