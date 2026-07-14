package components

import (
	"context"
	"strings"
	"testing"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
)

func TestClientTabsBuild(t *testing.T) {
	tabs := ClientTabs{
		StateKey: "section",
		Default:  getters.Static("Reports"),
		Tabs: []registry.Pair[string, getters.Getter[PageInterface]]{
			{Key: "Reports", Value: getters.Static[PageInterface](FieldText{Getter: getters.Static("reports body")})},
			{Key: "Intel", Value: getters.Static[PageInterface](FieldText{Getter: getters.Static("intel body")})},
		},
	}

	html := renderNode(t, tabs.Build(context.Background()))

	for _, want := range []string{
		`x-data="{&#34;section&#34;:&#34;Reports&#34;}"`,
		`md:flex-col`, // ribbon: row on narrow, stacked tabs from md; content always below ribbon
		`section === &#34;Reports&#34;`,
		`section === &#34;Intel&#34;`,
		`btn-primary`,
		`btn-ghost`,
		`Reports`,
		`Intel`,
		`reports body`,
		`intel body`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in html: %s", want, html)
		}
	}
}

func TestClientTabsLayoutVertical(t *testing.T) {
	tabs := ClientTabs{
		StateKey: "section",
		Layout:   ClientTabsLayoutVertical,
		Tabs: []registry.Pair[string, getters.Getter[PageInterface]]{
			{Key: "A", Value: getters.Static[PageInterface](FieldText{Getter: getters.Static("a")})},
			{Key: "B", Value: getters.Static[PageInterface](FieldText{Getter: getters.Static("b")})},
		},
	}
	html := renderNode(t, tabs.Build(context.Background()))
	if !strings.Contains(html, `w-full flex-col`) {
		t.Fatalf("expected vertical ribbon classes in html: %s", html)
	}
	if strings.Contains(html, `sm:flex-row`) || strings.Contains(html, `md:flex-row`) {
		t.Fatalf("vertical layout should keep content under tabs (no side-by-side flex-row): %s", html)
	}
	if strings.Contains(html, `md:flex-col`) {
		t.Fatalf("vertical layout should not use md:flex-col on ribbon (already column): %s", html)
	}
}

func TestClientTabsLayoutHorizontal(t *testing.T) {
	tabs := ClientTabs{
		StateKey: "section",
		Layout:   ClientTabsLayoutHorizontal,
		Tabs: []registry.Pair[string, getters.Getter[PageInterface]]{
			{Key: "A", Value: getters.Static[PageInterface](FieldText{Getter: getters.Static("a")})},
		},
	}
	html := renderNode(t, tabs.Build(context.Background()))
	if !strings.Contains(html, `flex-wrap`) {
		t.Fatalf("expected horizontal ribbon with flex-wrap: %s", html)
	}
}
