// Package pages contains explanations and code examples for pages in Lago.
//
// # Plugin Pages (pages.go)
//
// Pages in Lago represent frontend screens, navigation views, or widgets.
// They must satisfy the components.PageInterface and return a gomponents.Node from Build().
//
// # 1. Creating a Page using Existing Components
//
// You can build pages entirely in Go code using standard layouts and fields provided by lago/components:
//
//	type DashboardPage struct {
//		components.Page
//	}
//
//	func (p *DashboardPage) Build(ctx context.Context) gomponents.Node {
//		return components.Render(components.ShellBase{
//			Children: []components.PageInterface{
//				&components.LayoutCard{
//					Children: []components.PageInterface{
//						&components.EscapedString{Content: "Dashboard Welcome!"},
//					},
//				},
//			},
//		}, ctx)
//	}
//
//	func (p *DashboardPage) GetKey() string     { return p.Key }
//	func (p *DashboardPage) GetRoles() []string { return p.Roles }
//
// # 2. Creating a Page using Embedded Templates
//
// You can use standard Go HTML templates embedded inside your binary utilizing components.TemplateFSComponent:
//
//	//go:embed templates
//	var embeddedFS embed.FS
//
//	type ProfilePage struct {
//		components.Page
//	}
//
//	func (p *ProfilePage) Build(ctx context.Context) gomponents.Node {
//		return components.TemplateFSComponent{
//			Page:             components.Page{Key: p.Key, Roles: p.Roles},
//			Filesystem:       embeddedFS,
//			TemplatePatterns: []string{"templates/*.html"},
//			TemplateName:     "profile.html",
//			TemplateContext: getters.Static[any](map[string]any{
//				"Username": "John",
//			}),
//		}.Build(ctx)
//	}
//
// # 3. Creating a Page using Hardcoded Template Strings
//
// To render simple raw template strings directly without managing files, use components.TemplateComponent:
//
//	var cardTemplate = template.Must(template.New("card").Parse(`
//		<div class="card"><h3>{{.Title}}</h3></div>
//	`))
//
//	type HardcodedPage struct {
//		components.Page
//	}
//
//	func (p *HardcodedPage) Build(ctx context.Context) gomponents.Node {
//		return components.TemplateComponent{
//			Page:         components.Page{Key: p.Key, Roles: p.Roles},
//			Template:     *cardTemplate,
//			TemplateName: "card",
//			TemplateContext: getters.Static[any](map[string]any{
//				"Title": "Raw Template Component",
//			}),
//		}.Build(ctx)
//	}
//
// # 4. Creating a Page using External Template Files
//
// To serve templates from the host filesystem separate from the compiled binary, use os.DirFS:
//
//	type ExternalPage struct {
//		components.Page
//	}
//
//	func (p *ExternalPage) Build(ctx context.Context) gomponents.Node {
//		return components.TemplateFSComponent{
//			Page:             components.Page{Key: p.Key, Roles: p.Roles},
//			Filesystem:       os.DirFS("/opt/app/templates"),
//			TemplatePatterns: []string{"*.html"},
//			TemplateName:     "landing.html",
//			TemplateContext:  nil,
//		}.Build(ctx)
//	}
//
// # 5. Mixing Template-based and Component-based Pages
//
// You can compose layouts where standard Go components and HTML templates live side-by-side:
//
//	type MixedPage struct {
//		components.Page
//	}
//
//	func (p *MixedPage) Build(ctx context.Context) gomponents.Node {
//		return components.Render(components.ShellBase{
//			Children: []components.PageInterface{
//				&components.LayoutCard{
//					Children: []components.PageInterface{
//						&components.EscapedString{Content: "First Card"},
//					},
//				},
//				&components.TemplateComponent{
//					Template:     *cardTemplate,
//					TemplateName: "card",
//					TemplateContext: getters.Static[any](map[string]any{
//						"Title": "Second Card from Template",
//					}),
//				},
//			},
//		}, ctx)
//	}
//
// # Standard html/template Package Reference
//
// Refer to Go standard library [html/template] package.
package pages
