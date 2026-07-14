package components

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// MapDisplayWireMaxBytes is the maximum WebSocket binary frame payload size for the
// MapDisplay protocol (gzip(compressed CBOR)); matches RFC frame sizing practice at 1 MiB.
const MapDisplayWireMaxBytes = 1048576

// MapDisplayViewportDecodeMaxBytes caps decompressed CBOR from client viewport messages (gzip bomb mitigation).
const MapDisplayViewportDecodeMaxBytes int64 = 65536

// EncodeMapDisplayWire returns a gzip-compressed CBOR payload suitable as a WebSocket binary frame body.
// It verifies that the resulting payload fits within the protocol maximum size constraint [MapDisplayWireMaxBytes].
func EncodeMapDisplayWire(cborPayload []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(cborPayload); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if len(out) > MapDisplayWireMaxBytes {
		return nil, fmt.Errorf("map display wire exceeds protocol maximum (1048576 bytes): compressed=%d", len(out))
	}
	return out, nil
}

// DecodeMapDisplayWire decompresses one gzip-wrapped payload from the client (viewport updates).
// It enforces safety limits using an io.LimitReader up to maxDecompressed to prevent gzip-bomb exploits.
func DecodeMapDisplayWire(compressed []byte, maxDecompressed int64) ([]byte, error) {
	if len(compressed) > MapDisplayWireMaxBytes {
		return nil, fmt.Errorf("map display wire exceeds protocol maximum (1048576 bytes)")
	}
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	limited := io.LimitReader(zr, maxDecompressed+1)
	out, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(out)) > maxDecompressed {
		return nil, fmt.Errorf("map display viewport payload exceeds maximum (%d bytes)", maxDecompressed)
	}
	return out, nil
}
