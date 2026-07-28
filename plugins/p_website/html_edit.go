package p_website

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
)

const blankPageStarterHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>New page</title>
</head>
<body>
  <h1>New page</h1>
</body>
</html>
`

// IsEditableHTMLName reports whether a VNode name is editable in the GrapesJS builder.
func IsEditableHTMLName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".html", ".htm", ".tmpl":
		return true
	default:
		return false
	}
}

func editablePageGetter(pageKey string) getters.Getter[any] {
	return getters.Any(getters.Map(getters.Key[p_filesystem.VNode](pageKey), func(_ context.Context, page p_filesystem.VNode) (bool, error) {
		return IsEditableHTMLName(page.Name), nil
	}))
}
