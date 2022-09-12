import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  queryText?: string;
}


export const defaultQuery: Partial<MyQuery> = {
  queryText: 'select current timestamp - 20 minutes as timeseries, 10 as value from sysibm.sysdummy1',
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  host?: string;
  port?: string;
  database?: string;
  user?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  password?: string;
}
