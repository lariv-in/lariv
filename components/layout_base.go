package components

import (
	"context"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutBase struct {
	Children []PageInterface
}

func (e LayoutBase) Build(ctx context.Context) Node {
	title, titlePresent := ctx.Value("PWA_APP_NAME").(string)
	group := Group{}
	for _, child := range e.Children {
		group = append(group, child.Build(ctx))
	}
	return HTML(
		Lang("en"), Attr("data-theme", "light"),
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			If(titlePresent, Title(title)),
			If(!titlePresent, Title("Lago")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx-ext-ws@2.0.4"), Integrity("sha384-1RwI/nvUSrMRuNj7hX1+27J8XDdCoSLf0EjEyF69nacuWyiJYoQ/j39RT1mSnd2G"), CrossOrigin("anonymous")),
			Script(Src("https://unpkg.com/htmx-ext-alpine-morph@2.0.0/alpine-morph.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/@alpinejs/morph@3.x.x/dist/cdn.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/apexcharts")),
			Link(Href("https://api.fontshare.com/v2/css?f[]=satoshi@300,400,500,600,700&display=swap"), Rel("stylesheet")),
			Link(Href("https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;500;600;700&display=swap"), Rel("stylesheet")),
			Link(
				Rel("stylesheet"),
				Href("https://cdn.jsdelivr.net/npm/heroicons-css@0.1.1/heroicons.min.css"),
				Type("text/css"),
			),
			Script(Src("//unpkg.com/alpinejs"), Defer()),
			Script(Raw("htmx.config.defaultSwapStyle = 'morph'")),
			Link(Href("https://cdn.jsdelivr.net/npm/daisyui@5/daisyui.css"), Rel("stylesheet"), Type("text/css")),
			Script(Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4")),
			StyleEl(Type("text/tailwindcss"), Raw(
				`@theme {`+
					`--font-sans: "Satoshi", ui-sans-serif, system-ui, sans-serif;`+
					`--font-mono: "Roboto Mono", monospace;`+
					`}`+
					`:root {`+
					`font-family: var(--font-sans);`+
					`}`+
					`#global-loading-indicator {`+
					`opacity: 0;`+
					`transition: opacity 200ms ease-in;`+
					`}`+
					`#global-loading-indicator.htmx-request {`+
					`opacity: 1;`+
					`}`,
			)),
		),
		Body(Class("hide-right font-sans"), Attr("hx-indicator", "#global-loading-indicator"), Attr("hx-push-url", "true"), Attr("hx-ext", "alpine-morph"), group),
	)

}

func (e LayoutBase) GetChildren() []PageInterface {
	return e.Children
}
