package components

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParseFormatEnabledTableColumns(t *testing.T) {
	m := ParseEnabledTableColumnsParam(" email , Name ")
	if !m["email"] || !m["Name"] {
		t.Fatalf("got %v", m)
	}
	if got := FormatEnabledTableColumnsQuery(m); got != "Name,email" {
		t.Fatalf("format: %q", got)
	}
	empty := ParseEnabledTableColumnsParam("")
	if len(empty) != 0 {
		t.Fatalf("got %v", empty)
	}
}

func TestTableToggleColumnsURLPreservesParams(t *testing.T) {
	cols := []TableColumn{
		{Name: "a", Label: "A"},
		{Name: "b", Label: "B"},
	}
	u, err := url.Parse("/list?sort=a+ASC&foo=bar")
	if err != nil {
		t.Fatal(err)
	}
	req := &http.Request{URL: u}
	out, err := url.Parse(tableToggleColumnsURL(req, "cols", "a", cols))
	if err != nil {
		t.Fatal(err)
	}
	q := out.Query()
	if q.Get("foo") != "bar" {
		t.Fatalf("foo lost: %v", q)
	}
	if q.Get("sort") != "a ASC" {
		t.Fatalf("sort lost: %v", q)
	}
	if q.Get("page") != "1" {
		t.Fatalf("page: %v", q)
	}
	// toggling a off leaves only b
	if q.Get("cols") != "b" {
		t.Fatalf("cols: %v", q)
	}
}

func TestFilterTableColumnsByEnabledMap(t *testing.T) {
	cols := []TableColumn{
		{Name: "", Label: "Always"},
		{Name: "x", Label: "X"},
		{Name: "y", Label: "Y"},
	}
	m := map[string]bool{"x": true}
	got := FilterTableColumnsByEnabledMap(cols, m)
	if len(got) != 2 || got[0].Label != "Always" || got[1].Name != "x" {
		t.Fatalf("got %#v", got)
	}
}
