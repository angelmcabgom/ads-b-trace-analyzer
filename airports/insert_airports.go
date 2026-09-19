package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// need type for row

type CoordsPoint struct {
	Longitude float64 // X
	Latitude  float64 // Y
	Z         float64 // altitude, meters
}

type AirportRow struct {
	id           int32
	ident        string
	airportType  string
	name         string
	location     CoordsPoint
	elevationFt  int32
	isoCountry   string
	municipality string
	icaoCode     string
	iataCode     string
}

const feetToMeters = 0.3048

func main() {
	start := time.Now()
	rows := downloadAndParseAirportData()
	dbConn := connectIntoDb()
	err := insertIntoDb(rows, dbConn)

	if err != nil {
		log.Fatalf("[main] Error when insert into DB: %s", err)
	}

	elapsed := time.Since(start)
	// fmt.Print(len(rows))
	fmt.Printf("Total time: %v\n", elapsed)
}

// can easily be refactored into a general download function
func downloadAndParseAirportData() []AirportRow {
	resp, err := http.Get("https://davidmegginson.github.io/ourairports-data/airports.csv")

	if err != nil {
		log.Fatal("Error downloading airportData")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error reading downloaded data")
	}

	reader := csv.NewReader(bytes.NewBuffer(body))

	airportRows := []AirportRow{}

	_, err = reader.Read() // skip header row

	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	for {
		row, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Print(err)
				break
			}
		}

		elevFt := parseInt32(row[6])

		airportRows = append(airportRows, AirportRow{
			id:          parseInt32(row[0]),
			ident:       row[1],
			airportType: row[2],
			name:        row[3],
			location: CoordsPoint{
				Longitude: parseFloat64(row[5]), // row[5] is lon
				Latitude:  parseFloat64(row[4]), // row[4] is lat
				Z:         float64(elevFt) * feetToMeters,
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

func parseInt32(s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		fmt.Printf("Failed to parse int 32: %v", v)
	}

	return int32(v)
}

func parseFloat32(s string) float32 {
	v, err := strconv.ParseFloat(s, 32)

	if err != nil {
		fmt.Printf("Failed to parse float 32: %v", v)
	}
	return float32(v)
}

func parseFloat64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		fmt.Printf("Failed to parse float 64: %v", v)
	}
	return float64(v)
}

func connectIntoDb() *sql.DB {
	err := godotenv.Load("../envs/.local.env")

	if err != nil {
		log.Fatalf("Error when connecting into db: %v", err)
	}

	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error after trying to connect to DB: %v", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatalf("Error when pinging db: %s, connStr: %s", err, connStr)
	}

	return db
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
