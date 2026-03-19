package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/aiworkshop/bcapi/pkg/models"
)

// Compile-time interface assertions.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is the plugin instance for a single configured datasource.
type Datasource struct {
	httpClient      *http.Client
	apiUrl          string
	apiKey          string
	resourceHandler backend.CallResourceHandler
}

// NewDatasource creates a new datasource instance from the provided settings.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("load plugin settings: %w", err)
	}

	apiUrl := config.ApiUrl
	if apiUrl == "" {
		apiUrl = models.DefaultApiUrl
	}

	ds := &Datasource{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiUrl:     apiUrl,
		apiKey:     config.Secrets.ApiKey,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stations", ds.handleStations)
	ds.resourceHandler = httpadapter.New(mux)

	return ds, nil
}

// Dispose cleans up datasource instance resources.
func (d *Datasource) Dispose() {}

// QueryData handles multiple queries and returns multiple responses.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}
	return response, nil
}

type queryModel struct {
	QueryType string `json:"queryType"`
	StationID string `json:"stationId"`
}

func (d *Datasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}

	var (
		frames data.Frames
		err    error
	)

	switch qm.QueryType {
	case "station_information":
		frames, err = d.queryStationInformation(ctx, qm.StationID)
	case "station_status":
		frames, err = d.queryStationStatus(ctx, qm.StationID)
	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unknown query type: %q", qm.QueryType))
	}

	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	return backend.DataResponse{Frames: frames}
}

// CallResource forwards resource requests to the registered HTTP mux.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return d.resourceHandler.CallResource(ctx, req, sender)
}

// CheckHealth tests connectivity to the Bicing API.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.apiKey == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "API key is missing",
		}, nil
	}

	if _, err := d.doRequest(ctx, "station_information"); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to reach Bicing API: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
