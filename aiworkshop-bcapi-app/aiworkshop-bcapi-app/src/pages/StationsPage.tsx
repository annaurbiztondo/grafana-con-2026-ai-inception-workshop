import React, { useEffect, useState } from 'react';
import { css } from '@emotion/css';
import { DataFrame, DataSourceInstanceSettings, GrafanaTheme2 } from '@grafana/data';
import { getBackendSrv, getDataSourceSrv, PluginPage, toDataQueryResponse } from '@grafana/runtime';
import { Combobox, Spinner, Tooltip, useStyles2 } from '@grafana/ui';
import { lastValueFrom } from 'rxjs';

interface Station {
  station_id: string;
  name: string;
  lat: number;
  lon: number;
  altitude: number;
  address: string;
  cross_street: string;
  post_code: string;
  capacity: number;
  physical_configuration: string;
  is_charging_station: boolean;
  rental_methods: string;
}

function frameToStations(frame: DataFrame): Station[] {
  const vals = (name: string): any[] => frame.fields.find((f) => f.name === name)?.values ?? [];

  const station_ids = vals('station_id');
  const names = vals('name');
  const lats = vals('lat');
  const lons = vals('lon');
  const altitudes = vals('altitude');
  const addresses = vals('address');
  const cross_streets = vals('cross_street');
  const post_codes = vals('post_code');
  const capacities = vals('capacity');
  const physical_configurations = vals('physical_configuration');
  const is_charging_stations = vals('is_charging_station');
  const rental_methods_list = vals('rental_methods');

  const stations: Station[] = [];
  for (let i = 0; i < frame.length; i++) {
    stations.push({
      station_id: station_ids[i] ?? '',
      name: names[i] ?? '',
      lat: lats[i] ?? 0,
      lon: lons[i] ?? 0,
      altitude: altitudes[i] ?? 0,
      address: addresses[i] ?? '',
      cross_street: cross_streets[i] ?? '',
      post_code: post_codes[i] ?? '',
      capacity: capacities[i] ?? 0,
      physical_configuration: physical_configurations[i] ?? '',
      is_charging_station: is_charging_stations[i] ?? false,
      rental_methods: rental_methods_list[i] ?? '',
    });
  }
  return stations;
}

function StationDetails({ station }: { station: Station }) {
  const s = useStyles2(getStyles);
  return (
    <div className={s.tooltipContent}>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>ID</span>
        <span>{station.station_id}</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Address</span>
        <span>{station.address}</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>District</span>
        <span>{station.cross_street}</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Post code</span>
        <span>{station.post_code}</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Capacity</span>
        <span>{station.capacity} docks</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Coordinates</span>
        <span>
          {station.lat.toFixed(5)}, {station.lon.toFixed(5)}
        </span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Altitude</span>
        <span>{station.altitude} m</span>
      </div>
      <div className={s.tooltipRow}>
        <span className={s.tooltipLabel}>Type</span>
        <span>{station.physical_configuration}</span>
      </div>
      {station.is_charging_station && (
        <div className={s.tooltipRow}>
          <span>⚡ Charging station</span>
        </div>
      )}
      {station.rental_methods && (
        <div className={s.tooltipRow}>
          <span className={s.tooltipLabel}>Rental</span>
          <span>{station.rental_methods}</span>
        </div>
      )}
    </div>
  );
}

function StationsPage() {
  const s = useStyles2(getStyles);
  const [datasources, setDatasources] = useState<DataSourceInstanceSettings[]>([]);
  const [selectedDs, setSelectedDs] = useState<DataSourceInstanceSettings | null>(null);
  const [stations, setStations] = useState<Station[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const dsList = getDataSourceSrv().getList({ type: 'aiworkshop-bcapi-datasource' });
    setDatasources(dsList);
    if (dsList.length > 0) {
      setSelectedDs(dsList[0]);
    }
  }, []);

  useEffect(() => {
    if (!selectedDs) {
      return;
    }

    let cancelled = false;

    const fetchStations = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await lastValueFrom(
          getBackendSrv().fetch<any>({
            url: '/api/ds/query',
            method: 'POST',
            data: {
              queries: [
                {
                  refId: 'A',
                  datasource: { uid: selectedDs.uid },
                  queryType: 'station_information',
                },
              ],
              from: 'now-5m',
              to: 'now',
            },
          })
        );

        if (cancelled) {
          return;
        }

        const queryResponse = toDataQueryResponse(response, []);
        const frame = queryResponse.data[0] as DataFrame | undefined;
        setStations(frame ? frameToStations(frame) : []);
      } catch (e: unknown) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : String(e));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    fetchStations();
    return () => {
      cancelled = true;
    };
  }, [selectedDs]);

  const dsOptions = datasources.map((ds) => ({ label: ds.name, value: ds.uid }));

  return (
    <PluginPage>
      <div className={s.controls}>
        <Combobox
          options={dsOptions}
          value={selectedDs?.uid ?? null}
          onChange={(opt) => {
            const ds = datasources.find((d) => d.uid === opt?.value);
            setSelectedDs(ds ?? null);
          }}
          placeholder="Select Bicing datasource…"
          width={40}
        />
      </div>

      {loading && (
        <div className={s.centered}>
          <Spinner size={24} />
        </div>
      )}

      {error && <div className={s.error}>{error}</div>}

      {!loading && !error && stations.length > 0 && (
        <ul className={s.list}>
          {stations.map((station) => (
            <Tooltip key={station.station_id} content={<StationDetails station={station} />} placement="right">
              <li className={s.listItem}>
                <span className={s.stationName}>{station.name}</span>
                <span className={s.stationMeta}>
                  {station.address}
                  {station.cross_street ? ` · ${station.cross_street}` : ''}
                </span>
              </li>
            </Tooltip>
          ))}
        </ul>
      )}

      {!loading && !error && stations.length === 0 && selectedDs && (
        <div className={s.empty}>No stations found.</div>
      )}

      {!selectedDs && datasources.length === 0 && (
        <div className={s.empty}>No Bicing datasource configured. Add one in the datasource settings.</div>
      )}
    </PluginPage>
  );
}

export default StationsPage;

const getStyles = (theme: GrafanaTheme2) => ({
  controls: css`
    margin-bottom: ${theme.spacing(2)};
  `,
  centered: css`
    display: flex;
    justify-content: center;
    padding: ${theme.spacing(4)};
  `,
  list: css`
    list-style: none;
    margin: 0;
    padding: 0;
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    overflow: hidden;
  `,
  listItem: css`
    display: flex;
    flex-direction: column;
    padding: ${theme.spacing(1, 2)};
    border-bottom: 1px solid ${theme.colors.border.weak};
    cursor: default;
    &:last-child {
      border-bottom: none;
    }
    &:hover {
      background: ${theme.colors.action.hover};
    }
  `,
  stationName: css`
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  stationMeta: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    margin-top: ${theme.spacing(0.25)};
  `,
  tooltipContent: css`
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(0.5)};
    min-width: 240px;
    padding: ${theme.spacing(0.5)};
  `,
  tooltipRow: css`
    display: flex;
    gap: ${theme.spacing(1)};
  `,
  tooltipLabel: css`
    color: ${theme.colors.text.secondary};
    min-width: 80px;
    flex-shrink: 0;
  `,
  error: css`
    color: ${theme.colors.error.text};
    padding: ${theme.spacing(2)};
  `,
  empty: css`
    color: ${theme.colors.text.secondary};
    padding: ${theme.spacing(2)};
  `,
});
