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
