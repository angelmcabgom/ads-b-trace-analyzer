package geojson

import (
	"ads-b-geom-parser/trace"
	"encoding/json"
	"fmt"
	"os"
)

type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
}


func ParseTraceIntoGeoJSON(traceJson trace.PlaneTrace) GeoJSONFeature {
	coordinates := make([][]float64, 0, len(traceJson.Trace))
	for _, pt := range traceJson.Trace {
		coord := []float64{
			float64(pt.Longitude),
			float64(pt.Latitude),
			float64(pt.Altitude),
			}
		coordinates = append(coordinates, coord)
	}

	props := map[string]interface{}{
		"icao":         traceJson.ICAO,
		"registration": traceJson.R,
		"type":         traceJson.T,
		"description":  traceJson.Desc,
		"start_time":   traceJson.Timestamp,
	}

	return GeoJSONFeature{
		Type:       "Feature",
		Properties: props,
		Geometry: GeoJSONGeometry{
			Type:        "LineString",
			Coordinates: coordinates,
		},
	}
}

func SaveFile(path string, feature GeoJSONFeature) error {
	output, err := json.MarshalIndent(feature, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal GeoJSON: %w", err)
	}
	if err := os.WriteFile(path, output, 0644); err != nil {
		return fmt.Errorf("failed to write GeoJSON file: %w", err)
	}
	fmt.Printf("Successfully saved GeoJSON trace to: %s\n", path)
	return nil
}
