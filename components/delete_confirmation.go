package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DeleteConfirmation struct {
	Title     string
	Message   string
	CancelUrl Getter
	Classes   string
}

func (e DeleteConfirmation) Build(ctx context.Context) Node {
	cancelUrl := fmt.Sprintf("%s", IfOrGetter(e.CancelUrl, ctx, "#"))

	return Div(Class("container mx-auto "+e.Classes),
		H2(Class("text-xl font-bold text-error"), Text(e.Title)),
		P(Class("my-2"), Text(e.Message)),
		FormEl(Class("flex gap-2 my-4"), Method("post"),
			Button(Type("submit"), Class("btn btn-error"), Text("Confirm Delete")),
			A(Href(cancelUrl), Class("btn btn-ghost"), Text("Cancel")),
		),
	)
}
