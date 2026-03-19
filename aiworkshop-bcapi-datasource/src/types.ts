import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export const QueryType = {
  StationInformation: 'station_information',
  StationStatus: 'station_status',
} as const;

export type QueryTypeValue = (typeof QueryType)[keyof typeof QueryType];

export interface MyQuery extends DataQuery {
  queryType: QueryTypeValue;
  stationId?: string;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  queryType: QueryType.StationInformation,
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  apiUrl?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}
