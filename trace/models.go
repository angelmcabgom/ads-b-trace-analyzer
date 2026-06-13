package trace

import (
	"encoding/json"
	"fmt"
)

// --- STRUCTS (Kept identical to your original code) ---

type PlaneTrace struct {
	ICAO      string       `json:"icao"`
	R         string       `json:"r"`
	T         string       `json:"t"`
	DBFlags   int          `json:"dbFlags"`
	Desc      string       `json:"desc"`
	Timestamp float64      `json:"timestamp"`
	Trace     []TracePoint `json:"trace"`
}

type TraceMeta struct {
	Type      string  `json:"type"`
	Flight    string  `json:"flight"`
	AltGeom   int     `json:"alt_geom"`
	Track     float64 `json:"track"`
	BaroRate  int     `json:"baro_rate"`
	GeomRate  int     `json:"geom_rate"`
	Squawk    string  `json:"squawk"`
	Emergency string  `json:"emergency"`
}

type TracePoint struct {
	TimeOffset     float64
	Latitude       float32
	Longitude      float32
	Altitude       int
	Track          float64
	GroundSpeed    float64
	Unknown6       int
	VerticalRate   int
	MetaData       *TraceMeta
	PositionSource string
	AltGeom        *float64
	GeometricRate  int
	Heading        *int
	Roll           *float64
}

// custom unmarshaler

func (tp *TracePoint) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	fields := []any{
		&tp.TimeOffset,
		&tp.Latitude,
		&tp.Longitude,
		&tp.Altitude,
		&tp.Track,
		&tp.GroundSpeed,
		&tp.Unknown6,
		&tp.VerticalRate,
		&tp.MetaData,
		&tp.PositionSource,
		&tp.AltGeom,
		&tp.GeometricRate,
		&tp.Heading,
		&tp.Roll,
	}

	for i, field := range fields {
		if i >= len(raw) {
			break
		}
		if err := json.Unmarshal(raw[i], field); err != nil {
			return fmt.Errorf("field %d: %w", i, err)
		}
	}

	return nil
}