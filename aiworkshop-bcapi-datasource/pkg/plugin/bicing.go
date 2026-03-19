package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// bicingResponse is the common envelope returned by both Bicing API endpoints.
type bicingResponse[T any] struct {
	LastUpdated int64 `json:"last_updated"`
	TTL         int   `json:"ttl"`
	Data        struct {
		Stations []T `json:"stations"`
	} `json:"data"`
}

type stationInfo struct {
	StationID             string   `json:"station_id"`
	Name                  string   `json:"name"`
	Lat                   float64  `json:"lat"`
	Lon                   float64  `json:"lon"`
	Altitude              float64  `json:"altitude"`
	Address               string   `json:"address"`
	CrossStreet           string   `json:"cross_street"`
	PostCode              string   `json:"post_code"`
	Capacity              int64    `json:"capacity"`
	PhysicalConfiguration string   `json:"physical_configuration"`
	IsChargingStation     bool     `json:"is_charging_station"`
	RentalMethods         []string `json:"rental_methods"`
}

type stationStatus struct {
	StationID              string `json:"station_id"`
	NumBikesAvailable      int64  `json:"num_bikes_available"`
	NumBikesAvailableTypes struct {
		Mechanical int64 `json:"mechanical"`
		Ebike      int64 `json:"ebike"`
	} `json:"num_bikes_available_types"`
	NumDocksAvailable int64  `json:"num_docks_available"`
	NumBikesDisabled  int64  `json:"num_bikes_disabled"`
	NumDocksDisabled  int64  `json:"num_docks_disabled"`
	Status            string `json:"status"`
	IsInstalled       bool   `json:"is_installed"`
	IsRenting         bool   `json:"is_renting"`
	IsReturning       bool   `json:"is_returning"`
	LastReported      int64  `json:"last_reported"`
}

func (d *Datasource) doRequest(ctx context.Context, path string) ([]byte, error) {
	url := strings.TrimRight(d.apiUrl, "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func (d *Datasource) queryStationInformation(ctx context.Context, stationID string) (data.Frames, error) {
	body, err := d.doRequest(ctx, "station_information")
	if err != nil {
		return nil, err
	}

	var result bicingResponse[stationInfo]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode station_information: %w", err)
	}

	stations := result.Data.Stations
	if stationID != "" {
		filtered := stations[:0]
		for _, s := range stations {
			if s.StationID == stationID {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}

	n := len(stations)
	ids := make([]string, n)
	names := make([]string, n)
	lats := make([]float64, n)
	lons := make([]float64, n)
	alts := make([]float64, n)
	addrs := make([]string, n)
	crosses := make([]string, n)
	postcodes := make([]string, n)
	caps := make([]int64, n)
	configs := make([]string, n)
	charging := make([]bool, n)
	rentals := make([]string, n)

	for i, s := range stations {
		ids[i] = s.StationID
		names[i] = s.Name
		lats[i] = s.Lat
		lons[i] = s.Lon
		alts[i] = s.Altitude
		addrs[i] = s.Address
		crosses[i] = s.CrossStreet
		postcodes[i] = s.PostCode
		caps[i] = s.Capacity
		configs[i] = s.PhysicalConfiguration
		charging[i] = s.IsChargingStation
		rentals[i] = strings.Join(s.RentalMethods, ",")
	}

	frame := data.NewFrame("station_information",
		data.NewField("station_id", nil, ids),
		data.NewField("name", nil, names),
		data.NewField("lat", nil, lats),
		data.NewField("lon", nil, lons),
		data.NewField("altitude", nil, alts),
		data.NewField("address", nil, addrs),
		data.NewField("cross_street", nil, crosses),
		data.NewField("post_code", nil, postcodes),
		data.NewField("capacity", nil, caps),
		data.NewField("physical_configuration", nil, configs),
		data.NewField("is_charging_station", nil, charging),
		data.NewField("rental_methods", nil, rentals),
	)

	return data.Frames{frame}, nil
}

func (d *Datasource) queryStationStatus(ctx context.Context, stationID string) (data.Frames, error) {
	body, err := d.doRequest(ctx, "station_status")
	if err != nil {
		return nil, err
	}

	var result bicingResponse[stationStatus]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode station_status: %w", err)
	}

	stations := result.Data.Stations
	if stationID != "" {
		filtered := stations[:0]
		for _, s := range stations {
			if s.StationID == stationID {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}

	n := len(stations)
	ids := make([]string, n)
	bikesAvail := make([]int64, n)
	mechanical := make([]int64, n)
	ebike := make([]int64, n)
	docksAvail := make([]int64, n)
	bikesDisabled := make([]int64, n)
	docksDisabled := make([]int64, n)
	statuses := make([]string, n)
	installed := make([]bool, n)
	renting := make([]bool, n)
	returning := make([]bool, n)
	lastReported := make([]time.Time, n)

	for i, s := range stations {
		ids[i] = s.StationID
		bikesAvail[i] = s.NumBikesAvailable
		mechanical[i] = s.NumBikesAvailableTypes.Mechanical
		ebike[i] = s.NumBikesAvailableTypes.Ebike
		docksAvail[i] = s.NumDocksAvailable
		bikesDisabled[i] = s.NumBikesDisabled
		docksDisabled[i] = s.NumDocksDisabled
		statuses[i] = s.Status
		installed[i] = s.IsInstalled
		renting[i] = s.IsRenting
		returning[i] = s.IsReturning
		lastReported[i] = time.Unix(s.LastReported, 0).UTC()
	}

	frame := data.NewFrame("station_status",
		data.NewField("station_id", nil, ids),
		data.NewField("num_bikes_available", nil, bikesAvail),
		data.NewField("num_bikes_mechanical", nil, mechanical),
		data.NewField("num_bikes_ebike", nil, ebike),
		data.NewField("num_docks_available", nil, docksAvail),
		data.NewField("num_bikes_disabled", nil, bikesDisabled),
		data.NewField("num_docks_disabled", nil, docksDisabled),
		data.NewField("status", nil, statuses),
		data.NewField("is_installed", nil, installed),
		data.NewField("is_renting", nil, renting),
		data.NewField("is_returning", nil, returning),
		data.NewField("last_reported", nil, lastReported),
	)

	return data.Frames{frame}, nil
}

// handleStations is an HTTP resource handler that returns a list of stations
// for use in the query editor dropdown.
func (d *Datasource) handleStations(w http.ResponseWriter, r *http.Request) {
	body, err := d.doRequest(r.Context(), "station_information")
	if err != nil {
		log.DefaultLogger.Error("handleStations: doRequest failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	var result bicingResponse[stationInfo]
	if err := json.Unmarshal(body, &result); err != nil {
		log.DefaultLogger.Error("handleStations: unmarshal failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type option struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	opts := make([]option, len(result.Data.Stations))
	for i, s := range result.Data.Stations {
		opts[i] = option{Label: s.Name, Value: s.StationID}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
