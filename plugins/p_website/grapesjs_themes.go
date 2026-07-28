package p_website

import (
	"context"
	_ "embed"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

//go:embed grapesjs_theme.css
var grapesJSDefaultThemeCSS string

const defaultThemeID = "p_website.default"

const defaultThemeFontsCSS = "https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500;600;700&family=Manrope:wght@400;500;600;700&display=swap"

func pluginGrapesJSThemes() lariv.PluginFeatures[lariv.GrapesJSTheme] {
	return lariv.PluginFeatures[lariv.GrapesJSTheme]{
		Entries: []registry.Pair[string, lariv.GrapesJSTheme]{
			{
				Key: defaultThemeID,
				Value: lariv.GrapesJSTheme{
					Label:       "Default",
					CSS:         strings.TrimSpace(grapesJSDefaultThemeCSS),
					Stylesheets: []string{defaultThemeFontsCSS},
				},
			},
			{
				Key: "p_website.mvp",
				Value: lariv.GrapesJSTheme{
					Label:       "MVP.css",
					Stylesheets: []string{"https://andybrewer.github.io/mvp/mvp.css"},
				},
			},
			{
				Key: "p_website.tacit",
				Value: lariv.GrapesJSTheme{
					Label:       "Tacit",
					Stylesheets: []string{"https://cdn.jsdelivr.net/gh/yegor256/tacit@gh-pages/tacit-css-1.9.7.min.css"},
				},
			},
			{
				Key: "p_website.pico",
				Value: lariv.GrapesJSTheme{
					Label:       "Pico",
					Stylesheets: []string{"https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"},
				},
			},
			{
				Key: "p_website.water",
				Value: lariv.GrapesJSTheme{
					Label:       "Water.css",
					Stylesheets: []string{"https://cdn.jsdelivr.net/npm/water.css@2/out/water.css"},
				},
			},
			{
				Key: "p_website.marx",
				Value: lariv.GrapesJSTheme{
					Label:       "Marx",
					Stylesheets: []string{"https://cdn.jsdelivr.net/npm/marx-css@5/css/marx.min.css"},
				},
			},
		},
	}
}

func grapesJSThemeChoices(ctx context.Context) ([]registry.Pair[string, string], error) {
	app, ok := lariv.AppFromContext(ctx)
	if !ok || app == nil {
		return nil, nil
	}
	pairs := app.GrapesJSThemes.AllStable()
	if pairs == nil {
		return nil, nil
	}
	out := make([]registry.Pair[string, string], 0, len(*pairs))
	for _, p := range *pairs {
		label := p.Value.Label
		if label == "" {
			label = p.Key
		}
		out = append(out, registry.Pair[string, string]{Key: p.Key, Value: label})
	}
	return out, nil
}

func grapesJSThemePairGetter(ctx context.Context) (registry.Pair[string, string], error) {
	key, err := getters.Key[string]("$in.Theme")(ctx)
	if err != nil {
		return registry.Pair[string, string]{}, err
	}
	if key == "" {
		return registry.Pair[string, string]{}, nil
	}
	choices, err := grapesJSThemeChoices(ctx)
	if err != nil {
		return registry.Pair[string, string]{}, err
	}
	if p, ok := registry.PairFromPairs(key, choices); ok {
		return p, nil
	}
	return registry.Pair[string, string]{Key: key, Value: key}, nil
}

func lookupGrapesJSTheme(app *lariv.App, themeID string) (lariv.GrapesJSTheme, bool) {
	if app == nil || themeID == "" {
		return lariv.GrapesJSTheme{}, false
	}
	return app.GrapesJSThemes.Get(themeID)
}
