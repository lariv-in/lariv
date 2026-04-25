package p_seer_aisstream

import (
	"context"
	"fmt"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AISStreamPositionReport struct {
	AISStreamMessageTypeModel

	MessageID                 int32
	RepeatIndicator           int32
	UserID                    int32 `gorm:"index"`
	Valid                     bool
	NavigationalStatus        int32
	RateOfTurn                int32
	SOG                       float64
	PositionAccuracy          bool
	Longitude                 float64
	Latitude                  float64
	COG                       float64
	TrueHeading               int32
	Timestamp                 int32
	SpecialManoeuvreIndicator int32
	Spare                     int32
	RAIM                      bool
	CommunicationState        int32
}

func (AISStreamPositionReport) TableName() string {
	return typedTableName(string(aisstream.POSITION_REPORT))
}

type AISStreamStandardClassBPositionReport struct {
	AISStreamMessageTypeModel

	MessageID                 int32
	RepeatIndicator           int32
	UserID                    int32 `gorm:"index"`
	Valid                     bool
	Spare1                    int32
	SOG                       float64
	PositionAccuracy          bool
	Longitude                 float64
	Latitude                  float64
	COG                       float64
	TrueHeading               int32
	Timestamp                 int32
	Spare2                    int32
	ClassBUnit                bool
	ClassBDisplay             bool
	ClassBDSC                 bool
	ClassBBand                bool
	ClassBMsg22               bool
	AssignedMode              bool
	RAIM                      bool
	CommunicationStateIsITDMA bool
	CommunicationState        int32
}

func (AISStreamStandardClassBPositionReport) TableName() string {
	return typedTableName(string(aisstream.STANDARD_CLASS_B_POSITION_REPORT))
}

type AISStreamPayloadOnly struct {
	AISStreamMessageTypeModel
}

type AISStreamUnknownMessage struct{ AISStreamPayloadOnly }
type AISStreamAddressedSafetyMessage struct{ AISStreamPayloadOnly }
type AISStreamAddressedBinaryMessage struct{ AISStreamPayloadOnly }
type AISStreamAidsToNavigationReport struct{ AISStreamPayloadOnly }
type AISStreamAssignedModeCommand struct{ AISStreamPayloadOnly }
type AISStreamBaseStationReport struct{ AISStreamPayloadOnly }
type AISStreamBinaryAcknowledge struct{ AISStreamPayloadOnly }
type AISStreamBinaryBroadcastMessage struct{ AISStreamPayloadOnly }
type AISStreamChannelManagement struct{ AISStreamPayloadOnly }
type AISStreamCoordinatedUTCInquiry struct{ AISStreamPayloadOnly }
type AISStreamDataLinkManagementMessage struct{ AISStreamPayloadOnly }
type AISStreamDataLinkManagementMessageData struct{ AISStreamPayloadOnly }
type AISStreamExtendedClassBPositionReport struct{ AISStreamPayloadOnly }
type AISStreamGnssBroadcastBinaryMessage struct{ AISStreamPayloadOnly }
type AISStreamGroupAssignmentCommand struct{ AISStreamPayloadOnly }
type AISStreamInterrogation struct{ AISStreamPayloadOnly }
type AISStreamLongRangeAisBroadcastMessage struct{ AISStreamPayloadOnly }
type AISStreamMultiSlotBinaryMessage struct{ AISStreamPayloadOnly }
type AISStreamSafetyBroadcastMessage struct{ AISStreamPayloadOnly }
type AISStreamShipStaticData struct{ AISStreamPayloadOnly }
type AISStreamSingleSlotBinaryMessage struct{ AISStreamPayloadOnly }
type AISStreamStandardSearchAndRescueAircraftReport struct{ AISStreamPayloadOnly }
type AISStreamStaticDataReport struct{ AISStreamPayloadOnly }

func (AISStreamUnknownMessage) TableName() string {
	return typedTableName(string(aisstream.UNKNOWN_MESSAGE))
}
func (AISStreamAddressedSafetyMessage) TableName() string {
	return typedTableName(string(aisstream.ADDRESSED_SAFETY_MESSAGE))
}
func (AISStreamAddressedBinaryMessage) TableName() string {
	return typedTableName(string(aisstream.ADDRESSED_BINARY_MESSAGE))
}
func (AISStreamAidsToNavigationReport) TableName() string {
	return typedTableName(string(aisstream.AIDS_TO_NAVIGATION_REPORT))
}
func (AISStreamAssignedModeCommand) TableName() string {
	return typedTableName(string(aisstream.ASSIGNED_MODE_COMMAND))
}
func (AISStreamBaseStationReport) TableName() string {
	return typedTableName(string(aisstream.BASE_STATION_REPORT))
}
func (AISStreamBinaryAcknowledge) TableName() string {
	return typedTableName(string(aisstream.BINARY_ACKNOWLEDGE))
}
func (AISStreamBinaryBroadcastMessage) TableName() string {
	return typedTableName(string(aisstream.BINARY_BROADCAST_MESSAGE))
}
func (AISStreamChannelManagement) TableName() string {
	return typedTableName(string(aisstream.CHANNEL_MANAGEMENT))
}
func (AISStreamCoordinatedUTCInquiry) TableName() string {
	return typedTableName(string(aisstream.COORDINATED_UTC_INQUIRY))
}
func (AISStreamDataLinkManagementMessage) TableName() string {
	return typedTableName(string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE))
}
func (AISStreamDataLinkManagementMessageData) TableName() string {
	return typedTableName(string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE_DATA))
}
func (AISStreamExtendedClassBPositionReport) TableName() string {
	return typedTableName(string(aisstream.EXTENDED_CLASS_B_POSITION_REPORT))
}
func (AISStreamGnssBroadcastBinaryMessage) TableName() string {
	return typedTableName(string(aisstream.GNSS_BROADCAST_BINARY_MESSAGE))
}
func (AISStreamGroupAssignmentCommand) TableName() string {
	return typedTableName(string(aisstream.GROUP_ASSIGNMENT_COMMAND))
}
func (AISStreamInterrogation) TableName() string {
	return typedTableName(string(aisstream.INTERROGATION))
}
func (AISStreamLongRangeAisBroadcastMessage) TableName() string {
	return typedTableName(string(aisstream.LONG_RANGE_AIS_BROADCAST_MESSAGE))
}
func (AISStreamMultiSlotBinaryMessage) TableName() string {
	return typedTableName(string(aisstream.MULTI_SLOT_BINARY_MESSAGE))
}
func (AISStreamSafetyBroadcastMessage) TableName() string {
	return typedTableName(string(aisstream.SAFETY_BROADCAST_MESSAGE))
}
func (AISStreamShipStaticData) TableName() string {
	return typedTableName(string(aisstream.SHIP_STATIC_DATA))
}
func (AISStreamSingleSlotBinaryMessage) TableName() string {
	return typedTableName(string(aisstream.SINGLE_SLOT_BINARY_MESSAGE))
}
func (AISStreamStandardSearchAndRescueAircraftReport) TableName() string {
	return typedTableName(string(aisstream.STANDARD_SEARCH_AND_RESCUE_AIRCRAFT_REPORT))
}
func (AISStreamStaticDataReport) TableName() string {
	return typedTableName(string(aisstream.STATIC_DATA_REPORT))
}

func savePositionReport(ctx context.Context, db *gorm.DB, parent AISStreamMessage, packet aisstream.AisStreamMessage) error {
	if packet.Message.PositionReport == nil {
		return fmt.Errorf("missing PositionReport payload")
	}
	p := packet.Message.PositionReport
	payload, err := payloadJSON(packet, string(aisstream.POSITION_REPORT))
	if err != nil {
		return err
	}
	row := AISStreamPositionReport{
		AISStreamMessageTypeModel: newPayloadBase(parent, payload),
		MessageID:                 p.MessageID,
		RepeatIndicator:           p.RepeatIndicator,
		UserID:                    p.UserID,
		Valid:                     p.Valid,
		NavigationalStatus:        p.NavigationalStatus,
		RateOfTurn:                p.RateOfTurn,
		SOG:                       p.Sog,
		PositionAccuracy:          p.PositionAccuracy,
		Longitude:                 p.Longitude,
		Latitude:                  p.Latitude,
		COG:                       p.Cog,
		TrueHeading:               p.TrueHeading,
		Timestamp:                 p.Timestamp,
		SpecialManoeuvreIndicator: p.SpecialManoeuvreIndicator,
		Spare:                     p.Spare,
		RAIM:                      p.Raim,
		CommunicationState:        p.CommunicationState,
	}
	return db.WithContext(ctx).Create(&row).Error
}

func saveStandardClassBPositionReport(ctx context.Context, db *gorm.DB, parent AISStreamMessage, packet aisstream.AisStreamMessage) error {
	if packet.Message.StandardClassBPositionReport == nil {
		return fmt.Errorf("missing StandardClassBPositionReport payload")
	}
	p := packet.Message.StandardClassBPositionReport
	payload, err := payloadJSON(packet, string(aisstream.STANDARD_CLASS_B_POSITION_REPORT))
	if err != nil {
		return err
	}
	row := AISStreamStandardClassBPositionReport{
		AISStreamMessageTypeModel: newPayloadBase(parent, payload),
		MessageID:                 p.MessageID,
		RepeatIndicator:           p.RepeatIndicator,
		UserID:                    p.UserID,
		Valid:                     p.Valid,
		Spare1:                    p.Spare1,
		SOG:                       p.Sog,
		PositionAccuracy:          p.PositionAccuracy,
		Longitude:                 p.Longitude,
		Latitude:                  p.Latitude,
		COG:                       p.Cog,
		TrueHeading:               p.TrueHeading,
		Timestamp:                 p.Timestamp,
		Spare2:                    p.Spare2,
		ClassBUnit:                p.ClassBUnit,
		ClassBDisplay:             p.ClassBDisplay,
		ClassBDSC:                 p.ClassBDsc,
		ClassBBand:                p.ClassBBand,
		ClassBMsg22:               p.ClassBMsg22,
		AssignedMode:              p.AssignedMode,
		RAIM:                      p.Raim,
		CommunicationStateIsITDMA: p.CommunicationStateIsItdma,
		CommunicationState:        p.CommunicationState,
	}
	return db.WithContext(ctx).Create(&row).Error
}

func savePayloadOnly(model func(AISStreamMessageTypeModel) any, messageType string) func(context.Context, *gorm.DB, AISStreamMessage, aisstream.AisStreamMessage) error {
	return func(ctx context.Context, db *gorm.DB, parent AISStreamMessage, packet aisstream.AisStreamMessage) error {
		payload, err := payloadJSON(packet, messageType)
		if err != nil {
			return err
		}
		return db.WithContext(ctx).Create(model(newPayloadBase(parent, payload))).Error
	}
}

func init() {
	registerAISStreamMessageType(string(aisstream.POSITION_REPORT), AISStreamPositionReport{}, savePositionReport)
	registerAISStreamMessageType(string(aisstream.STANDARD_CLASS_B_POSITION_REPORT), AISStreamStandardClassBPositionReport{}, saveStandardClassBPositionReport)
	registerAISStreamMessageType(string(aisstream.UNKNOWN_MESSAGE), AISStreamUnknownMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamUnknownMessage{AISStreamPayloadOnly{b}} }, string(aisstream.UNKNOWN_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.ADDRESSED_SAFETY_MESSAGE), AISStreamAddressedSafetyMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamAddressedSafetyMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.ADDRESSED_SAFETY_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.ADDRESSED_BINARY_MESSAGE), AISStreamAddressedBinaryMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamAddressedBinaryMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.ADDRESSED_BINARY_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.AIDS_TO_NAVIGATION_REPORT), AISStreamAidsToNavigationReport{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamAidsToNavigationReport{AISStreamPayloadOnly{b}}
	}, string(aisstream.AIDS_TO_NAVIGATION_REPORT)))
	registerAISStreamMessageType(string(aisstream.ASSIGNED_MODE_COMMAND), AISStreamAssignedModeCommand{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamAssignedModeCommand{AISStreamPayloadOnly{b}} }, string(aisstream.ASSIGNED_MODE_COMMAND)))
	registerAISStreamMessageType(string(aisstream.BASE_STATION_REPORT), AISStreamBaseStationReport{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamBaseStationReport{AISStreamPayloadOnly{b}} }, string(aisstream.BASE_STATION_REPORT)))
	registerAISStreamMessageType(string(aisstream.BINARY_ACKNOWLEDGE), AISStreamBinaryAcknowledge{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamBinaryAcknowledge{AISStreamPayloadOnly{b}} }, string(aisstream.BINARY_ACKNOWLEDGE)))
	registerAISStreamMessageType(string(aisstream.BINARY_BROADCAST_MESSAGE), AISStreamBinaryBroadcastMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamBinaryBroadcastMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.BINARY_BROADCAST_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.CHANNEL_MANAGEMENT), AISStreamChannelManagement{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamChannelManagement{AISStreamPayloadOnly{b}} }, string(aisstream.CHANNEL_MANAGEMENT)))
	registerAISStreamMessageType(string(aisstream.COORDINATED_UTC_INQUIRY), AISStreamCoordinatedUTCInquiry{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamCoordinatedUTCInquiry{AISStreamPayloadOnly{b}} }, string(aisstream.COORDINATED_UTC_INQUIRY)))
	registerAISStreamMessageType(string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE), AISStreamDataLinkManagementMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamDataLinkManagementMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE_DATA), AISStreamDataLinkManagementMessageData{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamDataLinkManagementMessageData{AISStreamPayloadOnly{b}}
	}, string(aisstream.DATA_LINK_MANAGEMENT_MESSAGE_DATA)))
	registerAISStreamMessageType(string(aisstream.EXTENDED_CLASS_B_POSITION_REPORT), AISStreamExtendedClassBPositionReport{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamExtendedClassBPositionReport{AISStreamPayloadOnly{b}}
	}, string(aisstream.EXTENDED_CLASS_B_POSITION_REPORT)))
	registerAISStreamMessageType(string(aisstream.GNSS_BROADCAST_BINARY_MESSAGE), AISStreamGnssBroadcastBinaryMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamGnssBroadcastBinaryMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.GNSS_BROADCAST_BINARY_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.GROUP_ASSIGNMENT_COMMAND), AISStreamGroupAssignmentCommand{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamGroupAssignmentCommand{AISStreamPayloadOnly{b}}
	}, string(aisstream.GROUP_ASSIGNMENT_COMMAND)))
	registerAISStreamMessageType(string(aisstream.INTERROGATION), AISStreamInterrogation{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamInterrogation{AISStreamPayloadOnly{b}} }, string(aisstream.INTERROGATION)))
	registerAISStreamMessageType(string(aisstream.LONG_RANGE_AIS_BROADCAST_MESSAGE), AISStreamLongRangeAisBroadcastMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamLongRangeAisBroadcastMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.LONG_RANGE_AIS_BROADCAST_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.MULTI_SLOT_BINARY_MESSAGE), AISStreamMultiSlotBinaryMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamMultiSlotBinaryMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.MULTI_SLOT_BINARY_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.SAFETY_BROADCAST_MESSAGE), AISStreamSafetyBroadcastMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamSafetyBroadcastMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.SAFETY_BROADCAST_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.SHIP_STATIC_DATA), AISStreamShipStaticData{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamShipStaticData{AISStreamPayloadOnly{b}} }, string(aisstream.SHIP_STATIC_DATA)))
	registerAISStreamMessageType(string(aisstream.SINGLE_SLOT_BINARY_MESSAGE), AISStreamSingleSlotBinaryMessage{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamSingleSlotBinaryMessage{AISStreamPayloadOnly{b}}
	}, string(aisstream.SINGLE_SLOT_BINARY_MESSAGE)))
	registerAISStreamMessageType(string(aisstream.STANDARD_SEARCH_AND_RESCUE_AIRCRAFT_REPORT), AISStreamStandardSearchAndRescueAircraftReport{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any {
		return &AISStreamStandardSearchAndRescueAircraftReport{AISStreamPayloadOnly{b}}
	}, string(aisstream.STANDARD_SEARCH_AND_RESCUE_AIRCRAFT_REPORT)))
	registerAISStreamMessageType(string(aisstream.STATIC_DATA_REPORT), AISStreamStaticDataReport{}, savePayloadOnly(func(b AISStreamMessageTypeModel) any { return &AISStreamStaticDataReport{AISStreamPayloadOnly{b}} }, string(aisstream.STATIC_DATA_REPORT)))
}

func allTypedAISModels() []any {
	return []any{
		AISStreamPositionReport{},
		AISStreamStandardClassBPositionReport{},
		AISStreamUnknownMessage{},
		AISStreamAddressedSafetyMessage{},
		AISStreamAddressedBinaryMessage{},
		AISStreamAidsToNavigationReport{},
		AISStreamAssignedModeCommand{},
		AISStreamBaseStationReport{},
		AISStreamBinaryAcknowledge{},
		AISStreamBinaryBroadcastMessage{},
		AISStreamChannelManagement{},
		AISStreamCoordinatedUTCInquiry{},
		AISStreamDataLinkManagementMessage{},
		AISStreamDataLinkManagementMessageData{},
		AISStreamExtendedClassBPositionReport{},
		AISStreamGnssBroadcastBinaryMessage{},
		AISStreamGroupAssignmentCommand{},
		AISStreamInterrogation{},
		AISStreamLongRangeAisBroadcastMessage{},
		AISStreamMultiSlotBinaryMessage{},
		AISStreamSafetyBroadcastMessage{},
		AISStreamShipStaticData{},
		AISStreamSingleSlotBinaryMessage{},
		AISStreamStandardSearchAndRescueAircraftReport{},
		AISStreamStaticDataReport{},
	}
}

var _ = datatypes.JSON(nil)
