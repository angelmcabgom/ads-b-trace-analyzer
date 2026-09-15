package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// need type for row

type AirportRow struct {
	id 				int32
	ident 			string
	airportType		string
	name			string
	location		CoordsPoint
	elevation		int32
	iso_country		string
	municipality	string
	icao_code		string
	iata_code		string
}

type CoordsPoint struct {
	Latitude       float32
	Longitude      float32
}


func main(){
	start := time.Now()
	rows := downloadAirportData()
	elapsed := time.Since(start)
	fmt.Print(len(rows))
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
			id: parseInt32(row[0]),
			ident: row[1],
			airportType: row[2],
			name: row[3],
			location: CoordsPoint{
				Latitude: parseFloat32(row[4]),
				Longitude: parseFloat32(row[5]),
			},
			elevation: parseInt32(row[6]),
			iso_country: row[8],
			municipality: row[10],
			icao_code: row[12],
			iata_code: row[13],
		})
	}

	return airportRows
}

func parseInt32(s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		fmt.Printf("Failed to parse int 32, %e: %v", err, v)
	}

	return int32(v)
}

func parseFloat32(s string) float32 {
	v, err := strconv.ParseFloat(s, 32)

	if err != nil {
		fmt.Printf("Failed to parse float 32, %e: %v", err, v)
	}
	return float32(v)
}

func insertIntoDb(csvPath string) bool {
	
	
	return true
}
