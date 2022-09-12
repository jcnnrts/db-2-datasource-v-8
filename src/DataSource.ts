import { DataSourceInstanceSettings, ScopedVars, DataQueryRequest, DataQueryResponse, DataFrame } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { getTemplateSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars) {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      queryText: query.queryText ? templateSrv.replace(query.queryText, scopedVars) : '',
    };
  }

  async metricFindQuery(query: MyQuery, scopedVars: ScopedVars, options?: any) {
    const templateSrv = getTemplateSrv();

    const request = {
      targets: [
        {
          queryText: query.queryText ? templateSrv.replace(query.queryText, scopedVars) : '',
          refId: 'metricFindQuery',
        }
      ]
    } as DataQueryRequest<MyQuery>

    let res: DataQueryResponse;

    try {
      res = await this.query(request).toPromise();
    } catch(err){
      return Promise.reject(err);
    }

    if (!res.data.length || !res.data[0].fields.length){
      return [];
    }

    return (res.data[0] as DataFrame).fields[0].values.toArray().map((_) => ({ text: _.toString() }));
  }
  
}
