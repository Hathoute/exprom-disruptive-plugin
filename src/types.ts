import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  entity: string
  parameters: {[key: string]: string}
  withStreaming: boolean;
}

export const defaultQuery: Partial<MyQuery> = {
  entity: "Projects",
  parameters: {
    projects: "-1",
    devices: "-1",
  },
  withStreaming: false,
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  mongodbUrl: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
}

export interface MyVariableQuery {
  entity: string
  projects?: string
}
