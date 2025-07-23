# trip

- run makedatabase.sh w/ stops.txt from https://opendata.transport.nsw.gov.au/data/dataset/timetables-complete-gtfs
- create TFNSW_KEY environment variable w/ apikey - https://opendata.transport.nsw.gov.au/developers/userguide

need to run with `'fts5'` flag: `go run -tags 'fts5' main.go`
