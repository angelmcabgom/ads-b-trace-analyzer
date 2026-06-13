package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"


	"ads-b-geom-parser/metrics"
	"ads-b-geom-parser/analysis"
	"ads-b-geom-parser/geojson"
	"ads-b-geom-parser/trace"
)


func loadTraceData(path string) (trace.PlaneTrace, error) {
	var t trace.PlaneTrace
	data, err := os.ReadFile(path)
	if err != nil {
		return t, fmt.Errorf("failed to read file: %w", err)
	}
	if err := json.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return t, nil
}

func main() {
	defer metrics.TrackPerformance()()

	runAnalysis := flag.Bool("analyze", true, "Execute flight behavior analysis")
	outputPath := flag.String("out", "flight_track.geojson", "Destination path for the GeoJSON file")
	flag.Parse()

	mockJsonPath := filepath.Join("mock", "trace-big.mock.json")

	flightData, err := loadTraceData(mockJsonPath)
	if err != nil {
		log.Fatalf("Data loading error: %v", err)
	}

	if *runAnalysis {
		fmt.Println("Running flight path analysis...")
		if analysis.DetectGoAround(flightData.Trace) {
			fmt.Println("!! Go around detected for dataset")
		} else {
			fmt.Println("No anomalies detected.")
		}
	}

	geoJsonFeature := geojson.ParseTraceIntoGeoJSON(flightData)
	if err := geojson.SaveFile(*outputPath, geoJsonFeature); err != nil {
		log.Fatalf("Data output error: %v", err)
	}

	fmt.Print(geoJsonFeature)
}