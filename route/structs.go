package route

const (
	DatabaseName = "route.sqlite"
)

type Stop struct {
	ID           string
	Name         string
	Lat          float64
	Lon          float64
}
