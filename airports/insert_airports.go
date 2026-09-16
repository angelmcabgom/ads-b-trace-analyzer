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

type AirportRow struct {
	id           int32
	ident        string
	airportType  string
	name         string
	location     CoordsPoint
	elevation    int32
	iso_country  string
	municipality string
	icao_code    string
	iata_code    string
}

type CoordsPoint struct {
	Latitude  float32
	Longitude float32
}

func main() {
	start := time.Now()
	rows := downloadAirportData()
	dbConn := connectIntoDb()
	insertIntoDb(rows, dbConn)
	elapsed := time.Since(start)
	// fmt.Print(len(rows))
	fmt.Printf("Total time: %v\n", elapsed)
}

// can easily be refactored into a general download function
func downloadAirportData() []AirportRow {
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

	// skip first row
	for {
		row, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Print(err)
				break
			}
		}

		airportRows = append(airportRows, AirportRow{
			id:          parseInt32(row[0]),
			ident:       row[1],
			airportType: row[2],
			name:        row[3],
			location: CoordsPoint{
				Latitude:  parseFloat32(row[4]),
				Longitude: parseFloat32(row[5]),
			},
			elevation:    parseInt32(row[6]),
			iso_country:  row[8],
			municipality: row[10],
			icao_code:    row[12],
			iata_code:    row[13],
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

func insertIntoDb(dataFrame []AirportRow, db *sql.DB) bool {

	response, err := db.Query("select 1;")

	if err != nil {
		log.Fatalf("Error when querying database: %v", err)
	}

	defer response.Close()

	fmt.Print(response)

	// for _, airportRow := range dataFrame {
	// }

	return true
}
