package route

/*

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	// "time"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/google/go-querystring/query"
)


func NewClient(baseURL string) *TripClient {
    client := &TripClient{
        BaseURL:    baseURL,
        apiKey:     os.Getenv("TFNSW_KEY"),
        httpClient: &http.Client{},
    }

    return client
}

// grabs data from tfnsw base url in a TripClient + endpoint & params
func (tc *TripClient) fetchData(endpoint string, params any) ([]byte, error) {
    values, _ := query.Values(params)
    url := fmt.Sprintf("%s%s?%s", tc.BaseURL, endpoint, values.Encode())

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Error("Error when creating request", "err", err)
        return nil, err
    }

    req.Header.Add("Authorization", tc.apiKey)

    resp, err := tc.httpClient.Do(req)
    if err != nil {
        log.Error("Error when performing request", "err", err)
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Error("Non-200 response", "response", resp.Status)
        return nil, err
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Error("Error reading body", "err", err)
        return nil, err
    }

    return body, nil
}

// Search for stops from a given string
func (tc *TripClient) FindStop(name string) ([]Stop, error) {
    params := stopQuery{
        OutputFormat:      "rapidJSON",
        TypeSf:            "stop",
        NameSf:            name,
        CoordOutputFormat: "EPSG:4326",
    }

    data, err := tc.fetchData("/stop_finder", params)
    if err != nil {
        return nil, err
    }

    var parsed stopResponse
    json.Unmarshal(data, &parsed)

    return parsed.Stops, nil
}

// gets only current alerts at the current time
// for the current day
func (tc *TripClient) GetCurrentAlerts() ([]Alert, error) {
    now := time.Now().Format("02-01-2006")
    params := alertQuery{
        OutputFormat:            "rapidJSON",
        FilterPublicationStatus: "current",
        Date:                    now,
    }

    data, err := tc.fetchData("/add_info", params)
    if err != nil {
        return nil, err
    }

    var parsed alertResponse
    json.Unmarshal(data, &parsed)

    return parsed.Infos.Alerts, nil
}

// takes two stop ids and routes them
// TODO: add date parameter to see if we stop
// getting old data (~approx 10min old)
func (tc *TripClient) TripPlan(origin string, destination string) ([]Journey, error) {
    params := tripQuery{
        OutputFormat:      "rapidJSON",
        CoordOutputFormat: "EPSG:4326",
        DepArrMacro:       "dep", // trips departing now
        TypeOrigin:        "any",
        OriginID:          origin,
        TypeDestination:   "any",
        DestinationID:     destination,
        ExcludedMeans:     "11", // exclude school buses
    }

    data, err := tc.fetchData("/trip", params)
    if err != nil {
        return nil, err
    }

    var parsed tripResponse
    json.Unmarshal(data, &parsed)

    return parsed.Journeys, err
}
*/