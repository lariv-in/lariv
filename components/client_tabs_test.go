package components

import (
	"context"
	"strings"
	"testing"

	"github.com/lariv-in/lago/getters"
)

func TestClientTabsBuild(t *testing.T) {
	tabs := ClientTabs{
		StateKey: "section",
		Default:  getters.Static("Reports"),
		Tabs: map[string]getters.Getter[PageInterface]{
			"Reports": getters.Static[PageInterface](FieldText{Getter: getters.Static("reports body")}),
			"Intel":   getters.Static[PageInterface](FieldText{Getter: getters.Static("intel body")}),
		},
	}

	html := renderNode(t, tabs.Build(context.Background()))

	for _, want := range []string{
		`x-data="{&#34;section&#34;:&#34;Reports&#34;}"`,
		`md:flex-row`,
		`md:flex-col`,
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
