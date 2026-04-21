package p_seer_opensky

import (
	"testing"
)

func TestParseOpenSkyStatesBody(t *testing.T) {
	raw := `{"time":1700000000,"states":[["abc9f3","UAL123 ", "United States",1700000000,1700000000,-122.3743,37.6188,10000.0,false,220.5,45.0,-1.2,null,null,null,null,null,null]]}`

	got, err := ParseOpenSkyStatesBody([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Time != 1700000000 {
		t.Errorf("time: got %d", got.Time)
	}
	if len(got.Aircraft) != 1 {
		t.Fatalf("aircraft len: got %d", len(got.Aircraft))
	}
	a := got.Aircraft[0]
	if a.Icao24 != "abc9f3" {
		t.Errorf("icao24: got %q", a.Icao24)
	}
	if a.Lat != 37.6188 || a.Lng != -122.3743 {
		t.Errorf("position: got %f,%f", a.Lat, a.Lng)
	}
	if a.OnGround {
		t.Error("onGround want false")
	}
	if a.Velocity == nil || *a.Velocity != 220.5 {
		t.Errorf("velocity: %v", a.Velocity)
	}
	if a.Heading == nil || *a.Heading != 45.0 {
		t.Errorf("heading: %v", a.Heading)
	}
	if a.Altitude == nil || *a.Altitude != 10000.0 {
		t.Errorf("altitude: %v", a.Altitude)
	}
	if a.Callsign == nil || *a.Callsign != "UAL123" {
		t.Errorf("callsign: %v", a.Callsign)
	}
}

func TestParseOpenSkyStatesBodyNullStates(t *testing.T) {
	raw := `{"time":1700000000,"states":null}`
	got, err := ParseOpenSkyStatesBody([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Aircraft) != 0 {
		t.Errorf("want no aircraft, got %d", len(got.Aircraft))
	}
}

func TestParseOpenSkyStatesBodySkipsBadRow(t *testing.T) {
	raw := `{"time":1,"states":[null,[],["only_icao"],["abc","","",null,null,null,null]]}`
	got, err := ParseOpenSkyStatesBody([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Aircraft) != 0 {
		t.Errorf("want 0 aircraft, got %d", len(got.Aircraft))
	}
}
