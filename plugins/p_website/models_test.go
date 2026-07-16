package p_website

import (
	"testing"
)

func TestDBRouteFields(t *testing.T) {
	route := DBRoute{
		Path:      "/home",
		LTreePath: "home",
		PageID:    1,
		IsActive:  true,
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

	if !route.IsActive {
		t.Errorf("expected IsActive to be true")
	}

	if route.Model != nil {
		t.Errorf("expected Model to be nil")
	}
}
