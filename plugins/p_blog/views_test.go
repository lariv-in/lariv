package p_blog

import (
	"testing"
)

func TestGetPluginRegistration(t *testing.T) {
	pluginPair := GetPlugin()

	if pluginPair.Key != "p_blog" {
		t.Errorf("expected plugin Key to be 'p_blog', got %q", pluginPair.Key)
	}

	plugin := pluginPair.Value
	if plugin.VerboseName != "Blog" {
		t.Errorf("expected VerboseName 'Blog', got %q", plugin.VerboseName)
	}

	if len(plugin.Models) == 0 {
		t.Error("expected Models stages to be registered")
	}

	if len(plugin.Migrations) == 0 {
		t.Error("expected Migrations stages to be registered")
	}

	if len(plugin.Views) == 0 {
		t.Error("expected Views stages to be registered")
	}

	if len(plugin.Pages) == 0 {
		t.Error("expected Pages stages to be registered")
	}

	if len(plugin.Routes) == 0 {
		t.Error("expected Routes stages to be registered")
	}

	// Verify pluginModels entries
	modelsFeatures := pluginModels()
	if len(modelsFeatures.Entries) != 2 {
		t.Errorf("expected 2 model entries, got %d", len(modelsFeatures.Entries))
	}

	// Verify pluginRoutes entries
	routesFeatures := pluginRoutes()
	if len(routesFeatures.Entries) != 11 {
		t.Errorf("expected 11 route entries, got %d", len(routesFeatures.Entries))
	}

	// Verify pluginViews entries
	viewsFeatures := pluginViews()
	if len(viewsFeatures.Entries) != 11 {
		t.Errorf("expected 11 view entries, got %d", len(viewsFeatures.Entries))
	}
}
