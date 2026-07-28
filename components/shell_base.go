package components

import (
	"bytes"
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
)

// ShellBase represents the global root HTML document scaffold wrapper component.
// It compiles the standard HTML skeleton (doctype, head metadata, responsive viewport limits), imports essential CDN dependencies
// (Tailwind CSS v4, DaisyUI v5, HTMX v2, Alpine.js v3, morph/persist scripts), sets up standard themes (light/dark with toggle functions),
// handles global error toast notifications, and renders child elements inside the body container.
//
// Use Cases:
//   - Serves as the primary root layout container for all full-page views throughout the application.
//
// Example:
//
//	&components.ShellBase{
//	    ExtraHead: []components.PageInterface{
//	        &components.MapDisplayLibreHead{},
//	    },
//	    Children: []components.PageInterface{
//	        &components.LayoutSidebar{...},
//	    },
//	}
type ShellBase struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of main sub-components rendered within the document body.
	Children []PageInterface
	// ExtraHead represents the slice of custom header tags (e.g. metadata, script, link nodes) injected in the HTML head.
	ExtraHead []PageInterface
}

const shellBaseHeroiconCSS = `.heroicon {` +
	`display: inline-block;` +
	`width: 24px;` +
	`height: 24px;` +
	`background-color: currentColor;` +
	`-webkit-mask-image: var(--heroicon-url);` +
	`mask-image: var(--heroicon-url);` +
	`-webkit-mask-repeat: no-repeat;` +
	`mask-repeat: no-repeat;` +
	`-webkit-mask-size: 100% 100%;` +
	`mask-size: 100% 100%;` +
	`}` +
	`.heroicon-sm {` +
	`width: 16px;` +
	`height: 16px;` +
	`}` +
	`.heroicon-lg {` +
	`width: 32px;` +
	`height: 32px;` +
	`}`

const shellBaseToggleThemeJS = `function toggleTheme() { const d = Alpine.$data(document.body); d.theme = d.theme === 'light' ? 'dark' : 'light'; localStorage.setItem('theme', d.theme); }`

const shellBaseHTMXConfigJS = `htmx.config.defaultSwapStyle = 'morph';` +
	`htmx.config.responseHandling = [` +
	`{code:"422", swap: true},` +
	`{code:"204", swap: false},` +
	`{code:"[23]..", swap: true},` +
	`{code:"[45]..", swap: false, error: true},` +
	`{code:"...", swap: false}` +
	`];`

const shellBaseThemeCSS = `@theme {` +
	`--font-sans: "Satoshi", ui-sans-serif, system-ui, sans-serif;` +
	`--font-mono: "Roboto Mono", monospace;` +
	`}` +
	`:root {` +
	`font-family: var(--font-sans);` +
	`}` +
	`[data-theme="dark"] {` +
	`--color-base-100: oklch(14% 0.014 253);` +
	`--color-base-200: oklch(24% 0.014 253);` +
	`--color-base-300: oklch(30% 0.016 252);` +
	`}` +
	`#global-loading-indicator {` +
	`opacity: 0;` +
	`transition: opacity 200ms ease-in;` +
	`}` +
	`#global-loading-indicator.htmx-request {` +
	`opacity: 1;` +
	`}`

// Body compiles the core page content wrapper inside the parent HTML document shell structure, including global indicators and error toasts.
func (e ShellBase) Body(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}

	var globalError string
	if errVal, _ := getters.Key[error]("$error._global")(ctx); errVal != nil {
		globalError = errVal.Error()
	}

	return Execute(w, "shell_base_body", struct {
		Children    template.HTML
		GlobalError string
	}{Children: children, GlobalError: globalError})
}

// Build compiles the ShellBase component into a complete HTML document including stylesheet scripts and tags.
func (e ShellBase) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if cat == nil {
		cat = EmptyCatalog{}
	}
	var registryHead bytes.Buffer
	for _, item := range cat.HeadNodes() {
		if err := Render(item, cat, ctx, &registryHead); err != nil {
			return err
		}
	}

	extraHead, err := RenderChildren(cat, ctx, e.ExtraHead)
	if err != nil {
		return err
	}

	var bodyBuf bytes.Buffer
	if err := e.Body(cat, ctx, &bodyBuf); err != nil {
		return err
	}

	return Execute(w, "shell_base", struct {
		HeroiconCSS    template.CSS
		ToggleThemeJS  template.JS
		HTMXConfigJS   template.JS
		ThemeCSS       template.CSS
		RegistryHead   template.HTML
		ExtraHead      template.HTML
		Body           template.HTML
	}{
		HeroiconCSS:   template.CSS(shellBaseHeroiconCSS),
		ToggleThemeJS: template.JS(shellBaseToggleThemeJS),
		HTMXConfigJS:  template.JS(shellBaseHTMXConfigJS),
		ThemeCSS:      template.CSS(shellBaseThemeCSS),
		RegistryHead:  template.HTML(registryHead.String()),
		ExtraHead:     extraHead,
		Body:          template.HTML(bodyBuf.String()),
	})
}

// GetKey returns the unique key identifier for this ShellBase component.
func (e ShellBase) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShellBase.
func (e ShellBase) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShellBase) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShellBase) SetChildren(children []PageInterface) {
	e.Children = children
}
