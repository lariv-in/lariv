package p_seer_aisstream

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
)

type aisStreamMapDataHandler struct{}

func (aisStreamMapDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("p_seer_aisstream: map data: db from context", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	vessels, err := buildAISStreamMapVessels(r.Context(), db)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		slog.Error("p_seer_aisstream: map data: load", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if vessels == nil {
		vessels = []aisStreamMapVessel{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vessels); err != nil {
		slog.Debug("p_seer_aisstream: map data: encode", "error", err)
	}
}
