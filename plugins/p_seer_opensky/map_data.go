package p_seer_opensky

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/lariv-in/lago/getters"
)

type openSkyMapDisplayVector struct {
	X float64 `json:"x" cbor:"x"`
	Y float64 `json:"y" cbor:"y"`
}

type openSkyMapDisplayPosition struct {
	Lat float64 `json:"lat" cbor:"lat"`
	Lng float64 `json:"lng" cbor:"lng"`
}

type openSkyMapDisplayPoint struct {
	Position  openSkyMapDisplayPosition `json:"position" cbor:"position"`
	Direction openSkyMapDisplayVector   `json:"direction,omitempty" cbor:"direction,omitempty"`
	Velocity  openSkyMapDisplayVector   `json:"velocity,omitempty" cbor:"velocity,omitempty"`
	Time      int64                     `json:"time,omitempty" cbor:"time,omitempty"`
	Link      string                    `json:"link,omitempty" cbor:"link,omitempty"`
}

type openSkyMapViewportMessage struct {
	Type   string                 `json:"type" cbor:"type"`
	Bounds *openSkyViewportBounds `json:"bounds" cbor:"bounds"`
	Zoom   float64                `json:"zoom" cbor:"zoom"`
}

const openSkyViewportMarginDeg = 0.25
const openSkyMaxFrameBytes = 1 << 20

type openSkyMapDataHandler struct{}

func (h openSkyMapDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") ||
		!headerContainsToken(r.Header.Get("Connection"), "upgrade") ||
		key == "" {
		http.Error(w, "bad websocket request", http.StatusBadRequest)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		slog.Error("p_seer_opensky: map websocket hijack failed", "error", err)
		return
	}
	ws := &openSkyMapWebSocketConn{conn: conn, reader: rw.Reader}
	defer ws.close()

	accept := openSkyWebSocketAccept(key)
	if _, err := fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", accept); err != nil {
		slog.Warn("p_seer_opensky: map websocket handshake failed", "error", err)
		return
	}
	if err := rw.Flush(); err != nil {
		slog.Warn("p_seer_opensky: map websocket handshake flush failed", "error", err)
		return
	}

	ctx := r.Context()
	if _, err := getters.DBFromContext(ctx); err != nil {
		slog.Error("p_seer_opensky: map websocket: db from context", "error", err)
		return
	}

	var writeMu sync.Mutex
	if err := sendOpenSkyMapDisplayPoints(ctx, ws, &writeMu, nil); err != nil {
		if !errors.Is(err, ctx.Err()) {
			slog.Warn("p_seer_opensky: map websocket: initial send failed", "error", err)
		}
		return
	}

	stopTicker := make(chan struct{})
	defer close(stopTicker)
	if interval := Config.PollEvery(); interval > 0 {
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-stopTicker:
					return
				case <-ticker.C:
					if err := sendOpenSkyMapDisplayPoints(ctx, ws, &writeMu, nil); err != nil {
						if !errors.Is(err, ctx.Err()) {
							slog.Warn("p_seer_opensky: map websocket: tick send failed", "error", err)
						}
						return
					}
				}
			}
		}()
	}

	for {
		opcode, payload, err := ws.readFrame()
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, ctx.Err()) {
				slog.Debug("p_seer_opensky: map websocket receive closed", "error", err)
			}
			return
		}
		if opcode != 0x1 && opcode != 0x2 {
			continue
		}
		var msg openSkyMapViewportMessage
		if err := cbor.Unmarshal(payload, &msg); err != nil {
			slog.Debug("p_seer_opensky: map websocket ignored malformed message", "error", err)
			continue
		}
		if msg.Type != "mapDisplayViewport" {
			continue
		}
		vp := msg.Bounds
		if vp != nil {
			vp = &openSkyViewportBounds{
				West:  vp.West - openSkyViewportMarginDeg,
				South: vp.South - openSkyViewportMarginDeg,
				East:  vp.East + openSkyViewportMarginDeg,
				North: vp.North + openSkyViewportMarginDeg,
			}
		}
		if err := sendOpenSkyMapDisplayPoints(ctx, ws, &writeMu, vp); err != nil {
			if !errors.Is(err, ctx.Err()) {
				slog.Warn("p_seer_opensky: map websocket: viewport send failed", "error", err)
			}
			return
		}
	}
}

func sendOpenSkyMapDisplayPoints(ctx context.Context, ws *openSkyMapWebSocketConn, writeMu *sync.Mutex, bounds *openSkyViewportBounds) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	aircraft, err := buildOpenSkyMapAircraft(ctx, db, bounds)
	if err != nil {
		return err
	}
	if aircraft == nil {
		aircraft = []openSkyMapAircraft{}
	}
	payload := openSkyMapDisplayPoints(aircraft)
	b, err := cbor.Marshal(payload)
	if err != nil {
		return err
	}
	if len(b) > openSkyMaxFrameBytes {
		return fmt.Errorf("opensky map payload exceeds 1 MiB: bytes=%d points=%d", len(b), len(payload))
	}
	writeMu.Lock()
	defer writeMu.Unlock()
	return ws.writeBinary(b)
}

func headerContainsToken(header, token string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func openSkyWebSocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

type openSkyMapWebSocketConn struct {
	conn   net.Conn
	reader *bufio.Reader
	write  sync.Mutex
}

func (c *openSkyMapWebSocketConn) close() error {
	return c.conn.Close()
}

func (c *openSkyMapWebSocketConn) readFrame() (byte, []byte, error) {
	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, header); err != nil {
			return 0, nil, err
		}
		fin := header[0]&0x80 != 0
		opcode := header[0] & 0x0f
		masked := header[1]&0x80 != 0
		length := uint64(header[1] & 0x7f)

		if !fin {
			return 0, nil, fmt.Errorf("fragmented websocket frames unsupported")
		}
		if !masked {
			return 0, nil, fmt.Errorf("client websocket frame not masked")
		}
		switch length {
		case 126:
			var b [2]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return 0, nil, err
			}
			length = uint64(binary.BigEndian.Uint16(b[:]))
		case 127:
			var b [8]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return 0, nil, err
			}
			length = binary.BigEndian.Uint64(b[:])
		}
		if opcode >= 0x8 && length > 125 {
			return 0, nil, fmt.Errorf("websocket control frame too large")
		}
		if length > 1<<20 {
			return 0, nil, fmt.Errorf("websocket frame exceeds 1 MiB")
		}

		var mask [4]byte
		if _, err := io.ReadFull(c.reader, mask[:]); err != nil {
			return 0, nil, err
		}
		payload := make([]byte, int(length))
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return 0, nil, err
		}
		for i := range payload {
			payload[i] ^= mask[i%4]
		}

		switch opcode {
		case 0x1:
			return opcode, payload, nil
		case 0x2:
			return opcode, payload, nil
		case 0x8:
			return 0, nil, io.EOF
		case 0x9:
			if err := c.writeFrame(0xA, payload); err != nil {
				return 0, nil, err
			}
		case 0xA:
			continue
		case 0x0:
			return 0, nil, fmt.Errorf("websocket continuation frame unsupported")
		default:
			return 0, nil, fmt.Errorf("unsupported websocket opcode %d", opcode)
		}
	}
}

func (c *openSkyMapWebSocketConn) writeBinary(b []byte) error {
	return c.writeFrame(0x2, b)
}

func (c *openSkyMapWebSocketConn) writeFrame(opcode byte, payload []byte) error {
	c.write.Lock()
	defer c.write.Unlock()

	if len(payload) > openSkyMaxFrameBytes {
		return fmt.Errorf("websocket payload exceeds 1 MiB")
	}
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	case len(payload) <= 65535:
		header = append(header, 126, byte(len(payload)>>8), byte(len(payload)))
	default:
		header = append(header, 127)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(len(payload)))
		header = append(header, b[:]...)
	}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}

func openSkyMapDisplayPoints(aircraft []openSkyMapAircraft) []openSkyMapDisplayPoint {
	out := make([]openSkyMapDisplayPoint, 0, len(aircraft))
	for _, a := range aircraft {
		headingRad := a.Heading * math.Pi / 180
		dx := math.Sin(headingRad)
		dy := math.Cos(headingRad)
		vel := openSkyMapDisplayVector{}
		if a.VelocityMps > 0 {
			dE := a.VelocityMps * dx
			dN := a.VelocityMps * dy
			latRad := a.Lat * math.Pi / 180
			vel.X = dE / (111320 * math.Max(0.01, math.Cos(latRad)))
			vel.Y = dN / 111320
		}
		out = append(out, openSkyMapDisplayPoint{
			Position: openSkyMapDisplayPosition{Lat: a.Lat, Lng: a.Lng},
			Direction: openSkyMapDisplayVector{
				X: dx,
				Y: dy,
			},
			Velocity: vel,
			Time:     a.LastContact,
			Link:     a.DetailPath,
		})
	}
	return out
}
