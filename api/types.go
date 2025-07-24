package api

import "database/sql"

// TripClientV1, for v1 of the API
type TripClient struct {
	db     *sql.DB // route searching
	apiKey string
}

// stops
/* type stopQuery struct {
    OutputFormat      string `url:"outputFormat"`
    TypeSf            string `url:"type_sf"`
    NameSf            string `url:"name_sf"`
    CoordOutputFormat string `url:"coordOutputFormat"`
}

type stopResponse struct {
    Stops []Stop `json:"locations"`
} */

type Stop struct {
	ID           string `json:"id"`
	Name         string `json:"disassembledName"`
	MatchQuality int    `json:"matchQuality"`
	Modes        []int  `json:"modes"`
}

type alertResponse struct {
	Infos struct {
		Alerts []Alert `json:"current"`
	} `json:"infos"`
}

type Alert struct {
	Content  string        `json:"content"`
	ID       string        `json:"id"`
	Priority string        `json:"priority"`
	URL      string        `json:"url"`
	URLText  string        `json:"urlText"`
	Type     string        `json:"type"`
	Affected AffectedItems `json:"affected"`
}

type AffectedItems struct {
	Lines []AffectedLines `json:"lines"`
	Stops []AffectedStops `json:"stops"`
}

type AffectedLines struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
}

type AffectedStops struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type alertQuery struct {
	OutputFormat            string `url:"outputFormat"`
	FilterPublicationStatus string `url:"filterPublicationStatus"`
	Date                    string `url:"filterDateValid"`
}

type tripQuery struct {
	OutputFormat      string `url:"outputFormat"`
	CoordOutputFormat string `url:"coordOutputFormat"`
	DepArrMacro       string `url:"depArrMacro"`
	TypeOrigin        string `url:"type_origin"`
	OriginID          string `url:"name_origin"`
	TypeDestination   string `url:"type_destination"`
	DestinationID     string `url:"name_destination"`
	ExcludedMeans     string `url:"excludedMeans"`
	Time              string `url:"itdTime"`
}

type tripResponse struct {
	Journeys []Journey `json:"journeys"`
}

type Journey struct {
	IsAdditional bool  `json:"isAdditional"` // indicates it's not the "preferred" journey
	Legs         []Leg `json:"legs"`
	Rating       int   `json:"rating"`
}

type Leg struct {
	Origin               Location        `json:"origin"`
	Destination          Location        `json:"destination"`
	Duration             int             `json:"duration"`
	Distance             int             `json:"distance"`
	Transportation       *Transportation `json:"transportation"`
	StopSequence         []JourneyStop   `json:"stopSequence"`
	IsRealtimeControlled bool            `json:"isRealtimeControlled"`
}

type Location struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	DisassembledName       string    `json:"disassembledName"`
	ArrivalTimePlanned     string    `json:"arrivalTimePlanned"`
	ArrivalTimeEstimated   string    `json:"arrivalTimeEstimated"`
	DepartureTimePlanned   string    `json:"departureTimePlanned"`
	DepartureTimeEstimated string    `json:"departureTimeEstimated"`
	Coord                  []float64 `json:"coord"`
	Type                   string    `json:"type"`
}

type JourneyStop struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	DisassembledName     string    `json:"disassembledName"`
	ArrivalTimePlanned   string    `json:"arrivalTimePlanned"`
	DepartureTimePlanned string    `json:"departureTimePlanned"`
	Coord                []float64 `json:"coord"`
	Type                 string    `json:"type"`
}

type Transportation struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Number           string      `json:"number"`
	Description      string      `json:"description"`
	DisassembledName string      `json:"disassembledName"`
	Destination      Destination `json:"destination"`
	IconID           int         `json:"iconId"`
}

type Destination struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
