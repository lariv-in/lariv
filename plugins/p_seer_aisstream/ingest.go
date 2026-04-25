package p_seer_aisstream

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func ingestAISStreamPacket(ctx context.Context, db *gorm.DB, packet aisstream.AisStreamMessage) error {
	if db == nil {
		return nil
	}
	rawMetadata, err := marshalPostgresJSON(packet.MetaData)
	if err != nil {
		return fmt.Errorf("metadata json: %w", err)
	}
	rawMessage, err := marshalPostgresJSON(packet.Message)
	if err != nil {
		return fmt.Errorf("message json: %w", err)
	}
	msg := AISStreamMessage{
		MessageType: string(packet.MessageType),
		ReceivedAt:  time.Now().UTC(),
		RawMetadata: datatypes.JSON(rawMetadata),
		RawMessage:  datatypes.JSON(rawMessage),
	}
	applyEnvelopeFields(&msg, packet)

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Create(&msg).Error; err != nil {
		tx.Rollback()
		return err
	}
	if handler, ok := AISStreamMessageTypes.Get(msg.MessageType); ok && handler.Save != nil {
		if err := handler.Save(ctx, tx, msg, packet); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		slog.Warn("p_seer_aisstream: unknown AIS message type", "message_type", msg.MessageType)
	}
	return tx.Commit().Error
}

func applyEnvelopeFields(msg *AISStreamMessage, packet aisstream.AisStreamMessage) {
	if msg == nil {
		return
	}
	meta := packet.MetaData
	msg.ShipName = metadataString(meta, "ShipName")
	msg.MMSI = metadataString(meta, "MMSI")
	msg.TimeUTC = parseAISTimeUTC(metadataString(meta, "time_utc"))
	if lat, ok := metadataFloat(meta, "latitude"); ok {
		msg.Latitude = lat
	}
	if lng, ok := metadataFloat(meta, "longitude"); ok {
		msg.Longitude = lng
	}

	switch packet.MessageType {
	case aisstream.POSITION_REPORT:
		if p := packet.Message.PositionReport; p != nil {
			msg.MMSI = firstNonEmpty(msg.MMSI, strconv.FormatInt(int64(p.UserID), 10))
			msg.Latitude = p.Latitude
			msg.Longitude = p.Longitude
			msg.SOG = &p.Sog
			msg.COG = &p.Cog
			heading := float64(p.TrueHeading)
			if p.TrueHeading == 511 {
				heading = p.Cog
			}
			msg.Heading = &heading
		}
	case aisstream.STANDARD_CLASS_B_POSITION_REPORT:
		if p := packet.Message.StandardClassBPositionReport; p != nil {
			msg.MMSI = firstNonEmpty(msg.MMSI, strconv.FormatInt(int64(p.UserID), 10))
			msg.Latitude = p.Latitude
			msg.Longitude = p.Longitude
			msg.SOG = &p.Sog
			msg.COG = &p.Cog
			heading := float64(p.TrueHeading)
			if p.TrueHeading == 511 {
				heading = p.Cog
			}
			msg.Heading = &heading
		}
	case aisstream.SHIP_STATIC_DATA:
		if p := packet.Message.ShipStaticData; p != nil {
			msg.MMSI = firstNonEmpty(msg.MMSI, strconv.FormatInt(int64(p.UserID), 10))
			msg.ShipName = firstNonEmpty(msg.ShipName, strings.TrimSpace(p.Name))
		}
	case aisstream.BASE_STATION_REPORT:
		if p := packet.Message.BaseStationReport; p != nil {
			msg.MMSI = firstNonEmpty(msg.MMSI, strconv.FormatInt(int64(p.UserID), 10))
			msg.Latitude = p.Latitude
			msg.Longitude = p.Longitude
		}
	case aisstream.STANDARD_SEARCH_AND_RESCUE_AIRCRAFT_REPORT:
		if p := packet.Message.StandardSearchAndRescueAircraftReport; p != nil {
			msg.MMSI = firstNonEmpty(msg.MMSI, strconv.FormatInt(int64(p.UserID), 10))
			msg.Latitude = p.Latitude
			msg.Longitude = p.Longitude
			msg.SOG = &p.Sog
			msg.COG = &p.Cog
			msg.Heading = &p.Cog
		}
	}
}

func firstNonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
