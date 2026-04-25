package p_seer_opensky

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
)

type openSkyMapDataHandler struct{}

func (openSkyMapDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("p_seer_opensky: map data: db from context", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	aircraft, err := buildOpenSkyMapAircraft(r.Context(), db)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		slog.Error("p_seer_opensky: map data: load", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if aircraft == nil {
		aircraft = []openSkyMapAircraft{}
	}
	if err := json.NewEncoder(w).Encode(aircraft); err != nil {
		slog.Debug("p_seer_opensky: map data: encode", "error", err)
	}
}
