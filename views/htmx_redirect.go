package views

import (
	"fmt"
	"net/http"
)

// HtmxRedirect performs a redirect that is HTMX-aware: for HX-Request it sets
// HX-Redirect and responds with 200; otherwise it behaves like http.Redirect
// with the given status code.
func HtmxRedirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	fmt.Println("HtmxRedirect", url, code, r.Header.Get("HX-Request"), r.Header.Get("HX-Boosted"))
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, code)
}
