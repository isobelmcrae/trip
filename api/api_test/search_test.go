package api_test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/state"

	_ "github.com/mattn/go-sqlite3"
)

func TestSearchStopSanitiseSearch(t *testing.T) {
	verify := func(search string, output string) {
		got := api.SanitiseSeach(search)
		if got != output {
			t.Errorf("%s != %s, got %s", search, output, got)
		}
	}

	verify(`syd airport`, `syd* airport*`)
	verify(`international airport`, `international* airport*`)
}

func TestSearchStop(t *testing.T) {
	db, err := sql.Open("sqlite3", state.DatabasePath)
    if err != nil {
        log.Fatal(err)
    }
	defer db.Close()

	tc := api.NewClient(db)

	verify := func(search string, ID ...string) {
		stops := tc.FindStop(search)
		found := make([]bool, len(ID))
		for _, stop := range stops {
			for i, id := range ID {
				if stop.ID == id {
					found[i] = true
				}
			}
		}

		for i, id := range ID {
			if !found[i] {
				t.Errorf("stop %s not found in search for %s", id, search)
			}
		}
	}
	
	verify("syd airport",
		"202020", // Sydney Domestic Airport Station
		"202030", // Sydney International Airport Station
	)

	verify("int airport",
		"202030", // Sydney International Airport Station
	)

	verify("central",
		"200060", // Central Station
	)
}
