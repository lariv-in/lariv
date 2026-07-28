package components

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"slices"
	"sort"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
)

type AppsGrid struct {
	components.Page
	Apps getters.Getter[[]lariv.Plugin]
}

func (e AppsGrid) GetKey() string {
	return e.Key
}

func (e AppsGrid) GetRoles() []string {
	return e.Roles
}

type appsGridItem struct {
	Href        string
	VerboseName string
	XShow       string
	Icon        template.HTML
}

func (e AppsGrid) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	var apps []lariv.Plugin
	if e.Apps != nil {
		if appsVal, err := e.Apps(ctx); err == nil {
			apps = appsVal
		}
	}

	if len(apps) == 0 {
		if app, ok := lariv.AppFromContext(ctx); ok && app != nil {
			pluginsMap := app.Plugins.AllStable()
			roleName := p_users.RoleFromContext(ctx, "dashboard.AppsGrid")
			for _, pluginItem := range *pluginsMap {
				plugin := pluginItem.Value
				if plugin.Type == lariv.PluginTypeApp {
					if roleName != "superuser" && len(plugin.Roles) > 0 {
						if !slices.Contains(plugin.Roles, roleName) {
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
	}

	items := make([]appsGridItem, 0, len(apps))
	for _, app := range apps {
		var href string
		if app.URL != nil {
			href = app.URL.String()
		}
		icon, err := components.RenderHTML(components.Icon{Name: app.Icon, Classes: "w-8 h-8"}, cat, ctx)
		if err != nil {
			return err
		}
		items = append(items, appsGridItem{
			Href:        href,
			VerboseName: app.VerboseName,
			XShow:       fmt.Sprintf("'%s'.toLowerCase().includes(search.toLowerCase())", app.VerboseName),
			Icon:        icon,
		})
	}

	return execute(w, "apps_grid", struct {
		Apps []appsGridItem
	}{Apps: items})
}
