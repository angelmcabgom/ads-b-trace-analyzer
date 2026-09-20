package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"ads-b-geom-parser/internal/ingest"
)

type RunwayRow struct {
	id                        int32
	airport_ref               int32
	airport_ident             string
	length_ft                 sql.NullInt32
	width_ft                  sql.NullInt32
	surface                   sql.NullString
	lighted                   bool
	closed                    bool
	le_ident                  sql.NullString
	le_latitude_deg           sql.NullFloat64
	le_longitude_deg          sql.NullFloat64
	le_elevation_ft           sql.NullInt32
	le_heading_deg_t          sql.NullFloat64
	le_displaced_threshold_ft sql.NullInt32
	he_ident                  sql.NullString
	he_latitude_deg           sql.NullFloat64
	he_longitude_deg          sql.NullFloat64
	he_elevation_ft           sql.NullInt32
	he_heading_deg_t          sql.NullFloat64
	he_displaced_threshold_ft sql.NullInt32
}

func main() {
	start := time.Now()

	reader := ingest.DownloadCSV("https://davidmegginson.github.io/ourairports-data/runways.csv")
	dbConn := ingest.ConnectDB()
	defer dbConn.Close()

	runwayRows := parseRunwayData(reader)

	ingest.ReportParseFailures(os.Stderr)

	if err := insertIntoDb(runwayRows, dbConn); err != nil {
		log.Fatalf("[main] Error when insert into DB: %s", err)
	}

	fmt.Printf("Total time: %v\n", time.Since(start))
}

// func mostly done handle null and allat
func parseRunwayData(readerReference *csv.Reader) []RunwayRow {
	runwayRows := []RunwayRow{}

	for {
		row, err := readerReference.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Print(err)
				break
			}
			log.Fatalf("Error when trying to open reader, %s", err)
		}

		runwayRows = append(runwayRows, RunwayRow{
			id:                        ingest.ParseInt32(row[0]),
			airport_ref:               ingest.ParseInt32(row[1]),
			airport_ident:             row[2],
			length_ft:                 ingest.NullInt32(row[3]),
			width_ft:                  ingest.NullInt32(row[4]),
			surface:                   ingest.NullString(row[5]),
			lighted:                   ingest.ParseBool(row[6]),
			closed:                    ingest.ParseBool(row[7]),
			le_ident:                  ingest.NullString(row[8]),
			le_latitude_deg:           ingest.NullFloat64(row[9]),
			le_longitude_deg:          ingest.NullFloat64(row[10]),
			le_elevation_ft:           ingest.NullInt32(row[11]),
			le_heading_deg_t:          ingest.NullFloat64(row[12]),
			le_displaced_threshold_ft: ingest.NullInt32(row[13]),
			he_ident:                  ingest.NullString(row[14]),
			he_latitude_deg:           ingest.NullFloat64(row[15]),
			he_longitude_deg:          ingest.NullFloat64(row[16]),
			he_elevation_ft:           ingest.NullInt32(row[17]),
			he_heading_deg_t:          ingest.NullFloat64(row[18]),
			he_displaced_threshold_ft: ingest.NullInt32(row[19]),
		})
	}

	return runwayRows
}

func insertIntoDb(runwayRows []RunwayRow, db *sql.DB) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback() // no-op after a successful Commit

	// geom is built from the two thresholds. ST_MakePoint takes (lon, lat) and
	// is strict, so a runway missing either endpoint two thirds of the
	// dataset - yields NULL rather than a line anchored at 0,0.
	stmt, err := tx.Prepare(`
    INSERT INTO public.runways
        (id, airport_ref, airport_ident, length_ft, width_ft, surface,
         lighted, closed,
         le_ident, le_latitude_deg, le_longitude_deg, le_elevation_ft,
         le_heading_deg_t, le_displaced_threshold_ft,
         he_ident, he_latitude_deg, he_longitude_deg, he_elevation_ft,
         he_heading_deg_t, he_displaced_threshold_ft,
         geom)
    VALUES
        ($1, $2, $3, $4, $5, $6,
         $7, $8,
         $9, $10, $11, $12,
         $13, $14,
         $15, $16, $17, $18,
         $19, $20,
         ST_SetSRID(ST_MakeLine(ST_MakePoint($11, $10), ST_MakePoint($17, $16)), 4326))
    ON CONFLICT (id) DO NOTHING`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, r := range runwayRows {
		_, err := stmt.Exec(
			r.id, r.airport_ref, r.airport_ident, r.length_ft, r.width_ft, r.surface,
			r.lighted, r.closed,
			r.le_ident, r.le_latitude_deg, r.le_longitude_deg, r.le_elevation_ft,
			r.le_heading_deg_t, r.le_displaced_threshold_ft,
			r.he_ident, r.he_latitude_deg, r.he_longitude_deg, r.he_elevation_ft,
			r.he_heading_deg_t, r.he_displaced_threshold_ft,
		)

		if err != nil {
			return fmt.Errorf("inserting runway %d: %w", r.id, err)
		}
	}

	return tx.Commit()
}
