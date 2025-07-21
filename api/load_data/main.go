package main

// go run -tags "icu json1 fts5 secure_delete" ./route/route_load_data ~/Downloads/tt/stops.txt

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/isobelmcrae/trip/api"
)

func ParseStations(filePath string) ([]api.StopSearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var utf8bom = []byte{0xef, 0xbb, 0xbf}

	bom := make([]byte, 3)
	_, err = file.Read(bom)
	if err != nil {
		return nil, fmt.Errorf("error reading for BOM check: %w", err)
	}

	if !bytes.Equal(bom, utf8bom) {
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("error seeking back to start of file: %w", err)
		}
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read the header to find column indices dynamically. This is robust.
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	fmt.Printf("header: %v\n", header)

	// Create a map to hold the index of each column we care about.
	columnIndex := make(map[string]int)
	requiredCols := []string{"stop_id", "stop_code", "stop_name", "stop_lat", "stop_lon", "location_type", "parent_station", "level_id"}
	for _, colName := range requiredCols {
		found := false
		for i, h := range header {
			if h == colName {
				columnIndex[colName] = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("missing required column in CSV: %s", colName)
		}
	}

	var stations []api.StopSearchResult
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record: %w", err)
		}

		if record[columnIndex["parent_station"]] == "" {
			lat, err := strconv.ParseFloat(record[columnIndex["stop_lat"]], 64)
			if err != nil {
				log.Printf("Warning: skipping row, could not parse latitude for stop ID %s", record[columnIndex["stop_id"]])
				continue
			}

			lon, err := strconv.ParseFloat(record[columnIndex["stop_lon"]], 64)
			if err != nil {
				log.Printf("Warning: skipping row, could not parse longitude for stop ID %s", record[columnIndex["stop_id"]])
				continue
			}

			station := api.StopSearchResult{
				StopId:   record[columnIndex["stop_id"]],
				Name: record[columnIndex["stop_name"]],
				Lat:  lat,
				Lon:  lon,
			}
			stations = append(stations, station)
		}
	}

	return stations, nil
}

func main() {
	stopsGtfsPath := os.Args[1] // go run api_load_schema ./stops.txt

	stops, err := ParseStations(stopsGtfsPath)

	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "app.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into stop(id, name, lat, lon) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for i, stop := range stops {
		_, err = stmt.Exec(stop.StopId, stop.Name, stop.Lat, stop.Lon)
		if err != nil {
			log.Printf("Error inserting stop %d (%s): %v", i, stop.StopId, err)
			continue // Skip this stop and continue with the next
		}
	}
	log.Printf("Inserted %d stops into the database", len(stops))

	err = tx.Commit()

	_, err = db.Exec(`
		INSERT INTO stop_fts (id, name)
		SELECT id, name FROM stop;
	`)
	if err != nil {
		log.Printf("Error rebuilding FTS index: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO stop_fts(stop_fts) VALUES('optimize');
	`)
	if err != nil {
		log.Printf("Error optimizing FTS index: %v", err)
	}

	if err != nil {
		log.Fatal(err)
	}
}
