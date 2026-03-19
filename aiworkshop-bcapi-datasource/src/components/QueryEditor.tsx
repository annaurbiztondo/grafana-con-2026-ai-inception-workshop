import React, { useEffect, useState } from 'react';
import { InlineField, Select, Stack } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery, QueryType, QueryTypeValue } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const queryTypeOptions: Array<SelectableValue<QueryTypeValue>> = [
  { label: 'Station Information', value: QueryType.StationInformation },
  { label: 'Station Status', value: QueryType.StationStatus },
];

const ALL_STATIONS_OPTION: SelectableValue<string> = { label: 'All stations', value: '' };

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [stationOptions, setStationOptions] = useState<Array<SelectableValue<string>>>([ALL_STATIONS_OPTION]);
  const [isLoadingStations, setIsLoadingStations] = useState(false);

  useEffect(() => {
    setIsLoadingStations(true);
    datasource
      .getStations()
      .then((stations) => {
        setStationOptions([
          ALL_STATIONS_OPTION,
          ...stations.map((s) => ({ label: s.label, value: s.value })),
        ]);
      })
      .catch(console.error)
      .finally(() => setIsLoadingStations(false));
  }, [datasource]);

  const onQueryTypeChange = (selected: SelectableValue<QueryTypeValue>) => {
    onChange({ ...query, queryType: selected.value! });
    onRunQuery();
  };

  const onStationChange = (selected: SelectableValue<string>) => {
    onChange({ ...query, stationId: selected?.value ?? '' });
    onRunQuery();
  };

  const selectedStation = stationOptions.find((o) => o.value === (query.stationId ?? '')) ?? ALL_STATIONS_OPTION;

  return (
    <Stack gap={0}>
      <InlineField label="Query Type" labelWidth={16}>
        <Select
          inputId="query-editor-query-type"
          options={queryTypeOptions}
          value={query.queryType}
          onChange={onQueryTypeChange}
          width={24}
        />
      </InlineField>
      <InlineField label="Station" labelWidth={16} tooltip="Filter results to a single station">
        <Select
          inputId="query-editor-station"
          options={stationOptions}
          value={selectedStation}
          onChange={onStationChange}
          isLoading={isLoadingStations}
          width={40}
          placeholder="All stations"
        />
      </InlineField>
    </Stack>
  );
}
