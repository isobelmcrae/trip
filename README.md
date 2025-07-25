# 🚉 trip

**Sydney public transport, in your terminal.**

![Demo](https://vhs.charm.sh/vhs-4pUxMU7ivISIfNGT3HOav2.gif)

---

## 🛠 Setup

### Requirements

* A recent version of `sqlite3` (tested with ≥ **3.50.2**)
* GTFS data: [**TFNSW Timetables Complete GTFS**](https://opendata.transport.nsw.gov.au/data/dataset/timetables-complete-gtfs)
* A `TFNSW_KEY` environment variable set to your [**TFNSW Open Data Hub API key**](https://opendata.transport.nsw.gov.au/developers/api-basics)

### Build the database

```bash
./makedatabase.sh /path/to/unzipped/gtfs
```

This creates `app.sqlite` with GTFS data.

### Run the app locally

```bash
go build -o trip -tags 'fts5' main.go
./trip
```

---

### SSH Server Mode

`trip` can be run in SSH mode to allow users to connect via `ssh`:

Edit `defaultSSHAddr` in `main.go` to change the port if needed.
Then build with permissions to bind to privileged ports and run in ssh mode:

```bash
go build -o trip -tags 'fts5' main.go
sudo setcap CAP_NET_BIND_SERVICE=+eip trip
./trip --ssh

# then connect from another terminal
ssh user@your.domain.here -p your_port
```
