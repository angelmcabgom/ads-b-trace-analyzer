// Package ingest holds helpers shared by the dataset insertion commands in cmd/.
package ingest

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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type CoordsPoint struct {
	Longitude float64 // X
	Latitude  float64 // Y
	Z         float64 // altitude, meters
}

const FeetToMeters = 0.3048

// DownloadCSV fetches a CSV file and returns a reader positioned after the header row.
func DownloadCSV(url string) *csv.Reader {
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("Error downloading %s", url)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error reading downloaded data")
	}

	reader := csv.NewReader(bytes.NewBuffer(body))

	_, err = reader.Read() // skip header row

	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	return reader
}


var parseFailures = map[string]int{}
var parseSamples = map[string]string{}

func recordParseFailure(kind, value string) {
	if parseFailures[kind] == 0 {
		parseSamples[kind] = value
	}

	parseFailures[kind]++
}


func ReportParseFailures(w io.Writer) {
	for kind, count := range parseFailures {
		fmt.Fprintf(w, "%s: %d unparseable values, first was %q\n", kind, count, parseSamples[kind])
	}
}

func ParseInt32(s string) int32 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		recordParseFailure("int32", s)
	}

	return int32(v)
}

func ParseFloat32(s string) float32 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 32)

	if err != nil {
		recordParseFailure("float32", s)
	}

	return float32(v)
}

func ParseFloat64(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		recordParseFailure("float64", s)
	}

	return v
}

func ParseBool(s string) bool {
	if s == "" {
		return false
	}

	v, err := strconv.ParseBool(s)

	if err != nil {
		recordParseFailure("bool", s)
	}

	return v
}

func NullInt32(s string) sql.NullInt32 {
	v, err := strconv.ParseInt(s, 10, 32)

	if err != nil && s != "" {
		recordParseFailure("int32", s)
	}

	return sql.NullInt32{Int32: int32(v), Valid: err == nil}
}

func NullFloat64(s string) sql.NullFloat64 {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil && s != "" {
		recordParseFailure("float64", s)
	}

	return sql.NullFloat64{Float64: v, Valid: err == nil}
}

func NullBool(s string) sql.NullBool {
	v, err := strconv.ParseBool(s)

	if err != nil && s != "" {
		recordParseFailure("bool", s)
	}

	return sql.NullBool{Bool: v, Valid: err == nil}
}

func NullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ConnectDB loads envs/.local.env (relative to the repo root) and opens a Postgres connection.
func ConnectDB() *sql.DB {
	err := godotenv.Load("../../envs/.local.env")

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
