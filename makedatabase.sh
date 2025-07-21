#!/usr/bin/env bash

sqlite3 app.sqlite < api/schema.sql

go run -tags "fts5" ./api/load_data/ \
	./app.sqlite ~/Downloads/tt/stops.txt

# run tests
go test -tags "fts5" ./api/api_test/