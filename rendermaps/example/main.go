package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/rendermaps"
	"github.com/isobelmcrae/trip/state"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env")
    }

	db, err := sql.Open("sqlite3", state.DatabasePath)
    if err != nil {
        log.Fatal(err)
    }
	defer db.Close()

	tc := api.NewClient(db)

	begin := tc.FindStopFirstOrPanic("circular quay")
	end := tc.FindStopFirstOrPanic("central station")

	journies, err := tc.TripPlan(context.TODO(), begin.ID, end.ID)
	if err != nil {
		log.Fatalf("cannot plan trip: %v", err)
	}

	journey := journies[0]
	fmt.Printf("journey: %v\n", journey)

	err = rendermaps.TripRender(journey)
	if err != nil {
		log.Fatalf("cannot render trip: %v", err)
	}
}
