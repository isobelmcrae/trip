package api

import (
	"regexp"

	"github.com/charmbracelet/log"
)

// go build -tags "icu json1 fts5 secure_delete"

var (
	gReplacementReg = regexp.MustCompile(`[^A-Za-z0-9\s]`)
	gWordReg        = regexp.MustCompile(`([A-Za-z0-9]+)`)
)

const (
	SearchStopMaxResults = 25
)

// used in tests, useless export
func SanitiseSeach(search string) string {
	finalSearch := gReplacementReg.ReplaceAllLiteralString(search, "")
	finalSearch = gWordReg.ReplaceAllString(finalSearch, "$1*")

	return finalSearch
}

type StopSearchResult struct {
	ID string
	Name   string
	Lat    float64
	Lon    float64
}

// this should never fail
func (tc *TripClient) FindStop(search string) []StopSearchResult {
	// assumed to finish quickly, context unnecessary
	rows, err := tc.db.Query(`
		select s.id, s.name, s.lat, s.lon
		from stop_fts as fts
			join stop as s on fts.id = s.id
		where fts.name match ?
		order by rank
		limit ?
	`, SanitiseSeach(search), SearchStopMaxResults)
	if err != nil {
		log.Fatalf("cannot perform search: %v", err)
	}

	results := make([]StopSearchResult, 0, SearchStopMaxResults)

	defer rows.Close()
	for rows.Next() {
		var id string
		var name string
		var lat float64
		var lon float64

		err = rows.Scan(&id, &name, &lat, &lon)
		if err != nil {
			log.Fatalf("cannot scan rows: %v", err)
		}

		results = append(results, StopSearchResult{
			ID: id,
			Name: name,
			Lat: lat,
			Lon: lon,
		})
	}

	return results
}

func (tc *TripClient) FindStopFirstOrPanic(search string) StopSearchResult {
	results := tc.FindStop(search)
	if len(results) == 0 {
		log.Fatalf("no results found for search: %s", search)
	}
	return results[0]
}
