package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/go-querystring/query"
)

const (
	apiV1 = "https://api.transport.nsw.gov.au/v1/tp"
)

func NewClient(db *sql.DB) *TripClient {
	client := &TripClient{
		db:     db,
		apiKey: os.Getenv("TFNSW_KEY"),
	}

	return client
}

var (
	ErrServerUnavailable      = errors.New("server unavailable")
	ErrServerInternalError    = errors.New("internal error")
	ErrServerNotAuthenticated = errors.New("not authenticated")
)

func (tc *TripClient) fetchData(ctx context.Context, endpoint string, params any) ([]byte, error) {
	values, err := query.Values(params)
	if err != nil {
		log.Error("Error when creating request", "err", err)
		return nil, err
	}

	url := fmt.Sprintf("%s%s?%s", apiV1, endpoint, values.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error("Error when creating request", "err", err)
		return nil, err
	}
	req.Header.Add("Authorization", "apikey "+tc.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when performing request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading body", "err", err)
		return nil, err
	}

	// The application calling the API has not been authenticated.
	if resp.StatusCode == 401 {
		log.Errorf("The application calling the API has not been authenticated: %s", string(body))
		return nil, ErrServerNotAuthenticated
	}

	if resp.StatusCode == 500 {
		log.Errorf("An internal error has occurred: %s", string(body))
		return nil, ErrServerInternalError
	}

	if resp.StatusCode == 503 {
		log.Errorf("The server is currently unavailable: %s", string(body))
		return nil, ErrServerUnavailable

	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("The server returned an unknown status %s", resp.Status)
		return nil, ErrServerInternalError
	}

	// success
	return body, nil
}

// gets only current alerts at the current time
// for the current day
func (tc *TripClient) GetCurrentAlerts(ctx context.Context) ([]Alert, error) {
	now := time.Now().Format("02-01-2006")
	params := alertQuery{
		OutputFormat:            "rapidJSON",
		FilterPublicationStatus: "current",
		Date:                    now,
	}

	data, err := tc.fetchData(ctx, "/add_info", params)
	if err != nil {
		return nil, err
	}

	var parsed alertResponse
	json.Unmarshal(data, &parsed)

	return parsed.Infos.Alerts, nil
}

func (tc *TripClient) TripPlan(ctx context.Context, origin string, destination string) ([]Journey, error) {
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

	data, err := tc.fetchData(ctx, "/trip", params)
	if err != nil {
		return nil, err
	}

	var parsed tripResponse
	json.Unmarshal(data, &parsed)

	return parsed.Journeys, err
}
