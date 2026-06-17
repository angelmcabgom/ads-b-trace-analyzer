package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"ads-b-geom-parser/analysis"
	"ads-b-geom-parser/geojson"
	"ads-b-geom-parser/metrics"
	"ads-b-geom-parser/trace"
)

func runMigration(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatalf("could not create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)

	if err != nil {
		log.Fatalf("could not init migration: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("cloud not run migration: %v", err)
	}

	log.Printf("migration ran successfully")
}

func loadTraceData(path string) (trace.PlaneTrace, error) {
	var t trace.PlaneTrace

	file, err := os.Open(path)
	if err != nil {
		return t, err
	}
	defer file.Close()

	var reader io.Reader = file

	gzReader, err := gzip.NewReader(file)
	if err == nil {
		reader = gzReader
		defer gzReader.Close()
	} else {
		file.Seek(0, 0)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return t, err
	}

	if err := json.Unmarshal(data, &t); err != nil {
		return t, err
	}

	return t, nil
}

func main() {
	defer metrics.TrackPerformance()()
	runAnalysis := flag.Bool("analyze", true, "Execute flight behavior analysis")
	outDir := flag.String("out", "output", "Directory to save GeoJSON files")
	inputPath := flag.String("in", "traces", "Path to a single file or directory")
	runMigrationsFlag := flag.Bool("migration", false, "Excute migrations")

	flag.Parse()

	if *runMigrationsFlag {
		err := godotenv.Load("envs/.local.env")

		if err != nil {
			log.Fatal("failed to load env vars")
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
			log.Fatalf("could not connect to db: %v", err)
		}

		runMigration(db)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("Could not create output directory: %v", err)
	}

	info, err := os.Stat(*inputPath)

	if err != nil {
		log.Fatalf("Error accessing input path: %v", err)
	}

	if info.IsDir() {
		fmt.Printf("Processing directory: %s\n", *inputPath)
		count := 0
		err := filepath.WalkDir(*inputPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".json") {
				processFile(path, *outDir, *runAnalysis)
				count++
				if count%100 == 0 {
					fmt.Printf("Processed %d files...\n", count)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		processFile(*inputPath, *outDir, *runAnalysis)
	}
}

func processFile(path, outDir string, runAnalysis bool) {
	flightData, err := loadTraceData(path)
	if err != nil {
		log.Printf("Skipping %s: %v", path, err)
		return
	}

	if runAnalysis {
		if analysis.DetectGoAround(flightData.Trace) {
			fmt.Printf("Go-around detected in: %s\n", path)
		}

		if analysis.DetectTakeOff(flightData.Trace) {
			fmt.Printf("Takeoff detected")
		}
	}

	geoJsonFeature := geojson.ParseTraceIntoGeoJSON(flightData)

	filename := filepath.Base(path)
	savePath := filepath.Join(outDir, strings.TrimSuffix(filename, ".json")+".geojson")

	if err := geojson.SaveFile(savePath, geoJsonFeature); err != nil {
		log.Printf("Failed to save %s: %v", savePath, err)
	}
}
