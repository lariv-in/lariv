package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ShellBase struct {
	Page
	Children []PageInterface
	ExtraHead []PageInterface
}

func (e ShellBase) Body(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return Body(Class("hide-right font-sans"), Attr("x-data", `{ theme: localStorage.getItem('theme') || 'light' }`), Attr(":data-theme", "theme"), Attr("hx-ext", "preload, alpine-morph"), Attr("hx-boost", "true"), Attr("hx-indicator", "#global-loading-indicator"), Attr("hx-push-url", "true"),
		Div(ID("global-loading-indicator"), Class("fixed top-0 left-0 w-full z-50"),
			Div(Class("h-0.5 bg-primary animate-pulse")),
		),
		group,
	)
}

func (e ShellBase) Build(ctx context.Context) Node {
	extraHeadGroup := Group{}
	for _, child := range e.ExtraHead {
		extraHeadGroup = append(extraHeadGroup, Render(child, ctx))
	}

	title, titlePresent := ctx.Value("PWA_APP_NAME").(string)
	return HTML(
		Lang("en"),
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			If(titlePresent, Title(title)),
			If(!titlePresent, Title("Lago")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx-ext-ws@2.0.4"), Integrity("sha384-1RwI/nvUSrMRuNj7hX1+27J8XDdCoSLf0EjEyF69nacuWyiJYoQ/j39RT1mSnd2G"), CrossOrigin("anonymous")),
			Script(Src("https://unpkg.com/htmx-ext-alpine-morph@2.0.0/alpine-morph.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/@alpinejs/morph@3.x.x/dist/cdn.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx-ext-preload@2.1.2")),
			Script(Src("https://cdn.jsdelivr.net/npm/apexcharts")),
			Link(Href("https://api.fontshare.com/v2/css?f[]=satoshi@300,400,500,600,700&display=swap"), Rel("stylesheet")),
			Link(Href("https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;500;600;700&display=swap"), Rel("stylesheet")),
			StyleEl(Raw(
				`.heroicon {`+
					`display: inline-block;`+
					`width: 24px;`+
					`height: 24px;`+
					`background-color: currentColor;`+
					`-webkit-mask-image: var(--heroicon-url);`+
					`mask-image: var(--heroicon-url);`+
					`-webkit-mask-repeat: no-repeat;`+
					`mask-repeat: no-repeat;`+
					`-webkit-mask-size: 100% 100%;`+
					`mask-size: 100% 100%;`+
					`}`,
			)),
			Script(Raw(`function toggleTheme() { const d = Alpine.$data(document.body); d.theme = d.theme === 'light' ? 'dark' : 'light'; localStorage.setItem('theme', d.theme); }`)),
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
			extraHeadGroup,
		),
		e.Body(ctx),
	)
}

func (e ShellBase) GetChildren() []PageInterface {
	return e.Children
}
