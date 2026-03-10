package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Modal struct {
	UID      string
	Title    string
	Children []PageInterface
	Classes  string
}

func (e Modal) Build(ctx context.Context) Node {
	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, child.Build(ctx))
	}

	uid := e.UID

	var titleNode Node
	if e.Title != "" {
		titleNode = H3(Class("font-bold text-lg mb-4"), Text(e.Title))
	}

	return El("dialog",
		ID(uid), Class("modal modal-open"),
		Attr("hx-push-url", "false"),
		Attr("hx-target", "this"),
		Attr("hx-swap", "outerHTML"),
		Div(Class("modal-box max-w-4xl "+e.Classes),
			FormEl(Method("dialog"),
				Button(Type("button"), Class("btn btn-sm btn-circle btn-ghost absolute right-2 top-2"),
					Attr("onclick", "document.getElementById('"+uid+"').remove()"),
					Icon{Name: "x-mark"}.Build(ctx),
				),
			),
			If(titleNode != nil, titleNode),
			Group(childNodes),
		),
		FormEl(Method("dialog"), Class("modal-backdrop"),
			Button(Attr("onclick", "document.getElementById('"+uid+"').remove()"), Text("close")),
		),
	)
}

func (e Modal) GetChildren() []PageInterface {
	return e.Children
}
