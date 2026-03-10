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
		Lang("en"),
		Attr("x-data", `{ theme: localStorage.getItem('theme') || 'light' }`),
		Attr(":data-theme", "theme"),
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			If(titlePresent, Title(title)),
			If(!titlePresent, Title("Lago")),
			Script(Type("module"), Src("https://cdn.jsdelivr.net/npm/@hotwired/turbo@latest/dist/turbo.es2017-esm.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/apexcharts")),
			Link(Href("https://api.fontshare.com/v2/css?f[]=satoshi@300,400,500,600,700&display=swap"), Rel("stylesheet")),
			Link(Href("https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;500;600;700&display=swap"), Rel("stylesheet")),
			StyleEl(Raw(
				`.heroicon {`+
					`display: inline-block;`+
					`width: 1.25em;`+
					`height: 1.25em;`+
					`background-color: currentColor;`+
					`-webkit-mask-image: var(--heroicon-url);`+
					`mask-image: var(--heroicon-url);`+
					`-webkit-mask-repeat: no-repeat;`+
					`mask-repeat: no-repeat;`+
					`-webkit-mask-size: 100% 100%;`+
					`mask-size: 100% 100%;`+
					`}`,
			)),
			Script(Raw(`function toggleTheme() { const d = Alpine.$data(document.documentElement); d.theme = d.theme === 'light' ? 'dark' : 'light'; localStorage.setItem('theme', d.theme); }`)),
			Script(Src("//unpkg.com/alpinejs"), Defer()),

			Link(Href("https://cdn.jsdelivr.net/npm/daisyui@5/daisyui.css"), Rel("stylesheet"), Type("text/css")),
			Script(Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4")),
			StyleEl(Type("text/tailwindcss"), Raw(
				`@theme {`+
					`--font-sans: "Satoshi", ui-sans-serif, system-ui, sans-serif;`+
					`--font-mono: "Roboto Mono", monospace;`+
					`}`+
					`:root {`+
					`font-family: var(--font-sans);`+
					`}`,
			)),
		),
		Body(Class("hide-right font-sans"), group),
	)

}

func (e LayoutBase) GetChildren() []PageInterface {
	return e.Children
}
