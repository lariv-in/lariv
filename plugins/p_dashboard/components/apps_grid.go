package components

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AppsGrid struct {
	components.Page
	Apps getters.Getter
}

func (e AppsGrid) Build(ctx context.Context) Node {
	var apps []lago.Plugin
	if e.Apps != nil {
		if val, ok := e.Apps(ctx).([]lago.Plugin); ok {
			apps = val
		}
	}

	if len(apps) == 0 {
		pluginsMap := lago.RegistryPlugins.AllStable()
		roleName, _ := ctx.Value("$render_key").(string)
		for _, pluginItem := range *pluginsMap {
			plugin := pluginItem.Value

			if plugin.Type == lago.PluginTypeApp {
				if len(plugin.RenderKeys) > 0 {
					if !slices.Contains(plugin.RenderKeys, roleName) {
						continue
					}
				}
				apps = append(apps, plugin)
			}
		}
		sort.Slice(apps, func(i, j int) bool {
			return apps[i].VerboseName < apps[j].VerboseName
		})
	}

	group := Group{}
	for _, app := range apps {
		group = append(group, A(
			Href(app.URL.String()),
			Class("btn btn-md h-auto flex-col space-y-1 py-4"),
			Attr("x-show", fmt.Sprintf("'%s'.toLowerCase().includes(search.toLowerCase())", app.VerboseName)),
			Attr("x-cloak"), components.Render(components.Icon{Name: app.Icon, Classes: "w-8 h-8"}, ctx), Div(
				Class("text-sm truncate min-w-0 w-full"),
				Text(app.VerboseName),
			),
		))

	}
	return Div(Class("container max-w-5xl mx-auto p-6 @container"), Attr("x-data", "{ search: ''}"),
		Div(Class("mb-4"),
			Input(Type("text"), Attr("x-model", "search"), Placeholder("Search apps..."), Class("input input-bordered w-full")),
		),
		Div(Class("grid grid-cols-2 @md:grid-cols-4 @2xl:grid-cols-6 gap-2"), group))
}
