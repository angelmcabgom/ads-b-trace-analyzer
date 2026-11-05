package trace

import (
	"encoding/json"
	"os"
	"testing"
)

// First point of mock/trace-landing.mock.json (A320 at cruise).
func TestUnmarshalTracePointSpeedAndTrack(t *testing.T) {
	raw := `[0.0, 43.956206, -1.457716, 34000, 472.3, 26.1, 0, -64]`

	var tp TracePoint
	if err := json.Unmarshal([]byte(raw), &tp); err != nil {
		t.Fatal(err)
	}

	if tp.GroundSpeed != 472.3 {
		t.Errorf("GroundSpeed = %v, want 472.3 (readsb index 4)", tp.GroundSpeed)
	}
	if tp.Track != 26.1 {
		t.Errorf("Track = %v, want 26.1 (readsb index 5)", tp.Track)
	}
}

// Sanity check across the real mocks: track is a heading in degrees,
// and some points must be faster than any valid heading could be.
func TestMocksSpeedAndTrackRanges(t *testing.T) {
	for _, f := range []string{
		"../mock/trace.mock.json",
		"../mock/trace-big.mock.json",
		"../mock/trace-landing.mock.json",
	} {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var pt PlaneTrace
		if err := json.Unmarshal(data, &pt); err != nil {
			t.Fatalf("%s: %v", f, err)
		}

		maxSpeed := 0.0
		for i, p := range pt.Trace {
			if p.Track < 0 || p.Track >= 360 {
				t.Errorf("%s point %d: Track = %v, want [0, 360)", f, i, p.Track)
			}
			if p.GroundSpeed > maxSpeed {
				maxSpeed = p.GroundSpeed
			}
		}
		if pt.T != "C172" && maxSpeed < 360 {
			t.Errorf("%s (%s): max GroundSpeed = %v kt, expected a jet to exceed 360", f, pt.T, maxSpeed)
		}
	}
}

// Airborne point from mock/trace-big.mock.json (B772 climbing through FL300).
func TestUnmarshalTracePointAirspeedAndRoll(t *testing.T) {
	raw := `[12995.26, 50.802317, -2.895203, 30125, 413.3, 267.9, 0, 1088, null, "adsb_icao", 31325, 1056, 297, -0.9]`

	var tp TracePoint
	if err := json.Unmarshal([]byte(raw), &tp); err != nil {
		t.Fatal(err)
	}

	if tp.IndicatedAirspeed == nil || *tp.IndicatedAirspeed != 297 {
		t.Errorf("IndicatedAirspeed = %v, want 297 (readsb index 12)", tp.IndicatedAirspeed)
	}
	if tp.Roll == nil || *tp.Roll != -0.9 {
		t.Errorf("Roll = %v, want -0.9 (readsb index 13)", tp.Roll)
	}
}

// Points without air data (the C172 mock never has any) leave both fields nil.
func TestUnmarshalTracePointNullAirspeedAndRoll(t *testing.T) {
	raw := `[0.0, 51.154764, -0.189495, "ground", 0.1, null, 1, null, null, "adsb_icao", null, null, null, null]`

	var tp TracePoint
	if err := json.Unmarshal([]byte(raw), &tp); err != nil {
		t.Fatal(err)
	}

	if tp.IndicatedAirspeed != nil {
		t.Errorf("IndicatedAirspeed = %v, want nil", *tp.IndicatedAirspeed)
	}
	if tp.Roll != nil {
		t.Errorf("Roll = %v, want nil", *tp.Roll)
	}
}

// Cross-check against the real mocks: when readsb attaches the aircraft
// details object (index 8) with an "ias" key, index 12 must carry the same
// value. A heading would only match by coincidence.
func TestMocksIndicatedAirspeedMatchesDetails(t *testing.T) {
	for _, f := range []string{
		"../mock/trace.mock.json",
		"../mock/trace-big.mock.json",
		"../mock/trace-landing.mock.json",
	} {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}

		var pt PlaneTrace
		if err := json.Unmarshal(data, &pt); err != nil {
			t.Fatalf("%s: %v", f, err)
		}

		// Decode the details object separately: TraceMeta has no ias field.
		var raw struct {
			Trace [][]json.RawMessage `json:"trace"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s: %v", f, err)
		}

		checked := 0
		for i, p := range pt.Trace {
			var details struct {
				IAS *int `json:"ias"`
			}
			if len(raw.Trace[i]) <= 8 || json.Unmarshal(raw.Trace[i][8], &details) != nil || details.IAS == nil {
				continue
			}
			checked++
			if p.IndicatedAirspeed == nil || *p.IndicatedAirspeed != *details.IAS {
				t.Errorf("%s point %d: IndicatedAirspeed = %v, details ias = %d", f, i, p.IndicatedAirspeed, *details.IAS)
			}
		}
		if pt.T != "C172" && checked == 0 {
			t.Errorf("%s (%s): no points with details ias, cross-check did not run", f, pt.T)
		}
		t.Logf("%s: cross-checked %d points", f, checked)
	}
}
