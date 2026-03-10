package views

import (
	"log/slog"
	"net/http"

	"github.com/lariv-in/components"
)

type View struct {
	PageName string
	Registry *map[string]components.PageInterface
	Handlers map[string]func(View) http.Handler
}

func (v View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, isHandlerPresent := v.Handlers[r.Method]
	if !isHandlerPresent {
		slog.Error("Handler in view was not found")
		http.NotFound(w, r)
		return
	}
	handler(v).ServeHTTP(w, r)
}

func (v View) GetPage() (components.PageInterface, bool) {
	page, isPagePresent := (*v.Registry)[v.PageName]
	return page, isPagePresent
}

func (v View) RenderPage(w http.ResponseWriter, r *http.Request) {
	page, isPagePresent := v.GetPage()
	if !isPagePresent {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()

	if shell, ok := page.(components.Shell); ok {
		if isBoosted, _ := ctx.Value("isHtmxBoosted").(bool); isBoosted {
			shell.Body(ctx).Render(w)
			return
		}
	}

	page.Build(ctx).Render(w)
}

func (v View) WithMethod(method string, viewHandler func(View) http.Handler) View {
	v.Handlers[method] = viewHandler
	return v
}
