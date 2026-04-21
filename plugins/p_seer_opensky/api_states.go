package p_seer_opensky

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func statesAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), upstreamTimeout)
	defer cancel()

	out, status, hdr, err := FetchStates(ctx, r.URL.RawQuery)
	if err != nil {
		slog.Error("p_seer_opensky: fetch states", "error", err)
		if status == http.StatusTooManyRequests {
			if ra := hdr.Get("X-Rate-Limit-Retry-After-Seconds"); ra != "" {
				w.Header().Set("Retry-After", ra)
			}
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		if status >= 400 && status < 600 {
			http.Error(w, http.StatusText(status), status)
			return
		}
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if ra := hdr.Get("X-Rate-Limit-Retry-After-Seconds"); ra != "" {
		w.Header().Set("X-Rate-Limit-Retry-After-Seconds", ra)
	}
	if rem := hdr.Get("X-Rate-Limit-Remaining"); rem != "" {
		w.Header().Set("X-Rate-Limit-Remaining", rem)
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(out); err != nil {
		slog.Error("p_seer_opensky: encode json", "error", err)
	}
}

func registerStatesAPIRoute() {
	h := p_users.RequireAuth(http.HandlerFunc(statesAPIHandler))
	_ = lago.RegistryRoute.Register("seer_opensky.StatesAPIRoute", lago.Route{
		Path:    AppUrl + "api/states/",
		Handler: h,
	})
}
