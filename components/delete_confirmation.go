package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DeleteConfirmation struct {
	Page
	Title     string
	Message   string
	CancelUrl getters.Getter[string]
	Classes   string
}

func (e DeleteConfirmation) GetKey() string {
	return e.Key
}

func (e DeleteConfirmation) GetRoles() []string {
	return e.Roles
}

func (e DeleteConfirmation) Build(ctx context.Context) Node {
	cancelUrl := "#"
	if e.CancelUrl != nil {
		url, err := e.CancelUrl(ctx)
		if err != nil {
			slog.Error("DeleteConfirmation CancelUrl getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		cancelUrl = url
	}
	return Div(Class("container mx-auto "+e.Classes),
		H2(Class("text-xl font-bold text-error"), Text(e.Title)),
		P(Class("my-2"), Text(e.Message)),
		FormEl(Class("flex gap-2 my-4"), Method("post"),
			Button(Type("submit"), Class("btn btn-error"), Text("Confirm Delete")),
			A(Href(cancelUrl), Class("btn btn-ghost"), Text("Cancel")),
		),
	)
}
