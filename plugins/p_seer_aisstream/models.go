package p_seer_aisstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const AISStreamMessagesTable = "seer_aisstream_messages"

type AISStreamMessage struct {
	gorm.Model

	MessageType string     `gorm:"size:96;not null;index"`
	MMSI        string     `gorm:"size:32;index"`
	ShipName    string     `gorm:"size:128;index"`
	ReceivedAt  time.Time  `gorm:"not null;index"`
	TimeUTC     *time.Time `gorm:"index"`

	Position  lago.PGPoint `gorm:"type:point"`
	Longitude float64      `gorm:"-"`
	Latitude  float64      `gorm:"-"`
	SOG       *float64
	COG       *float64
	Heading   *float64

	RawMetadata datatypes.JSON `gorm:"type:jsonb"`
	RawMessage  datatypes.JSON `gorm:"type:jsonb;not null"`
}

func (AISStreamMessage) TableName() string {
	return AISStreamMessagesTable
}

func (m *AISStreamMessage) AfterFind(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if m.Position.Valid {
		m.Longitude = m.Position.P.X
		m.Latitude = m.Position.P.Y
	}
	return nil
}

func (m *AISStreamMessage) BeforeSave(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if aisValidLatLng(m.Latitude, m.Longitude) {
		m.Position = lago.NewPGPoint(m.Longitude, m.Latitude)
	} else {
		m.Position = lago.PGPoint{}
	}
	return nil
}

type AISStreamMessageTypeModel struct {
	gorm.Model

	AISStreamMessageID uint             `gorm:"not null;index"`
	AISStreamMessage   AISStreamMessage `gorm:"constraint:OnDelete:CASCADE;"`
	Payload            datatypes.JSON   `gorm:"type:jsonb;not null"`
}

type aisStreamMessageType struct {
	Model  any
	Save   func(context.Context, *gorm.DB, AISStreamMessage, aisstream.AisStreamMessage) error
	Render func(context.Context, *gorm.DB, uint) (string, error)
}

var AISStreamMessageTypes = registry.NewRegistry[aisStreamMessageType]()

func registerAISStreamMessageType(name string, model any, save func(context.Context, *gorm.DB, AISStreamMessage, aisstream.AisStreamMessage) error) {
	if err := AISStreamMessageTypes.Register(name, aisStreamMessageType{
		Model:  model,
		Save:   save,
		Render: renderTypedPayload(name),
	}); err != nil {
		panic(err)
	}
}

func newPayloadBase(parent AISStreamMessage, payload datatypes.JSON) AISStreamMessageTypeModel {
	return AISStreamMessageTypeModel{
		AISStreamMessageID: parent.ID,
		AISStreamMessage:   parent,
		Payload:            payload,
	}
}

func payloadJSON(packet aisstream.AisStreamMessage, messageType string) (datatypes.JSON, error) {
	raw, err := marshalPostgresJSON(packet.Message)
	if err != nil {
		return nil, err
	}
	var byType map[string]json.RawMessage
	if err := json.Unmarshal(raw, &byType); err != nil {
		return nil, err
	}
	if p, ok := byType[messageType]; ok && len(p) > 0 && string(p) != "null" {
		return datatypes.JSON(p), nil
	}
	return datatypes.JSON(raw), nil
}

func marshalPostgresJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// PostgreSQL jsonb rejects NUL characters, including JSON "\\u0000"
	// escapes. AIS binary payload strings can contain NUL, so preserve them as
	// literal backslash-u text instead of a decoded NUL code point.
	return bytes.ReplaceAll(b, []byte(`\u0000`), []byte(`\\u0000`)), nil
}

func renderTypedPayload(messageType string) func(context.Context, *gorm.DB, uint) (string, error) {
	return func(ctx context.Context, db *gorm.DB, messageID uint) (string, error) {
		if db == nil || messageID == 0 {
			return "", nil
		}
		var row struct {
			Payload datatypes.JSON
		}
		table := typedTableName(messageType)
		if table == "" {
			return "", nil
		}
		err := db.WithContext(ctx).Table(table).
			Select("payload").
			Where("ais_stream_message_id = ?", messageID).
			Order("id DESC").
			Limit(1).
			Scan(&row).Error
		if err != nil {
			return "", err
		}
		if len(row.Payload) == 0 {
			return "", nil
		}
		var v any
		if err := json.Unmarshal(row.Payload, &v); err != nil {
			return string(row.Payload), nil
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return string(row.Payload), nil
		}
		return string(b), nil
	}
}

func typedTableName(messageType string) string {
	switch messageType {
	case string(aisstream.POSITION_REPORT):
		return "seer_aisstream_position_reports"
	case string(aisstream.STANDARD_CLASS_B_POSITION_REPORT):
		return "seer_aisstream_standard_class_b_position_reports"
	case string(aisstream.EXTENDED_CLASS_B_POSITION_REPORT):
		return "seer_aisstream_extended_class_b_position_reports"
	case string(aisstream.LONG_RANGE_AIS_BROADCAST_MESSAGE):
		return "seer_aisstream_long_range_broadcast_messages"
	default:
		return "seer_aisstream_" + snake(messageType)
	}
}

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func aisValidLatLng(lat, lng float64) bool {
	if lat == 0 && lng == 0 {
		return false
	}
	if math.IsNaN(lat) || math.IsNaN(lng) || math.IsInf(lat, 0) || math.IsInf(lng, 0) {
		return false
	}
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func metadataString(meta map[string]interface{}, key string) string {
	v, ok := meta[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func metadataFloat(meta map[string]interface{}, key string) (float64, bool) {
	v, ok := meta[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func parseAISTimeUTC(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02 15:04:05.999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}
