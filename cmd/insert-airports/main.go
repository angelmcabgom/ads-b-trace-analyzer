package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"ads-b-geom-parser/internal/ingest"
)

type AirportRow struct {
	id           int32
	ident        string
	airportType  string
	name         string
	location     ingest.CoordsPoint
	elevationFt  int32
	isoCountry   string
	municipality string
	icaoCode     string
	iataCode     string
}

func main() {
	start := time.Now()
	rows := downloadAndParseAirportData()
	dbConn := ingest.ConnectDB()
	err := insertIntoDb(rows, dbConn)

	if err != nil {
		log.Fatalf("[main] Error when insert into DB: %s", err)
	}

	ingest.ReportParseFailures(os.Stderr)

	elapsed := time.Since(start)
	// fmt.Print(len(rows))
	fmt.Printf("Total time: %v\n", elapsed)
}

func downloadAndParseAirportData() []AirportRow {
	reader := ingest.DownloadCSV("https://davidmegginson.github.io/ourairports-data/airports.csv")

	airportRows := []AirportRow{}

	for {
		row, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Print(err)
				break
			}
		}

		elevFt := ingest.ParseInt32(row[6])

		airportRows = append(airportRows, AirportRow{
			id:          ingest.ParseInt32(row[0]),
			ident:       row[1],
			airportType: row[2],
			name:        row[3],
			location: ingest.CoordsPoint{
				Longitude: ingest.ParseFloat64(row[5]), // row[5] is lon
				Latitude:  ingest.ParseFloat64(row[4]), // row[4] is lat
				Z:         float64(elevFt) * ingest.FeetToMeters,
			},
			elevationFt:  elevFt,
			isoCountry:   row[8],
			municipality: row[10],
			icaoCode:     row[12],
			iataCode:     row[13],
		})
	}

	return airportRows
}

func insertIntoDb(dataFrame []AirportRow, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op after a successful Commit

	stmt, err := tx.Prepare(`
    INSERT INTO public.airports
        (id, ident, "type", "name", "location", elevation_ft,
         iso_country, municipality, icao_code, iata_code)
    VALUES
        ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6, $7), 4326), $8,
         $9, $10, $11, $12)
    ON CONFLICT (id) DO NOTHING`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, a := range dataFrame {
		_, err := stmt.Exec(
			a.id, a.ident, a.airportType, a.name,
			a.location.Longitude, a.location.Latitude, a.location.Z,
			a.elevationFt,
			a.isoCountry, a.municipality, a.icaoCode, a.iataCode,
		)
		if err != nil {
			return fmt.Errorf("inserting airport %d: %w", a.id, err)
		}
	}

	return tx.Commit()
}
