package lariv

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
	"gorm.io/gorm"
)

// AppBuilder accumulates plugin contributions and seals them into an [App].
type AppBuilder struct {
	plugins []registry.Pair[string, Plugin]

	// Optional direct chrome contributions (merged after plugins during Build).
	extraHeadNodes    PluginFeatures[components.PageInterface]
	extraTopbar       PluginFeatures[components.PageInterface]
	extraRightSidebar PluginFeatures[components.SidebarItem]
	extraAdmin        PluginFeatures[AdminPanelInterface]
}

// NewBuilder returns an empty [AppBuilder].
func NewBuilder() *AppBuilder {
	return &AppBuilder{}
}

// AddPlugin appends a named plugin contribution.
func (b *AppBuilder) AddPlugin(name string, p Plugin) *AppBuilder {
	b.plugins = append(b.plugins, registry.NewPair(name, p))
	return b
}

// AddPlugins appends multiple named plugin contributions.
func (b *AppBuilder) AddPlugins(plugins []registry.Pair[string, Plugin]) *AppBuilder {
	b.plugins = append(b.plugins, plugins...)
	return b
}

// RegisterHead registers a shell head node on the builder.
func (b *AppBuilder) RegisterHead(name string, node components.PageInterface) *AppBuilder {
	b.extraHeadNodes.Entries = append(b.extraHeadNodes.Entries, registry.NewPair(name, node))
	return b
}

// PatchHead registers a patch for a shell head node.
func (b *AppBuilder) PatchHead(name string, patch func(components.PageInterface) components.PageInterface) *AppBuilder {
	b.extraHeadNodes.Patches = append(b.extraHeadNodes.Patches, registry.NewPair(name, patch))
	return b
}

// RegisterTopbar registers a topbar item on the builder.
func (b *AppBuilder) RegisterTopbar(name string, node components.PageInterface) *AppBuilder {
	b.extraTopbar.Entries = append(b.extraTopbar.Entries, registry.NewPair(name, node))
	return b
}

// RegisterRightSidebar registers a right-sidebar item on the builder.
func (b *AppBuilder) RegisterRightSidebar(name string, item components.SidebarItem) *AppBuilder {
	b.extraRightSidebar.Entries = append(b.extraRightSidebar.Entries, registry.NewPair(name, item))
	return b
}

// RegisterAdmin registers an admin panel on the builder.
func (b *AppBuilder) RegisterAdmin(name string, panel AdminPanelInterface) *AppBuilder {
	b.extraAdmin.Entries = append(b.extraAdmin.Entries, registry.NewPair(name, panel))
	return b
}

// Build seals plugin contributions into an immutable [App].
// If db/config are zero, catalogs are still built (useful for tests); call [AppBuilder.LoadConfigFromFile] for full boot.
func (b *AppBuilder) Build() (*App, error) {
	return b.buildWith(LarivConfig{}, nil, false)
}

// BuildWith seals plugins including [CorePlugin], using the provided config and DB (for tests).
func (b *AppBuilder) BuildWith(config LarivConfig, db *gorm.DB) (*App, error) {
	return b.buildWith(config, db, true)
}

func (b *AppBuilder) buildWith(config LarivConfig, db *gorm.DB, prependCore bool) (*App, error) {
	plugins := b.plugins
	if prependCore {
		if db == nil {
			return nil, fmt.Errorf("lariv: Build with core plugin requires a database")
		}
		plugins = append([]registry.Pair[string, Plugin]{CorePlugin(db, config)}, plugins...)
	}

	app := &App{
		Config: config,
		DB:     db,
	}

	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[UsefulFilesystem] {
		return pair.Value.Migrations
	}), &app.Migrations)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[DBInitHook] {
		return pair.Value.DBInitHooks
	}), &app.DBInitHooks)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Config] {
		return pair.Value.Configs
	}), &app.Configs)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Generator] {
		return pair.Value.Generators
	}), &app.Generators)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[CommandFactory] {
		return pair.Value.CommandFactories
	}), &app.Commands)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[any] {
		return pair.Value.Models
	}), &app.Models)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[views.GlobalLayer] {
		return pair.Value.Layers
	}), &app.Layers)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[components.PageInterface] {
		return pair.Value.Pages
	}), &app.Pages)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[*views.View] {
		return pair.Value.Views
	}), &app.Views)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[Route] {
		return pair.Value.Routes
	}), &app.Routes)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[GrapesJSBlock] {
		return pair.Value.GrapesJSBlocks
	}), &app.GrapesJSBlocks)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[GrapesJSComponent] {
		return pair.Value.GrapesJSComponents
	}), &app.GrapesJSComponents)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[GrapesJSTrait] {
		return pair.Value.GrapesJSTraits
	}), &app.GrapesJSTraits)
	fillAppRegistry(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[GrapesJSTheme] {
		return pair.Value.GrapesJSThemes
	}), &app.GrapesJSThemes)

	headFeatures := mergePluginFeatures(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[components.PageInterface] {
		return pair.Value.HeadNodes
	}))
	headFeatures = headFeatures.Merge(b.extraHeadNodes)
	app.ShellHead = registry.NewImmutableRegistry(headFeatures.Build())

	topbarFeatures := mergePluginFeatures(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[components.PageInterface] {
		return pair.Value.Topbar
	}))
	topbarFeatures = topbarFeatures.Merge(b.extraTopbar)
	app.ShellTopbar = registry.NewImmutableRegistry(topbarFeatures.Build())

	sidebarFeatures := mergePluginFeatures(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[components.SidebarItem] {
		return pair.Value.RightSidebar
	}))
	sidebarFeatures = sidebarFeatures.Merge(b.extraRightSidebar)
	app.ShellRightSidebar = registry.NewImmutableRegistry(sidebarFeatures.Build())

	adminFeatures := mergePluginFeatures(MapSlice(plugins, func(pair registry.Pair[string, Plugin]) []func() PluginFeatures[AdminPanelInterface] {
		return pair.Value.Admin
	}))
	adminFeatures = adminFeatures.Merge(b.extraAdmin)
	app.Admin = registry.NewImmutableRegistry(adminFeatures.Build())

	app.Plugins = registry.NewImmutableRegistry(plugins)

	// Rebind all views to this App's page catalog.
	for _, pair := range *app.Views.AllStable() {
		if pair.Value != nil {
			pair.Value.PageLookup = app.Pages.Get
		}
	}

	return app, nil
}

// LoadConfigFromFile decodes TOML config, opens the DB, seals catalogs (with CorePlugin),
// runs PostConfig hooks, and initializes the database.
func (b *AppBuilder) LoadConfigFromFile(path string) (*App, error) {
	var config LarivConfig

	if path == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	resolvedPath := path
	if !filepath.IsAbs(resolvedPath) {
		if _, err := os.Stat(resolvedPath); err == nil {
			// use cwd path
		} else {
			exe, err := os.Executable()
			if err != nil {
				slog.Error("failed resolving executable path for config file", "err", err, "configPath", path)
				return nil, err
			}
			resolvedPath = filepath.Join(filepath.Dir(exe), resolvedPath)
		}
	}

	md, err := toml.DecodeFile(resolvedPath, &config)
	if err != nil {
		slog.Error("failed decoding config file", "err", err, "configPath", path, "resolvedPath", resolvedPath)
		return nil, err
	}

	db, err := GetDbConn(config)
	if err != nil {
		return nil, err
	}

	app, err := b.buildWith(config, db, true)
	if err != nil {
		return nil, err
	}

	for key, cfgPointer := range app.Configs.All() {
		if prim, ok := config.Plugins[key]; ok {
			if err := md.PrimitiveDecode(prim, cfgPointer); err != nil {
				slog.Error("failed decoding plugin config", "err", err, "plugin", key)
				return nil, err
			}
		}
		cfgPointer.PostConfig()
	}

	if err := app.InitDB(); err != nil {
		return nil, err
	}

	return app, nil
}

func fillAppRegistry[T any](features [][]func() PluginFeatures[T], target *registry.ImmutableRegistry[T]) {
	finalFeatures := mergePluginFeatures(features)
	*target = registry.NewImmutableRegistry(finalFeatures.Build())
}

func mergePluginFeatures[T any](features [][]func() PluginFeatures[T]) PluginFeatures[T] {
	finalFeatures := PluginFeatures[T]{}
	for _, feature := range features {
		if feature == nil {
			continue
		}
		for _, featureFn := range feature {
			finalFeatures = finalFeatures.Merge(featureFn())
		}
	}
	return finalFeatures
}

// MapSlice maps elements in a slice from type T to type R using a converter function.
func MapSlice[T any, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}
