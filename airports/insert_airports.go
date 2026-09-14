package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)


// need type for row

type AirportRow struct {
	id 	string
}

func main(){
	start := time.Now()
	rows := downloadAirportData()
	elapsed := time.Since(start)
	fmt.Print(rows)
	fmt.Printf("Total time: %v\n", elapsed)
}

// can easily be refactored into a general download function
func downloadAirportData() [][]string {
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
	
	rows := [][]string{}
	for {
		row, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Print(err)
				break
			}
		}

		rows = append(rows, row)
	}

	return rows
}
