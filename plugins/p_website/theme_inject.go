package p_website

import (
	"html"
	"regexp"
	"strings"

	"github.com/lariv-in/lariv"
)

var (
	reLarivThemeStyle = regexp.MustCompile(`(?is)<style\b[^>]*\bdata-lariv-theme\b[^>]*>.*?</style>\s*`)
	reLarivThemeLink  = regexp.MustCompile(`(?is)<link\b[^>]*\bdata-lariv-theme\b[^>]*>\s*`)
)

// themeHeadHTML returns link + style tags for a theme, marked with data-lariv-theme.
func themeHeadHTML(themeID string, theme lariv.GrapesJSTheme) string {
	if themeID == "" {
		return ""
	}
	var b strings.Builder
	escapedID := html.EscapeString(themeID)
	for _, href := range theme.Stylesheets {
		href = strings.TrimSpace(href)
		if href == "" {
			continue
		}
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`" data-lariv-theme="`)
		b.WriteString(escapedID)
		b.WriteString(`">` + "\n")
	}
	if css := strings.TrimSpace(theme.CSS); css != "" {
		b.WriteString(`<style data-lariv-theme="`)
		b.WriteString(escapedID)
		b.WriteString(`">` + "\n")
		b.WriteString(css)
		b.WriteString("\n</style>\n")
	}
	return b.String()
}

// stripLarivThemeAssets removes previously injected theme link/style tags.
func stripLarivThemeAssets(htmlDoc string) string {
	htmlDoc = reLarivThemeStyle.ReplaceAllString(htmlDoc, "")
	htmlDoc = reLarivThemeLink.ReplaceAllString(htmlDoc, "")
	return htmlDoc
}

// injectThemeAssets inserts (or replaces) theme stylesheets and CSS into an HTML document.
// Empty themeID strips any previous theme assets and returns the document unchanged otherwise.
func injectThemeAssets(htmlDoc string, themeID string, theme lariv.GrapesJSTheme) string {
	htmlDoc = stripLarivThemeAssets(htmlDoc)
	block := themeHeadHTML(themeID, theme)
	if block == "" {
		return htmlDoc
	}

	lower := strings.ToLower(htmlDoc)
	if idx := strings.Index(lower, "</head>"); idx >= 0 {
		return htmlDoc[:idx] + block + htmlDoc[idx:]
	}
	if idx := strings.Index(lower, "<body"); idx >= 0 {
		return htmlDoc[:idx] + "<head>\n" + block + "</head>\n" + htmlDoc[idx:]
	}
	if strings.Contains(lower, "<html") {
		return block + htmlDoc
	}
	return "<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n" +
		block +
		"</head>\n<body>\n" + htmlDoc + "\n</body>\n</html>\n"
}
