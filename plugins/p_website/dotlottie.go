package p_website

import "strings"

// Pinned @lottiefiles/dotlottie-wc CDN for published pages and the builder canvas.
const (
	dotLottieCDNVersion = "0.9.17"
	dotLottieCDNURL     = "https://unpkg.com/@lottiefiles/dotlottie-wc@" + dotLottieCDNVersion + "/dist/dotlottie-wc.js"
	dotLottieScriptAttr = "data-lariv-dotlottie"
)

func dotLottieScriptTag() string {
	return `<script type="module" src="` + dotLottieCDNURL + `" ` + dotLottieScriptAttr + `></script>`
}

// injectDotLottieScript inserts a single pinned DotLottie CDN module script when the
// HTML contains a <dotlottie-wc> element and the loader script is not already present.
// If no DotLottie element is used, the HTML is returned unchanged.
func injectDotLottieScript(html string) string {
	lower := strings.ToLower(html)
	if !strings.Contains(lower, "dotlottie-wc") {
		return html
	}
	if strings.Contains(lower, dotLottieScriptAttr) {
		return html
	}

	tag := dotLottieScriptTag()
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return html[:idx] + tag + "\n" + html[idx:]
	}
	return html + "\n" + tag + "\n"
}
