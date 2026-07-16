// Package components contains explanations and code examples for UI page components in Lariv.
//
// # UI Page Components (components.go)
//
// Page components structure HTML rendering.
//
// # Existing Component Categories
//
//   - Shell Scaffolds (ShellBase, ShellScaffold, ShellTopbarScaffold):
//     Form standard layouts containing sidebars, top headers, and alert notifications.
//   - Form Layouts (Form, MultiStepForm, FormListenBoostedPost):
//     Submit user requests, handling parsing and validation errors.
//   - Form Inputs (InputText, InputTextArea, InputPassword, InputSelect, InputCheckbox, InputForeignKey, etc.):
//     Render interactive fields.
//   - Data Displays (Table, Detail, Accordion, Timeline, MapDisplay):
//     Format structured data, maps, listings, and paginations.
//   - Layout Grid (LayoutCard, LayoutSidebar, ContainerRow, ContainerColumn):
//     Define styles, grid rows, and sidebars.
//   - Action Buttons (ButtonClear, ButtonDownload, ButtonLink, ButtonModal, ButtonSubmit):
//     Execute deletes, post states, or download items.
//   - Utilities (RawString, EscapedString, TemplateComponent, TemplateFSComponent):
//     Direct raw HTML output, escape text strings, or parse template files.
//
// # Creating a Custom Component
//
// Custom components must implement the components.PageInterface:
//
//	package myplugin
//
//	import (
//		"context"
//		"io"
//		"github.com/lariv-in/lariv/components"
//		"maragu.dev/gomponents"
//		html "maragu.dev/gomponents/html"
//	)
//
//	type BadgeComponent struct {
//		components.Page // embeds Key and Roles properties
//		Label           string
//		Color           string // e.g. "red", "green"
//	}
//
//	func (b BadgeComponent) GetKey() string {
//		return b.Key
//	}
//
//	func (b BadgeComponent) GetRoles() []string {
//		return b.Roles
//	}
//
//	func (b BadgeComponent) Build(ctx context.Context) gomponents.Node {
//		return gomponents.NodeFunc(func(w io.Writer) error {
//			return html.Span(
//				html.Class("badge badge-"+b.Color),
//				gomponents.Text(b.Label),
//			).Render(w)
//		})
//	}
package components
