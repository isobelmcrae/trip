package main

// go run -tags "icu json1 fts5 secure_delete" ./route/route_load_data ~/Downloads/tt/stops.txt

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/isobelmcrae/trip/api"
	_ "github.com/mattn/go-sqlite3"
)

func parseCSV[T any](filePath string, parser func([]string, map[string]int) (T, error)) ([]T, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", filePath, err)
	}
	defer file.Close()

	// Skip UTF-8 BOM if present
	bom := make([]byte, 3)
	file.Read(bom)
	if string(bom) != "\xef\xbb\xbf" {
		file.Seek(0, 0)
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header from %s: %w", filePath, err)
	}

	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[h] = i
	}

	var results []T
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record from %s: %w", filePath, err)
		}
		item, err := parser(record, colIndex)
		if err == nil {
			results = append(results, item)
		}
	}
	return results, nil
}

func main() {
	// go run api_load_schema ./app.sqlite ./stops.txt
	databasePath := os.Args[1]
	gtfsPath := os.Args[2]

	stops, _ := parseCSV(gtfsPath + "/stops.txt", func(r []string, c map[string]int) (api.StopSearchResult, error) {
		if r[c["parent_station"]] != "" {
			return api.StopSearchResult{}, fmt.Errorf("is child station")
		}
		lat, _ := strconv.ParseFloat(r[c["stop_lat"]], 64)
		lon, _ := strconv.ParseFloat(r[c["stop_lon"]], 64)
		return api.StopSearchResult{ID: r[c["stop_id"]], Name: r[c["stop_name"]], Lat: lat, Lon: lon}, nil
	})

	stopTimes, _ := parseCSV(gtfsPath + "/stop_times.txt", func(r []string, c map[string]int) (api.GtfsStopTime, error) {
		seq, _ := strconv.Atoi(r[c["stop_sequence"]])
		dist, _ := strconv.ParseFloat(r[c["shape_dist_traveled"]], 64)
		return api.GtfsStopTime{TripID: r[c["trip_id"]], StopID: r[c["stop_id"]], Sequence: seq, DistanceTraveled: dist }, nil
	})

	trips, _ := parseCSV(gtfsPath + "/trips.txt", func(r []string, c map[string]int) (api.GtfsTrip, error) {
		return api.GtfsTrip{TripID: r[c["trip_id"]], RouteID: r[c["route_id"]], ShapeID: r[c["shape_id"]]}, nil
	})

	shapePoints, _ := parseCSV(gtfsPath + "/shapes.txt", func(r []string, c map[string]int) (api.GtfsShapePoint, error) {
		lat, _ := strconv.ParseFloat(r[c["shape_pt_lat"]], 64)
		lon, _ := strconv.ParseFloat(r[c["shape_pt_lon"]], 64)
		seq, _ := strconv.Atoi(r[c["shape_pt_sequence"]])
		dist, _ := strconv.ParseFloat(r[c["shape_dist_traveled"]], 64)
		return api.GtfsShapePoint{ShapeID: r[c["shape_id"]], Lat: lat, Lon: lon, Sequence: seq, DistanceTraveled: dist}, nil
	})

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Use a transaction for massive speed improvement
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Inserting stops...")
	stmt, err := tx.Prepare("INSERT INTO stop(id, name, lat, lon) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range stops {
		stmt.Exec(s.ID, s.Name, s.Lat, s.Lon)
	}
	stmt.Close()

	log.Println("Inserting stop_times...")
	stmt, err = tx.Prepare("INSERT INTO stop_times(trip_id, stop_id, stop_sequence, shape_dist_traveled) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	for _, st := range stopTimes {
		stmt.Exec(st.TripID, st.StopID, st.Sequence, st.DistanceTraveled)
	}
	stmt.Close()

	log.Println("Inserting trips...")
	stmt, err = tx.Prepare("INSERT INTO trips(trip_id, route_id, shape_id) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range trips {
		stmt.Exec(t.TripID, t.RouteID, t.ShapeID)
	}
	stmt.Close()

	log.Println("Inserting shapes...")
	stmt, err = tx.Prepare("INSERT INTO shapes(shape_id, shape_pt_lat, shape_pt_lon, shape_pt_sequence, shape_dist_traveled) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	for _, sp := range shapePoints {
		stmt.Exec(sp.ShapeID, sp.Lat, sp.Lon, sp.Sequence, sp.DistanceTraveled)
	}
	stmt.Close()

	log.Println("Committing transaction...")
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	log.Println("Database load complete.")

	// Rebuild FTS index
	log.Println("Rebuilding FTS index...")
	db.Exec(`INSERT INTO stop_fts (id, name) SELECT id, name FROM stop;`)
	db.Exec(`INSERT INTO stop_fts(stop_fts) VALUES('optimize');`)

	log.Println("VACUUM + ANALYZE...")
	db.Exec(`VACUUM;`)
	db.Exec(`ANALYZE;`)
	log.Println("FTS index complete.")
}
