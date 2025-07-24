#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
	echo "Usage: $0 <path to GTFS folder>"
	exit 1
fi

sqlite3 app.sqlite < api/schema.sql

go run -tags "fts5" ./api_load_data/ \
	./app.sqlite $1

# run tests
go test -tags "fts5" ./api/api_test/
