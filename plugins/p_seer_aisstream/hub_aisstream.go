package p_seer_aisstream

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	aisstreamStreamURL  = "wss://stream.aisstream.io/v0/stream"
	subscriptionTimeout = 2 * time.Second
	debounceSub         = 800 * time.Millisecond
	reconnectMin        = 2 * time.Second
	reconnectMax        = 2 * time.Minute
)

type bbox struct {
	LaMin, LoMin, LaMax, LoMax float64
}

type subscriptionMsg struct {
	APIKey             string        `json:"APIKey"`
	BoundingBoxes      [][][]float64 `json:"BoundingBoxes"`
	FilterMessageTypes []string      `json:"FilterMessageTypes"`
}

var (
	hubOnce    sync.Once
	lastBBox   bbox
	bboxMu     sync.Mutex
	debounceT  *time.Timer
	debounceMu sync.Mutex

	hubRunMu  sync.Mutex
	activeC   *websocket.Conn
	hubWriteM sync.Mutex
)

// NotifyViewportFromParams schedules a websocket subscription update
// to match the map viewport (debounced, replaces previous subscription per API).
func NotifyViewportFromParams(lamin, lomin, lamax, lomax float64) {
	bboxMu.Lock()
	lastBBox = bbox{LaMin: lamin, LoMin: lomin, LaMax: lamax, LoMax: lomax}
	bboxMu.Unlock()
	debounceMu.Lock()
	if debounceT != nil {
		debounceT.Stop()
	}
	debounceT = time.AfterFunc(debounceSub, sendSubscriptionToActive)
	debounceMu.Unlock()
	ensureHub()
}

func peekLastBBox() (bbox, bool) {
	bboxMu.Lock()
	defer bboxMu.Unlock()
	if lastBBox == (bbox{}) {
		return lastBBox, false
	}
	return lastBBox, true
}

func sendSubscriptionToActive() {
	b, ok := peekLastBBox()
	if !ok {
		return
	}
	c := getActiveConn()
	if c == nil {
		return
	}
	msg := buildSubscription(b)
	if msg == nil {
		return
	}
	hubWriteM.Lock()
	defer hubWriteM.Unlock()
	if c != getActiveConn() {
		return
	}
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := c.WriteJSON(msg); err != nil {
		slog.Error("p_seer_aisstream: subscription write", "error", err)
		_ = c.Close()
	}
}

func getActiveConn() *websocket.Conn {
	hubRunMu.Lock()
	defer hubRunMu.Unlock()
	return activeC
}

func setActiveConn(c *websocket.Conn) {
	hubRunMu.Lock()
	activeC = c
	hubRunMu.Unlock()
}

func buildSubscription(b bbox) *subscriptionMsg {
	key := strings.TrimSpace(EffectiveAPIKey())
	if key == "" {
		return nil
	}
	box := [][]float64{
		{b.LaMin, b.LoMin},
		{b.LaMax, b.LoMax},
	}
	types := make([]string, len(aisMessageTypes))
	copy(types, aisMessageTypes)
	return &subscriptionMsg{
		APIKey:             key,
		BoundingBoxes:      [][][]float64{box},
		FilterMessageTypes: types,
	}
}

// ensureHub starts the long-lived background loop once an API key may exist
// and clients request viewport updates.
func ensureHub() {
	key := strings.TrimSpace(EffectiveAPIKey())
	if key == "" {
		return
	}
	hubOnce.Do(func() {
		go runHub(context.Background(), key)
	})
}

func runHub(ctx context.Context, apiKey string) {
	_ = apiKey
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	backoff := reconnectMin
	for {
		if ctx.Err() != nil {
			return
		}
		if strings.TrimSpace(EffectiveAPIKey()) == "" {
			time.Sleep(5 * time.Second)
			backoff = reconnectMin
			continue
		}
		if _, has := peekLastBBox(); !has {
			time.Sleep(1 * time.Second)
			backoff = reconnectMin
			continue
		}
		h, r, err := dialer.DialContext(ctx, aisstreamStreamURL, http.Header{})
		if err != nil {
			slog.Error("p_seer_aisstream: dial aisstream", "error", err)
			time.Sleep(backoff)
			if backoff < reconnectMax {
				backoff *= 2
				if backoff > reconnectMax {
					backoff = reconnectMax
				}
			}
			continue
		}
		_ = r.Body.Close()
		backoff = reconnectMin
		setActiveConn(h)
		// First subscription before read loop (3s total window to send).
		if b, ok := peekLastBBox(); ok {
			if msg := buildSubscription(b); msg != nil {
				_ = h.SetWriteDeadline(time.Now().Add(subscriptionTimeout))
				if err := h.WriteJSON(msg); err != nil {
					slog.Error("p_seer_aisstream: first subscription", "error", err)
					_ = h.Close()
					setActiveConn(nil)
					time.Sleep(backoff)
					continue
				}
			}
		}
		readHubLoop(ctx, h)
		_ = h.Close()
		setActiveConn(nil)
		time.Sleep(2 * time.Second)
	}
}

func readHubLoop(_ context.Context, c *websocket.Conn) {
	c.SetReadLimit(1 << 22) // 4MB max message
	for {
		_, b, err := c.ReadMessage()
		if err != nil {
			slog.Debug("p_seer_aisstream: read closed", "error", err)
			return
		}
		applyAISMessage(b)
	}
}
