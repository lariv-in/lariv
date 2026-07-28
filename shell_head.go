package lariv

import (
	"context"
	"io"

	"github.com/lariv-in/lariv/components"
)

type shellHeadTitle struct {
	components.Page
	Title string
}

func (t shellHeadTitle) Build(_ components.Catalog, _ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "<title>"+t.Title+"</title>")
	return err
}
