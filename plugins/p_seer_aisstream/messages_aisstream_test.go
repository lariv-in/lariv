package p_seer_aisstream

import (
	"testing"
)

func TestApplyAISMessagePositionReport(t *testing.T) {
	vesselByMMSI.Clear()
	const line = `{
  "MessageType": "PositionReport",
  "Metadata": {},
  "Message": {
    "PositionReport": {
      "Cog": 0,
      "Latitude": 51.44458833333333,
      "Longitude": 3.590816666666667,
      "MessageID": 1,
      "NavigationalStatus": 7,
      "PositionAccuracy": true,
      "Raim": true,
      "RateOfTurn": 0,
      "RepeatIndicator": 0,
      "Sog": 0,
      "Spare": 0,
      "SpecialManoeuvreIndicator": 0,
      "Timestamp": 12,
      "TrueHeading": 17,
      "UserID": 245473000,
      "Valid": true
    }
  }
}`
	applyAISMessage([]byte(line))
	v, ok := vesselByMMSI.Load("245473000")
	if !ok {
		t.Fatal("expected vessel after PositionReport")
	}
	if v.MMSI != "245473000" {
		t.Errorf("mmsi: got %q", v.MMSI)
	}
	if v.Lat < 51.4 || v.Lat > 51.5 {
		t.Errorf("lat: got %v", v.Lat)
	}
	if v.Cog != 17 {
		t.Errorf("cog (true heading): got %v want 17", v.Cog)
	}
}

func TestApplyAISMessageShipStaticData(t *testing.T) {
	vesselByMMSI.Clear()
	const line = `{
  "MessageType": "ShipStaticData",
  "Metadata": {},
  "Message": {
    "ShipStaticData": {
      "AisVersion": 2,
      "CallSign": "LBHF",
      "Destination": "COASTGUARD@@@@@@@@H",
      "MessageID": 5,
      "Name": "KV FARM",
      "RepeatIndicator": 0,
      "UserID": 257069200,
      "Valid": true
    }
  }
}`
	applyAISMessage([]byte(line))
	v, ok := vesselByMMSI.Load("257069200")
	if !ok {
		t.Fatal("expected name-only entry")
	}
	if v.Name != "KV FARM" {
		t.Errorf("name: got %q", v.Name)
	}
}

func TestVesselsInBbox(t *testing.T) {
	vesselByMMSI.Clear()
	vesselByMMSI.Store("1", vesselState{MMSI: "1", Lat: 10, Lng: 20, Cog: 0, Sog: 0, UpdtMs: 1e15})
	out := vesselsInBbox(9, 19, 11, 21)
	if len(out) != 1 || out[0].MMSI != "1" {
		t.Fatalf("got %#v", out)
	}
}
