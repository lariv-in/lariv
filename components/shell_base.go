package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ShellBase struct {
	Page
	Children  []PageInterface
	ExtraHead []PageInterface
}

// RegistryShellHeadNodes allows plugins to contribute additional tags to <head>.
// Items are rendered in sorted registry key order.
var RegistryShellHeadNodes = registry.NewRegistry[Node]()

func (e ShellBase) Body(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	globalError, _ := getters.Key[error]("$error._global")(ctx)

	if globalError != nil {
		group = append(group, Div(
			Class("toast toast-bottom toast-center z-50"),
			Div(
				Class("alert alert-error"),
				Text(globalError.Error()),
			),
		))
	}

	return Body(
		Class("hide-right font-sans"),
		Attr("x-data", `{ theme: localStorage.getItem('theme') || 'light' }`),
		Attr(":data-theme", "theme"),
		Attr("hx-ext", "alpine-morph"),
		Attr("hx-boost", "true"),
		Attr("hx-indicator", "#global-loading-indicator"),
		Attr("hx-push-url", "true"),
		Div(ID("global-loading-indicator"), Class("fixed top-0 left-0 w-full z-50"),
			Div(Class("h-0.5 bg-primary animate-pulse")),
		),
		group,
	)
}

func (e ShellBase) Build(ctx context.Context) Node {
	registryHeadGroup := Group{}
	for _, item := range *RegistryShellHeadNodes.AllStable(registry.RegisterOrder[Node]{}) {
		registryHeadGroup = append(registryHeadGroup, item.Value)
	}

	extraHeadGroup := Group{}
	for _, child := range e.ExtraHead {
		extraHeadGroup = append(extraHeadGroup, Render(child, ctx))
	}

	return Doctype(HTML(
		Lang("en"),
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/htmx-ext-ws@2.0.4"), Integrity("sha384-1RwI/nvUSrMRuNj7hX1+27J8XDdCoSLf0EjEyF69nacuWyiJYoQ/j39RT1mSnd2G"), CrossOrigin("anonymous")),
			Script(Src("https://unpkg.com/htmx-ext-alpine-morph@2.0.0/alpine-morph.js")),
			Script(Src("https://cdn.jsdelivr.net/npm/@alpinejs/morph@3.x.x/dist/cdn.min.js")),
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
					`}`+
					`.heroicon-sm {`+
					`width: 16px;`+
					`height: 16px;`+
					`}`+
					`.heroicon-lg {`+
					`width: 32px;`+
					`height: 32px;`+
					`}`,
			)),
			Script(Raw(`function toggleTheme() { const d = Alpine.$data(document.body); d.theme = d.theme === 'light' ? 'dark' : 'light'; localStorage.setItem('theme', d.theme); }`)),
			Script(Src("//unpkg.com/alpinejs"), Defer()),
			Script(Raw(
				`htmx.config.defaultSwapStyle = 'morph';`+
					`htmx.config.responseHandling = [`+
					`{code:"422", swap: true},`+
					`{code:"204", swap: false},`+
					`{code:"[23]..", swap: true},`+
					`{code:"[45]..", swap: false, error: true},`+
					`{code:"...", swap: false}`+
					`];`,
			)),
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
					`[data-theme="dark"] {`+
					`--color-base-100: oklch(14% 0.014 253);`+
					`--color-base-200: oklch(24% 0.014 253);`+
					`--color-base-300: oklch(30% 0.016 252);`+
					`}`+
					`#global-loading-indicator {`+
					`opacity: 0;`+
					`transition: opacity 200ms ease-in;`+
					`}`+
					`#global-loading-indicator.htmx-request {`+
					`opacity: 1;`+
					`}`,
			)),
			registryHeadGroup,
			extraHeadGroup,
		),
		e.Body(ctx),
	),
	)
}

func (e ShellBase) GetKey() string {
	return e.Key
}

func (e ShellBase) GetRoles() []string {
	return e.Roles
}

func (e ShellBase) GetChildren() []PageInterface {
	return e.Children
}

func (e *ShellBase) SetChildren(children []PageInterface) {
	e.Children = children
}
