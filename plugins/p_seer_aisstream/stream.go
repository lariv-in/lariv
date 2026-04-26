package p_seer_aisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

func startAISStreamWorkerIfConfigured(db *gorm.DB) {
	if db == nil || Config == nil {
		return
	}
	if !Config.Enabled {
		slog.Info("p_seer_aisstream: stream worker disabled")
		return
	}
	if Config.APIKey == "" {
		slog.Error("p_seer_aisstream: stream worker not started: apiKey required in [Plugins.p_seer_aisstream]")
		return
	}
	go runAISStreamWorker(context.Background(), db, Config)
}

func runAISStreamWorker(ctx context.Context, db *gorm.DB, cfg *AISStreamConfig) {
	backoff := 2 * time.Second
	for {
		err := runAISStreamSession(ctx, db, cfg)
		if err == nil {
			return
		}
		slog.Error("p_seer_aisstream: stream session ended", "error", err, "backoff", backoff.String())
		select {
		case <-ctx.Done():
			slog.Info("p_seer_aisstream: stream worker stopped")
			return
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > time.Minute {
			backoff = time.Minute
		}
	}
}

func runAISStreamSession(ctx context.Context, db *gorm.DB, cfg *AISStreamConfig) error {
	if cfg == nil {
		return nil
	}
	dialer := websocket.Dialer{
		Proxy:            websocket.DefaultDialer.Proxy,
		HandshakeTimeout: 20 * time.Second,
	}
	ws, resp, err := dialer.DialContext(ctx, cfg.StreamURL, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial %s: status %s: %w", cfg.StreamURL, resp.Status, err)
		}
		return fmt.Errorf("dial %s: %w", cfg.StreamURL, err)
	}
	defer ws.Close()

	sub := aisstream.SubscriptionMessage{
		APIKey:        cfg.APIKey,
		BoundingBoxes: [][][]float64{{{-90.0, -180.0}, {90.0, 180.0}}},
	}
	b, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
		return err
	}
	slog.Info("p_seer_aisstream: stream worker connected", "url", cfg.StreamURL)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		var packet aisstream.AisStreamMessage
		if err := json.Unmarshal(p, &packet); err != nil {
			slog.Error("p_seer_aisstream: packet json", "error", err)
			continue
		}
		if err := ingestAISStreamPacket(ctx, db, packet); err != nil {
			slog.Error("p_seer_aisstream: ingest", "error", err, "message_type", packet.MessageType)
		}
	}
}
