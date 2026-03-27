package getters

import "testing"

func TestTitleToFormSlug(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  Email Form  ", "email-form"},
		{"Phone #1", "phone-1"},
		{"", "form"},
		{"!!!", "form"},
		{"café", "caf"},
		{"already_slug", "already-slug"},
	}
	for _, tt := range tests {
		if got := TitleToFormSlug(tt.title); got != tt.want {
			t.Errorf("TitleToFormSlug(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestLabelToHTMLName(t *testing.T) {
	tests := []struct {
		label string
		want  string
	}{
		{"Hello World", "hello_world"},
		{"  Email Address  ", "email_address"},
		{"Phone #1", "phone_1"},
		{"", "field"},
		{"!!!", "field"},
		{"café", "caf"},
		{"user_name", "user_name"},
	}
	for _, tt := range tests {
		if got := LabelToHTMLName(tt.label); got != tt.want {
			t.Errorf("LabelToHTMLName(%q) = %q, want %q", tt.label, got, tt.want)
		}
	}
}
