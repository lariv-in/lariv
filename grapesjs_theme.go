package lariv

// GrapesJSTheme is a named CSS theme for the website page builder and published pages.
// The registry entry Key is the theme id stored on routes (e.g. "p_website.default").
type GrapesJSTheme struct {
	// Label is the human-readable name shown in theme selectors.
	Label string `json:"label"`
	// CSS is the theme stylesheet body injected into the builder canvas and published HTML.
	CSS string `json:"css"`
	// Stylesheets are optional external stylesheet URLs (fonts, etc.) loaded with the theme.
	Stylesheets []string `json:"stylesheets,omitempty"`
}
