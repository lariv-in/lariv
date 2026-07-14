// Package quickstart guides you through building a minimal Lago plugin that renders "Hello, World!".
//
// # Creating a Hello World Plugin
//
// Follow this step-by-step tutorial to define a plugin, route, view, page component, and bootstrap the server.
//
// # Step 1: Create the Plugin Entrypoint (app.go)
//
// Every plugin must define a key, type, and verbose name. If the plugin type is PluginTypeApp (a standalone application), it also specifies a landing URL and dashboard icon. Start by creating a minimal app.go file:
//
//	package myplugin
//
//	import (
//		"net/url"
//
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func GetPlugin() registry.Pair[string, lago.Plugin] {
//		u, _ := url.Parse("/hello/")
//		return registry.Pair[string, lago.Plugin]{
//			Key: "myplugin",
//			Value: lago.Plugin{
//				Type:        lago.PluginTypeApp,
//				VerboseName: "Hello Plugin",
//				Icon:        "sparkles",
//				URL:         u,
//			},
//		}
//	}
//
// # Step 2: Add HTTP Routing (routes.go)
//
// Define the path routes supported by your plugin. Create a routes.go file:
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func pluginRoutes() lago.PluginFeatures[lago.Route] {
//		return lago.PluginFeatures[lago.Route]{
//			Entries: []registry.Pair[string, lago.Route]{
//				{
//					Key: "myplugin.hello_route",
//					Value: lago.Route{
//						Path:    "/hello/",
//						Handler: lago.NewDynamicView("myplugin.hello_view"),
//					},
//				},
//			},
//		}
//	}
//
// Now, update your app.go file to register the routes feature stage:
//
//	package myplugin
//
//	import (
//		"net/url"
//
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func GetPlugin() registry.Pair[string, lago.Plugin] {
//		u, _ := url.Parse("/hello/")
//		return registry.Pair[string, lago.Plugin]{
//			Key: "myplugin",
//			Value: lago.Plugin{
//				Type:        lago.PluginTypeApp,
//				VerboseName: "Hello Plugin",
//				Icon:        "sparkles",
//				URL:         u,
//				Routes:      lago.PluginStages(pluginRoutes),
//			},
//		}
//	}
//
// # Step 3: Add the View Controller (views.go)
//
// Views act as controllers that link route paths to target pages. Create a views.go file:
//
//	package myplugin
//
//	import (
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//		"github.com/lariv-in/lago/views"
//	)
//
//	func pluginViews() lago.PluginFeatures[*views.View] {
//		return lago.PluginFeatures[*views.View]{
//			Entries: []registry.Pair[string, *views.View]{
//				{
//					Key:   "myplugin.hello_view",
//					Value: lago.GetPageView("myplugin.hello_page"),
//				},
//			},
//		}
//	}
//
// Update your app.go file to register both routes and views feature stages:
//
//	package myplugin
//
//	import (
//		"net/url"
//
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func GetPlugin() registry.Pair[string, lago.Plugin] {
//		u, _ := url.Parse("/hello/")
//		return registry.Pair[string, lago.Plugin]{
//			Key: "myplugin",
//			Value: lago.Plugin{
//				Type:        lago.PluginTypeApp,
//				VerboseName: "Hello Plugin",
//				Icon:        "sparkles",
//				URL:         u,
//				Routes:      lago.PluginStages(pluginRoutes),
//				Views:       lago.PluginStages(pluginViews),
//			},
//		}
//	}
//
// # Step 4: Create the Page Layout (pages.go)
//
// Pages render the final HTML output. Define a struct implementing components.PageInterface. Create a pages.go file:
//
//	package myplugin
//
//	import (
//		"context"
//
//		"github.com/lariv-in/lago/components"
//		"github.com/lariv-in/lago/registry"
//		"maragu.dev/gomponents"
//		"maragu.dev/gomponents/html"
//	)
//
//	type HelloPage struct {
//		components.Page // Embeds Key and Roles field helpers
//	}
//
//	func (p HelloPage) Build(ctx context.Context) gomponents.Node {
//		return html.Div(
//			html.H1(gomponents.Text("Hello, World!")),
//		)
//	}
//
//	func pluginPages() lago.PluginFeatures[components.PageInterface] {
//		return lago.PluginFeatures[components.PageInterface]{
//			Entries: []registry.Pair[string, components.PageInterface]{
//				{
//					Key:   "myplugin.hello_page",
//					Value: HelloPage{},
//				},
//			},
//		}
//	}
//
// Finally, update your app.go file to register pages, routes, and views feature stages:
//
//	package myplugin
//
//	import (
//		"net/url"
//
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//	)
//
//	func GetPlugin() registry.Pair[string, lago.Plugin] {
//		u, _ := url.Parse("/hello/")
//		return registry.Pair[string, lago.Plugin]{
//			Key: "myplugin",
//			Value: lago.Plugin{
//				Type:        lago.PluginTypeApp,
//				VerboseName: "Hello Plugin",
//				Icon:        "sparkles",
//				URL:         u,
//				Pages:       lago.PluginStages(pluginPages),
//				Routes:      lago.PluginStages(pluginRoutes),
//				Views:       lago.PluginStages(pluginViews),
//			},
//		}
//	}
//
// # Step 5: Bootstrap the Server (main.go)
//
// Load your plugin list inside main.go and bootstrap the kernel server engine:
//
//	package main
//
//	import (
//		"log"
//
//		"github.com/lariv-in/lago"
//		"github.com/lariv-in/lago/registry"
//		"myproject/myplugin" // import path to your new plugin
//	)
//
//	func main() {
//		plugins := []registry.Pair[string, lago.Plugin]{
//			myplugin.GetPlugin(),
//		}
//
//		config, err := lago.LoadConfigFromFile("config.toml", plugins)
//		if err != nil {
//			log.Fatalf("failed to load configuration: %v", err)
//		}
//
//		if err := lago.Start(config, plugins); err != nil {
//			log.Fatalf("failed to start server: %v", err)
//		}
//	}
//
// Once the server starts, it will print the local server URL to your console. Open that address in your browser (e.g. http://localhost:8080/hello/) to view the Hello World page.
//
// # Next Steps
//
// For a detailed breakdown of the application file structure, standard plugin files (app.go, config.go, pages.go, migrations.go, routes.go, models.go, views.go, commands.go), and architectural concepts (layers.go, components.go, querypatchers.go), refer to the documentation package: [github.com/lariv-in/lago/docs].
package quickstart
