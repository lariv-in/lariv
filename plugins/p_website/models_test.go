package p_website

import (
	"testing"

	"github.com/lariv-in/lariv/plugins/p_filesystem"
)

func TestDBRouteFields(t *testing.T) {
	route := DBRoute{
		Path:      "/home",
		LTreePath: "home",
		PageID:    1,
		References: []p_filesystem.VNode{
			{ID: 2, Name: "header.html"},
		},
		IsActive: true,
	}

	if route.Path != "/home" {
		t.Errorf("expected Path to be '/home', got %q", route.Path)
	}

	if route.LTreePath != "home" {
		t.Errorf("expected LTreePath to be 'home', got %q", route.LTreePath)
	}

	if route.PageID != 1 {
		t.Errorf("expected PageID to be 1, got %d", route.PageID)
	}

	if len(route.References) != 1 || route.References[0].ID != 2 {
		t.Errorf("expected 1 reference with ID 2, got %v", route.References)
	}

	if !route.IsActive {
		t.Errorf("expected IsActive to be true")
	}
}
