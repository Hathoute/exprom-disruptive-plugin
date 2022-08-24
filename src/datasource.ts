import {
    DataQueryRequest,
    DataQueryResponse,
    DataSourceInstanceSettings,
    MetricFindValue,
    ScopedVars
} from '@grafana/data';
import {BackendDataSourceResponse, DataSourceWithBackend, getBackendSrv, getTemplateSrv} from '@grafana/runtime';
import {MyDataSourceOptions, MyQuery, MyVariableQuery} from './types';
import {Observable} from "rxjs";

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
    constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
        super(instanceSettings);
    }


    interpolateVariablesInQueries(queries: MyQuery[], scopedVars: ScopedVars | {}): MyQuery[] {
        console.log("interpolateVariablesInQueries", queries, scopedVars)
        return super.interpolateVariablesInQueries(queries, scopedVars);
    }

    applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): Record<string, any> {
        console.log("applyTemplateVariables", query, scopedVars)
        return super.applyTemplateVariables(query, scopedVars);
    }

    async metricFindQuery(query: MyVariableQuery, options?: any): Promise<MetricFindValue[]> {
        const q: Partial<MyQuery> = {
            entity: query.entity,
        }

        if(query.projects) {
            q.parameters = {
                ...q.parameters,
                projects: getTemplateSrv().replace(query.projects, undefined, 'csv')
            }
        }

        console.log("metricFindQuery", q)

        const response = await this._internalQuery(q);
        console.log("postMetricQuery", response)
        const results = this._formatResults(response);

        return results.map(r => ({
                text: r.name,
                value: r.id
            })
        )
    }

    query(request: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
        console.log("query", request);

        for (const target of request.targets) {
            if(target.parameters.projects) {
                target.parameters.projects = getTemplateSrv()
                    .replace(target.parameters.projects, undefined, 'csv')
            }
            if(target.parameters.devices) {
                target.parameters.devices = getTemplateSrv()
                    .replace(target.parameters.devices, undefined, 'csv')
            }
        }

        return super.query(request);
    }

    async _internalQuery(pq: Partial<MyQuery>): Promise<BackendDataSourceResponse> {
        const q = this._buildQuery(pq);
        const result = await getBackendSrv().datasourceRequest<BackendDataSourceResponse>({
            url: 'api/ds/query',
            method: 'POST',
            data: {
                from: '5m',
                to: 'now',
                queries: [{
                    ...q,
                    datasource: this.name,
                    datasourceId: this.id
                }],
            },

        })

        return result.data
    }

    _buildQuery(pq: Partial<MyQuery>): MyQuery {
        return {
            entity: pq.entity ?? "Projects",
            parameters: pq.parameters ?? {},
            refId: pq.refId ?? "ref",
            withStreaming: pq.withStreaming ?? false
        }
    }

    _formatResults(response: BackendDataSourceResponse): any[] {
        const frames = response.results.ref.frames
        if(!frames || frames.length === 0) {
            throw new Error("No frames found");
        }

        const frame = frames[0];
        const result = [];
        for (let i = 0; i < frame.data?.values[0].length!; i++) {
            const obj: any = {}
            for(let j = 0; j < frame.schema?.fields.length!; j++) {
                const field = frame.schema?.fields[j]!;
                const value = frame.data?.values[j][i];
                let tValue: any;
                switch (field.type) {
                    case "number":
                        tValue = parseInt(value, 10)
                        break;
                    default:
                        tValue = value;
                }
                obj[field.name] = tValue;
            }
            result.push(obj);
        }

        return result
    }
}
