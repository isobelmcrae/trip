# trip 🚉

sydney public transport for your terminal

![https://vhs.charm.sh/vhs-3CWGX0CiqP2yfWXnOxASOC.gif]

# Setup

Required:
- recent version of `sqlite3` (tested with >= `3.50.2`)
- TFNSW ("Timetables Complete GTFS")[https://opendata.transport.nsw.gov.au/data/dataset/timetables-complete-gtfs]
- `TFNSW_KEY` environment variable set to a (TFNSW Open Data Hub API key)[https://opendata.transport.nsw.gov.au/developers/api-basics]

To create the database, run `makedatabase.sh` with the path to the unzipped GTFS data:
```bash
./makedatabase.sh /path/to/gtfs
```
Then build and run:
```bash
go build -o trip -tags 'fts5' main.go
./trip
```

## Server Deployment

`trip` offers a 'ssh' mode which allows one to host the app and support connections over `ssh`.

App uses port 22 by default - change `defaultSSHAddr` in `main.go` before build.
Build the app and allow it to bind to reserved ports (e.g. port 22).

```bash
go build -o trip -tags 'fts5' main.go
sudo setcap CAP_NET_BIND_SERVICE=+eip trip
```

Run the app with the `--ssh` flag:

```bash
./trip --ssh
# then connect 
ssh user@your.domain.here -p your_port # or similar
```

TODO: Add docs
TODO: Document systemd service

