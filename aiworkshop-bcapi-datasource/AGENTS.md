## Project knowledge

This repository contains a **Grafana plugin**. You must Read @./.config/AGENTS/instructions.md before doing changes.

## Barcelona Bicing API Data Source — Implementation Guide

This plugin is a Grafana backend data source that queries the Barcelona Bicing bike-sharing API.

### API

- Base URL: `https://cc-workshop-proxy.grafana.fun/bcapi/` (configurable)
- Auth: `Authorization: Bearer <token>` header
- Endpoints:
  - `GET /station_information` — static station details (name, lat, lon, capacity, etc.)
  - `GET /station_status` — real-time availability (bikes, docks, status, last_reported)
- Response envelope: `{ "last_updated": int, "ttl": int, "data": { "stations": [...] } }`

### Configuration (plugin settings)

- `jsonData.apiUrl` — base URL for the API proxy (defaults to `https://cc-workshop-proxy.grafana.fun/bcapi/`)
- `secureJsonData.apiKey` — bearer token (stored securely, backend-only)
- Guard against nil/empty `JSONData` in `LoadPluginSettings` — use `if len(source.JSONData) > 0` before unmarshalling

### Backend (Go) — `pkg/`

Key files:
- `pkg/models/settings.go` — `PluginSettings{ApiUrl}`, `SecretPluginSettings{ApiKey}`, `DefaultApiUrl` constant
- `pkg/plugin/datasource.go` — `Datasource` struct, `NewDatasource`, `QueryData`, `CallResource`, `CheckHealth`
- `pkg/plugin/bicing.go` — API structs, `doRequest`, `queryStationInformation`, `queryStationStatus`, `handleStations`

The `Datasource` struct implements:
- `backend.QueryDataHandler` — dispatches on `queryModel.QueryType` (`"station_information"` or `"station_status"`)
- `backend.CheckHealthHandler` — calls `station_information` endpoint to verify connectivity
- `backend.CallResourceHandler` — routes resource requests via `httpadapter.New(mux)` from `github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter`
- `instancemgmt.InstanceDisposer`

Resource handler: `GET /stations` returns `[{"label":"<name>","value":"<station_id>"}]` for use in the query editor dropdown.

Query model:
```go
type queryModel struct {
    QueryType string `json:"queryType"` // "station_information" or "station_status"
    StationID string `json:"stationId"` // optional — filters to a single station
}
```

Data frames return one row per station. `last_reported` is converted to `time.Time`. `rental_methods` is comma-joined.

### Frontend (TypeScript) — `src/`

Key files:
- `src/types.ts` — `QueryType` const object, `MyQuery{queryType, stationId}`, `MyDataSourceOptions{apiUrl}`
- `src/datasource.ts` — extends `DataSourceWithBackend`, adds `getStations()` via `this.getResource('stations')`
- `src/components/QueryEditor.tsx` — `Select` for query type + `Select` with `useEffect` to load stations
- `src/components/ConfigEditor.tsx` — API URL `Input` + API Key `SecretInput`

Query editor pattern for dynamic dropdowns:
```tsx
const [stationOptions, setStationOptions] = useState([]);
useEffect(() => {
  datasource.getStations().then(stations => setStationOptions([...])).catch(console.error);
}, [datasource]);
```

### Build

- Backend: `mage build:backend` — produces `dist/gpx_bcapi_linux_amd64`
- Frontend: `npm run build` — bundles to `dist/module.js`
- After rebuilding the backend, **restart Grafana** to pick up the new binary

### Provisioning

`provisioning/datasources/datasources.yml` sets `jsonData.apiUrl` and `secureJsonData.apiKey` for local dev.
