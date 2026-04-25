package p_seer_opensky

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	lago.OnDBInit("p_seer_opensky.bootstrap", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[OpenSkyState](db)
		startOpenSkyPollerIfConfigured(db)
		return db
	})
}

func startOpenSkyPollerIfConfigured(db *gorm.DB) {
	if db == nil {
		return
	}
	cfg := Config
	if cfg == nil {
		return
	}
	interval := cfg.PollEvery()
	if interval <= 0 {
		slog.Info("p_seer_opensky: poller off (pollInterval <= 0 or unparseable)")
		return
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		slog.Error("p_seer_opensky: poller not started: clientId and clientSecret required in [Plugins.p_seer_opensky]")
		return
	}

	hc := newFetchHTTPClient()
	tok := newOpenSkyTokenSource(hc, cfg.ClientID, cfg.ClientSecret)
	go runOpenSkyPoller(context.Background(), db, hc, tok, interval)
}

func runOpenSkyPoller(ctx context.Context, db *gorm.DB, hc *http.Client, tok *openSkyTokenSource, interval time.Duration) {
	slog.Info("p_seer_opensky: poller started", "interval", interval.String())
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		if err := ingestOnce(ctx, db, hc, tok); err != nil {
			slog.Error("p_seer_opensky: ingest", "error", err)
		}
		select {
		case <-ctx.Done():
			slog.Info("p_seer_opensky: poller stopped")
			return
		case <-tick.C:
		}
	}
}

func ingestOnce(ctx context.Context, db *gorm.DB, hc *http.Client, tok *openSkyTokenSource) error {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	env, err := FetchAllStatesGET(cctx, hc, tok)
	if err != nil {
		return err
	}
	if len(env.States) == 0 {
		return nil
	}
	const batch = 200
	rows := make([]OpenSkyState, 0, min(len(env.States), batch))
	snapshot := env.Time
	for i := range env.States {
		m, err := ToOpenSkyState(&env.States[i], snapshot)
		if err != nil {
			return err
		}
		if m == nil {
			continue
		}
		rows = append(rows, *m)
		if len(rows) >= batch {
			if err := insertIgnoreConflict(db, rows); err != nil {
				return err
			}
			rows = rows[:0]
		}
	}
	if len(rows) > 0 {
		if err := insertIgnoreConflict(db, rows); err != nil {
			return err
		}
	}
	return nil
}

func insertIgnoreConflict(db *gorm.DB, rows []OpenSkyState) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "icao24"}, {Name: "last_contact"}},
		DoNothing: true,
	}).CreateInBatches(&rows, 100).Error
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
