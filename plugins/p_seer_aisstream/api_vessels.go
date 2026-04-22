package p_seer_aisstream

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

// VesselsResponse is the JSON for GET /api/vessels/.
type VesselsResponse struct {
	Vessels []Vessel `json:"vessels"`
}

type vesselsErr struct {
	Error string `json:"error"`
}

func parseBboxQuery(q map[string][]string) (lamin, lomin, lamax, lomax float64, ok bool) {
	get := func(k string) (float64, bool) {
		vv := q[k]
		if len(vv) == 0 {
			return 0, false
		}
		f, err := strconv.ParseFloat(vv[0], 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}
	var a, b, c, d bool
	lamin, a = get("lamin")
	lomin, b = get("lomin")
	lamax, c = get("lamax")
	lomax, d = get("lomax")
	if !a || !b || !c || !d {
		return 0, 0, 0, 0, false
	}
	return lamin, lomin, lamax, lomax, true
}

func vesselsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.TrimSpace(EffectiveAPIKey()) == "" {
		// Supplements the global [LoggingLayer] line with a stable `reason=`
		// (see lago/layers.go: slog.Info "http_request").
		slog.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", http.StatusServiceUnavailable),
			slog.String("ip", r.RemoteAddr),
			slog.String("reason", "p_seer_aisstream: api key not configured (set [Plugins.p_seer_aisstream] apiKey in seer.toml)"),
		)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(vesselsErr{Error: "aisstream: api key not configured (set [Plugins.p_seer_aisstream] apiKey in seer.toml)"})
		return
	}
	lamin, lomin, lamax, lomax, ok := parseBboxQuery(r.URL.Query())
	if !ok {
		slog.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", http.StatusBadRequest),
			slog.String("ip", r.RemoteAddr),
			slog.String("reason", "p_seer_aisstream: need lamin, lomin, lamax, lomax query parameters"),
		)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(vesselsErr{Error: "need lamin, lomin, lamax, lomax query parameters"})
		return
	}
	NotifyViewportFromParams(lamin, lomin, lamax, lomax)
	out := VesselsResponse{Vessels: vesselsInBbox(lamin, lomin, lamax, lomax)}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(out); err != nil {
		slog.Error("p_seer_aisstream: encode json", "error", err)
	}
}

func registerVesselsAPIRoute() {
	h := p_users.RequireAuth(http.HandlerFunc(vesselsAPIHandler))
	_ = lago.RegistryRoute.Register("seer_aisstream.VesselsAPIRoute", lago.Route{
		Path:    AppUrl + "api/vessels/",
		Handler: h,
	})
}
