package p_google_genai

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenu()
	registerPages()
}

func registerMenu() {
	lago.RegistryPage.Register("googlegenai.Menu", &components.SidebarMenu{
		Title: getters.Static("Google GenAI"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title:  getters.Static("Status"),
				Url:    lago.RoutePath("googlegenai.PageRoute", nil),
				Active: true,
			},
		},
	})
}

func registerPages() {
	lago.RegistryPage.Register("googlegenai.Page", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "googlegenai.Menu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Classes: "max-w-3xl gap-4",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("Google GenAI")},
					&components.FieldSubtitle{Getter: getters.Static("Gemini / Vertex via google.golang.org/genai — same API surface as Local AI plugin.")},
					&components.FieldText{Getter: getters.Static("Model ids, backend, and auth come from [Plugins.p_google_genai]. Uses GenAI client stack aligned with google.golang.org/adk Gemini backends."), Classes: "text-base-content/80"},
					&components.FieldText{Getter: googleGenAIStatusGetter(), Classes: "whitespace-pre-wrap font-mono text-sm text-base-content/80"},
				},
			},
		},
	})
}

func googleGenAIStatusGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		_ = ctx
		return StatusSummary(), nil
	}
}
