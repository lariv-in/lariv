package p_seer_aisstream

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit("p_seer_aisstream.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[AISStreamMessage](db)
		lago.RegisterModel[AISStreamPositionReport](db)
		lago.RegisterModel[AISStreamStandardClassBPositionReport](db)
		lago.RegisterModel[AISStreamUnknownMessage](db)
		lago.RegisterModel[AISStreamAddressedSafetyMessage](db)
		lago.RegisterModel[AISStreamAddressedBinaryMessage](db)
		lago.RegisterModel[AISStreamAidsToNavigationReport](db)
		lago.RegisterModel[AISStreamAssignedModeCommand](db)
		lago.RegisterModel[AISStreamBaseStationReport](db)
		lago.RegisterModel[AISStreamBinaryAcknowledge](db)
		lago.RegisterModel[AISStreamBinaryBroadcastMessage](db)
		lago.RegisterModel[AISStreamChannelManagement](db)
		lago.RegisterModel[AISStreamCoordinatedUTCInquiry](db)
		lago.RegisterModel[AISStreamDataLinkManagementMessage](db)
		lago.RegisterModel[AISStreamDataLinkManagementMessageData](db)
		lago.RegisterModel[AISStreamExtendedClassBPositionReport](db)
		lago.RegisterModel[AISStreamGnssBroadcastBinaryMessage](db)
		lago.RegisterModel[AISStreamGroupAssignmentCommand](db)
		lago.RegisterModel[AISStreamInterrogation](db)
		lago.RegisterModel[AISStreamLongRangeAisBroadcastMessage](db)
		lago.RegisterModel[AISStreamMultiSlotBinaryMessage](db)
		lago.RegisterModel[AISStreamSafetyBroadcastMessage](db)
		lago.RegisterModel[AISStreamShipStaticData](db)
		lago.RegisterModel[AISStreamSingleSlotBinaryMessage](db)
		lago.RegisterModel[AISStreamStandardSearchAndRescueAircraftReport](db)
		lago.RegisterModel[AISStreamStaticDataReport](db)
		startAISStreamWorkerIfConfigured(db)
		return db
	})
}
