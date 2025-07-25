# Server Deployment

TODO: add docs

Build the thing and allow it to bind to reserved ports (i.e port 22)

```bash
go build -o trip -tags 'fts5' main.go
sudo setcap CAP_NET_BIND_SERVICE=+eip trip
```

TODO: document systemd service
