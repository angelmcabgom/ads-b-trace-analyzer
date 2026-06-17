package trace

import (
    "encoding/json"
    "fmt"
)

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

func (tp *TracePoint) UnmarshalJSON(data []byte) error {
    var raw []json.RawMessage
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }

    // intercept altitude with a raw message to prevent type panics
    var rawAltitude json.RawMessage

    fields := []any{
        &tp.TimeOffset,     
        &tp.Latitude,       
        &tp.Longitude,      
        &rawAltitude,       // mixed type handling ("ground", null, or int)
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

    // Evaluate the raw altitude chunk safely
    if len(rawAltitude) > 0 {
        var altInt int
        // Strategy 1: Try parsing it as a normal integer first
        if err := json.Unmarshal(rawAltitude, &altInt); err == nil {
            tp.Altitude = altInt
        } else {
            var altStr string
            // Strategy 2: If it's a string, see if it says "ground"
            if err := json.Unmarshal(rawAltitude, &altStr); err == nil && altStr == "ground" {
                tp.Altitude = 0 // Coerce runway positions cleanly to 0 feet
            } else if string(rawAltitude) == "null" {
                tp.Altitude = 0 // Treat missing telemetry gracefully
            } else {
                return fmt.Errorf("field 3 (Altitude) has unparseable value: %s", string(rawAltitude))
            }
        }
    }

    return nil
}