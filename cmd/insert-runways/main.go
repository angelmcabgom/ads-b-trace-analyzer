package main

import (
	"fmt"
	"time"

	"ads-b-geom-parser/internal/ingest"
)

type RunwayRow struct {
	id					string
	airport_ref
	airport_ident
	length_ft
	width_ft
	surface
	lighted
	closed
	le_ident
	le_latitude_deg
	le_longitude_deg
	le_elevation_ft
	le_heading_degT
	le_displaced_threshold_ft
	he_ident
	he_latitude_deg
	he_longitude_deg
	he_elevation_ft
	he_heading_degT
	he_displaced_threshold_ft
}


// TODO: define RunwayRow, parse runways.csv into it and insert into a runways table
// (needs its own migration first).
func main() {
	start := time.Now()

	reader := ingest.DownloadCSV("https://davidmegginson.github.io/ourairports-data/runways.csv")
	dbConn := ingest.ConnectDB()
	defer dbConn.Close()

	_ = reader

	fmt.Printf("Total time: %v\n", time.Since(start))
}


func parseRunwayData() {

}