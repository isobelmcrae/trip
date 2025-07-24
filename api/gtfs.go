package api

import (
	"database/sql"
	"fmt"
)

type GtfsStopTime struct {
	TripID   string
	StopID   string
	Sequence int
	DistanceTraveled float64
}

type GtfsTrip struct {
	TripID  string
	RouteID string
	ShapeID string
}

type GtfsShapePoint struct {
	ShapeID  string
	Lat      float64
	Lon      float64
	Sequence int
	DistanceTraveled float64
}

/*

"stopSequence": [
	{
		"id": "203114",
		"name": "UNSW Gate 2, High St, Randwick",
		"disassembledName": "UNSW Gate 2, High St",
		"arrivalTimePlanned": "",
		"departureTimePlanned": "2025-07-24T08:30:00Z",
		"coord": [
			-33.91527,
			151.228025
		],
		"type": "platform"
	},
	{
		"id": "203346",
		"name": "Anzac Pde before Addison St, Kensington",
		"disassembledName": "Anzac Pde before Addison St",
		"arrivalTimePlanned": "2025-07-24T08:33:00Z",
		"departureTimePlanned": "",
		"coord": [
			-33.912123,
			151.223384
		],
		"type": "platform"
	}
],

*/

/* func (tc *TripClient) FindShapes(originId string, destId string) ([]StopSearchResult, error) {
	rows, err := tc.db.Query(`
		select distinct t.shape_id
			from stop_times as origin_st
			join stop_times as dest_st on origin_st.trip_id = dest_st.trip_id
			join trips as t on t.trip_id = origin_st.trip_id
		where
			origin_st.stop_id = ? and
			dest_st.stop_id = ? and
			origin_st.stop_sequence < dest_st.stop_sequence;
	`, originId, destId)
	if err != nil {
		log.Fatalf("cannot perform search: %v", err)
	}


}
 */

func (tc *TripClient) GetJourneyLeg(originStopID, destStopID string) ([][2]float64, error) {
	var shapeID string
	var startDist, endDist float64

	journeyDefQuery := `
		SELECT
			T.shape_id,
			origin_st.shape_dist_traveled AS start_dist,
			dest_st.shape_dist_traveled AS end_dist
		FROM stop_times AS origin_st
		JOIN stop_times AS dest_st ON origin_st.trip_id = dest_st.trip_id
		JOIN trips AS T ON T.trip_id = origin_st.trip_id
		WHERE
			origin_st.stop_id = ? AND
			dest_st.stop_id = ? AND
			origin_st.stop_sequence < dest_st.stop_sequence AND
			T.shape_id IS NOT NULL AND T.shape_id != '' AND
			origin_st.shape_dist_traveled IS NOT NULL AND dest_st.shape_dist_traveled IS NOT NULL
		LIMIT 1
	`

	err := tc.db.QueryRow(journeyDefQuery, originStopID, destStopID).Scan(&shapeID, &startDist, &endDist)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no direct path with shape data found between stops %s and %s", originStopID, destStopID)
		}
		return nil, fmt.Errorf("error querying for journey definition: %w", err)
	}

	shapeLegQuery := `
        SELECT
            s.shape_pt_lat,
            s.shape_pt_lon
        FROM shapes AS s
        WHERE
            s.shape_id = ? AND
            s.shape_pt_sequence >= (
                SELECT ss.shape_pt_sequence
                FROM shapes AS ss
                WHERE ss.shape_id = ? AND ss.shape_dist_traveled <= ?
                ORDER BY ss.shape_dist_traveled DESC
                LIMIT 1
            ) AND
            s.shape_pt_sequence <= (
                SELECT es.shape_pt_sequence
                FROM shapes AS es
                WHERE es.shape_id = ? AND es.shape_dist_traveled >= ?
                ORDER BY es.shape_dist_traveled ASC
                LIMIT 1
            )
        ORDER BY s.shape_pt_sequence ASC
    `

	rows, err := tc.db.Query(shapeLegQuery, shapeID, shapeID, startDist, shapeID, endDist)
	if err != nil {
		return nil, fmt.Errorf("error querying for shape points for leg %s: %w", shapeID, err)
	}
	defer rows.Close()

	// --- Step 3: Scan the coordinates into the final list. ---
	var path [][2]float64
	for rows.Next() {
		var lat, lon float64
		if err := rows.Scan(&lat, &lon); err != nil {
			return nil, fmt.Errorf("error scanning shape point: %w", err)
		}
		path = append(path, [2]float64{lat, lon})
	}

	return path, nil
}
