package lariv

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	pprof_http "net/http/pprof"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// App is a sealed, immutable application catalog produced by [AppBuilder.Build].
// It owns registries, config, DB, and the HTTP handler for one independent app instance.
type App struct {
	Config LarivConfig
	DB     *gorm.DB

	Pages            registry.ImmutableRegistry[components.PageInterface]
	Views            registry.ImmutableRegistry[*views.View]
	Routes           registry.ImmutableRegistry[Route]
	Layers           registry.ImmutableRegistry[views.GlobalLayer]
	Models           registry.ImmutableRegistry[any]
	Configs          registry.ImmutableRegistry[Config]
	Migrations       registry.ImmutableRegistry[UsefulFilesystem]
	DBInitHooks      registry.ImmutableRegistry[DBInitHook]
	Generators       registry.ImmutableRegistry[Generator]
	Commands         registry.ImmutableRegistry[CommandFactory]
	Plugins          registry.ImmutableRegistry[Plugin]
	ShellHead         registry.ImmutableRegistry[components.PageInterface]
	ShellTopbar       registry.ImmutableRegistry[components.PageInterface]
	ShellRightSidebar registry.ImmutableRegistry[components.SidebarItem]
	Admin              registry.ImmutableRegistry[AdminPanelInterface]
	GrapesJSBlocks     registry.ImmutableRegistry[GrapesJSBlock]
	GrapesJSComponents registry.ImmutableRegistry[GrapesJSComponent]
	GrapesJSTraits     registry.ImmutableRegistry[GrapesJSTrait]
	GrapesJSThemes     registry.ImmutableRegistry[GrapesJSTheme]

	handler http.Handler
}

// Page implements [components.Catalog].
func (a *App) Page(name string) (components.PageInterface, bool) {
	return a.Pages.Get(name)
}

// HeadNodes implements [components.Catalog].
func (a *App) HeadNodes() []components.PageInterface {
	pairs := *a.ShellHead.AllStable()
	out := make([]components.PageInterface, len(pairs))
	for i, p := range pairs {
		out[i] = p.Value
	}
	return out
}

// TopbarItems implements [components.Catalog].
func (a *App) TopbarItems() []components.PageInterface {
	pairs := *a.ShellTopbar.AllStable()
	out := make([]components.PageInterface, len(pairs))
	for i, p := range pairs {
		out[i] = p.Value
	}
	return out
}

// RightSidebarItems implements [components.Catalog].
func (a *App) RightSidebarItems() []components.NamedSidebarItem {
	pairs := *a.ShellRightSidebar.AllStable()
	out := make([]components.NamedSidebarItem, len(pairs))
	for i, p := range pairs {
		out[i] = components.NamedSidebarItem{Key: p.Key, Item: p.Value}
	}
	return out
}

// PageView returns a [views.View] that resolves pages from this App's page catalog.
func (a *App) PageView(pageName string) *views.View {
	return &views.View{
		PageName:   pageName,
		PageLookup: a.Pages.Get,
	}
}

// View looks up a named view.
func (a *App) View(name string) (*views.View, bool) {
	return a.Views.Get(name)
}

// Route looks up a named route.
func (a *App) Route(name string) (Route, bool) {
	return a.Routes.Get(name)
}

// Router builds the ServeMux from this App's routes (and optional pprof endpoints).
func (a *App) Router() *http.ServeMux {
	baseRouter := http.NewServeMux()
	if a.Config.Debug {
		baseRouter.HandleFunc("/pprof/", pprof_http.Index)
		fmt.Printf("Added debug route for profile index at /pprof/\n")
		baseRouter.HandleFunc("/pprof/cmdline/", pprof_http.Cmdline)
		fmt.Printf("Added debug route for profile cmdline index at /pprof/cmdline/\n")
		baseRouter.HandleFunc("/pprof/profile/", pprof_http.Profile)
		fmt.Printf("Added debug route for profile 'profile' index at /pprof/profile/\n")
		baseRouter.HandleFunc("/pprof/symbol/", pprof_http.Symbol)
		fmt.Printf("Added debug route for profile symbol index at /pprof/symbol/\n")
		baseRouter.HandleFunc("/pprof/trace/", pprof_http.Trace)
		fmt.Printf("Added debug route for profile trace index at /pprof/trace/\n")
		for _, profile := range pprof.Profiles() {
			profileName := profile.Name()
			profileRoute := fmt.Sprintf("/pprof/%s/", profileName)
			baseRouter.Handle(profileRoute, pprof_http.Handler(profile.Name()))
			fmt.Printf("Added debug route for profile %s at %s\n", profileName, profileRoute)
		}
	}
	for _, route := range a.Routes.All() {
		if strings.HasSuffix(route.Path, "/") {
			baseRouter.Handle(route.Path+"{$}", route.Handler)
		} else {
			baseRouter.Handle(route.Path, route.Handler)
		}
	}
	return baseRouter
}

// Handler returns the fully wrapped HTTP handler (App context, global layers, CORS, router).
func (a *App) Handler() http.Handler {
	if a.handler != nil {
		return a.handler
	}
	var router http.Handler = a.Router()
	for _, layer := range *a.Layers.AllStable() {
		router = layer.Value.Next(router)
	}
	router = http.NewCrossOriginProtection().Handler(router)
	a.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, r.WithContext(ContextWithApp(r.Context(), a)))
	})
	return a.handler
}

// StartServer listens for HTTP requests using this App's handler.
func (a *App) StartServer() error {
	router := a.Handler()

	slog.Warn("Using plain http without tls, ensure this is running in debug or behind a reverse proxy")
	if a.Config.UDS != "" {
		if err := os.Remove(a.Config.UDS); err != nil && !os.IsNotExist(err) {
			return err
		}
		ln, err := net.Listen("unix", a.Config.UDS)
		if err != nil {
			return err
		}
		if err := os.Chmod(a.Config.UDS, 0o777); err != nil {
			return err
		}
		defer ln.Close()
		slog.Info("Listening", "UDS", a.Config.UDS)
		return http.Serve(ln, router)
	}
	slog.Info("Listening", "TCP", a.Config.Address)
	slog.Info("Listening", "http", "http://"+a.Config.Address)
	return http.ListenAndServe(a.Config.Address, router)
}

// Start runs the Cobra CLI (HTTP server root, generate, plugin commands).
func (a *App) Start() error {
	rootCmd := &cobra.Command{
		Use:   "lariv",
		Short: "Lariv web framework",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.StartServer()
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Run data generators to seed the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.RunGenerators()
			return nil
		},
	})

	for _, pair := range *a.Commands.AllStable() {
		rootCmd.AddCommand(pair.Value(a.Config))
	}

	return rootCmd.Execute()
}

// InitDB runs migrations and DB init hooks for this App.
func (a *App) InitDB() error {
	return initDBWith(a.DB, a.Config, &a.Migrations, &a.DBInitHooks)
}

// RunGenerators runs seed generators registered on this App.
func (a *App) RunGenerators() {
	runGeneratorsWith(a.Config, a.DB, a.Generators.All())
}
