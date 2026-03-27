package getters

import "strings"

// TitleToFormSlug derives a URL path segment from a form title (lowercase, hyphen-separated,
// ASCII letters and digits only; max 160 runes). Empty or non-slug input becomes "form".
func TitleToFormSlug(title string) string {
	title = strings.TrimSpace(strings.ToLower(title))
	var b strings.Builder
	lastHyphen := false
	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case r == ' ', r == '-', r == '.', r == '_':
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "form"
	}
	runes := []rune(s)
	if len(runes) > 160 {
		s = string(runes[:160])
		s = strings.TrimRight(s, "-")
		if s == "" {
			return "form"
		}
	}
	return s
}

// LabelToHTMLName derives a stable HTML name attribute from a human-readable label
// (lowercase, non-alphanumerics become underscores, collapsed and trimmed).
func LabelToHTMLName(label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastUnderscore = false
		case r == ' ', r == '-', r == '.', r == '_':
			if b.Len() > 0 && !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	s := strings.Trim(b.String(), "_")
	if s == "" {
		return "field"
	}
	return s
}
