import { from, Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import {
  CustomVariableSupport,
  DataQueryRequest,
  DataQueryResponse,
  MetricFindValue,
  SelectableValue,
} from '@grafana/data';

import { ALL_ACCOUNTS_OPTION } from './components/Account';
import { VariableQueryEditor } from './components/VariableQueryEditor/VariableQueryEditor';
import { CloudWatchDatasource } from './datasource';
import { DEFAULT_VARIABLE_QUERY } from './defaultQueries';
import { migrateVariableQuery } from './migrations/variableQueryMigrations';
import { ResourcesAPI } from './resources/ResourcesAPI';
import { standardStatistics } from './standardStatistics';
import { VariableQuery, VariableQueryType } from './types';

export class CloudWatchVariableSupport extends CustomVariableSupport<CloudWatchDatasource, VariableQuery> {
  constructor(private readonly resources: ResourcesAPI) {
    super();
    this.query = this.query.bind(this);
  }

  editor = VariableQueryEditor;

  query(request: DataQueryRequest<VariableQuery>): Observable<DataQueryResponse> {
    const queryObj = migrateVariableQuery(request.targets[0]);
    return from(this.execute(queryObj)).pipe(map((data) => ({ data })));
  }

  async execute(query: VariableQuery) {
    try {
      switch (query.queryType) {
        case VariableQueryType.Regions:
          return this.handleRegionsQuery();
        case VariableQueryType.Namespaces:
          return this.handleNamespacesQuery();
        case VariableQueryType.Metrics:
          return this.handleMetricsQuery(query);
        case VariableQueryType.DimensionKeys:
          return this.handleDimensionKeysQuery(query);
        case VariableQueryType.DimensionValues:
          return this.handleDimensionValuesQuery(query);
        case VariableQueryType.EBSVolumeIDs:
          return this.handleEbsVolumeIdsQuery(query);
        case VariableQueryType.EC2InstanceAttributes:
          return this.handleEc2InstanceAttributeQuery(query);
        case VariableQueryType.ResourceArns:
          return this.handleResourceARNsQuery(query);
        case VariableQueryType.Statistics:
          return this.handleStatisticsQuery();
        case VariableQueryType.LogGroups:
          return this.handleLogGroupsQuery(query);
        case VariableQueryType.Accounts:
          return this.handleAccountsQuery(query);
      }
    } catch (error) {
      console.error(`Could not run CloudWatchMetricFindQuery ${query}`, error);
      return [];
    }
  }
  async handleLogGroupsQuery({ region, logGroupPrefix }: VariableQuery) {
    return this.resources
      .getLogGroups({
        region,
        logGroupNamePrefix: logGroupPrefix,
        listAllLogGroups: true,
      })
      .then((logGroups) =>
        logGroups.map((lg) => {
          return {
            text: lg.value.name,
            value: lg.value.arn,
            expandable: true,
          };
        })
      );
  }

  async handleRegionsQuery() {
    return this.resources.getRegions().then((regions) => regions.map(selectableValueToMetricFindOption));
  }

  async handleNamespacesQuery() {
    return this.resources.getNamespaces().then((namespaces) => namespaces.map(selectableValueToMetricFindOption));
  }

  async handleMetricsQuery({ namespace, region }: VariableQuery) {
    return this.resources
      .getMetrics({ namespace, region })
      .then((metrics) => metrics.map(selectableValueToMetricFindOption));
  }

  async handleDimensionKeysQuery({ namespace, region }: VariableQuery) {
    return this.resources
      .getDimensionKeys({ namespace, region })
      .then((keys) => keys.map(selectableValueToMetricFindOption));
  }

  async handleDimensionValuesQuery({ namespace, region, dimensionKey, metricName, dimensionFilters }: VariableQuery) {
    if (!dimensionKey || !metricName) {
      return [];
    }
    return this.resources
      .getDimensionValues({
        region,
        namespace,
        metricName,
        dimensionKey,
        dimensionFilters,
      })
      .then((values) => values.map(selectableValueToMetricFindOption));
  }

  async handleEbsVolumeIdsQuery({ region, instanceID }: VariableQuery) {
    if (!instanceID) {
      return [];
    }
    return this.resources.getEbsVolumeIds(region, instanceID).then((ids) => ids.map(selectableValueToMetricFindOption));
  }

  async handleEc2InstanceAttributeQuery({ region, attributeName, ec2Filters }: VariableQuery) {
    if (!attributeName) {
      return [];
    }
    return this.resources
      .getEc2InstanceAttribute(region, attributeName, ec2Filters ?? {})
      .then((values) => values.map(selectableValueToMetricFindOption));
  }

  async handleResourceARNsQuery({ region, resourceType, tags }: VariableQuery) {
    if (!resourceType) {
      return [];
    }
    const keys = await this.resources.getResourceARNs(region, resourceType, tags ?? {});
    return keys.map(selectableValueToMetricFindOption);
  }

  async handleStatisticsQuery() {
    return standardStatistics.map((s: string) => ({
      text: s,
      value: s,
      expandable: true,
    }));
  }

  allMetricFindValue: MetricFindValue = { text: 'All', value: ALL_ACCOUNTS_OPTION.value, expandable: true };
  async handleAccountsQuery({ region }: VariableQuery) {
    return this.resources.getAccounts({ region }).then((accounts) => {
      const metricFindOptions = accounts.map((account) => ({
        text: account.label,
        value: account.id,
        expandable: true,
      }));

      return metricFindOptions.length ? [this.allMetricFindValue, ...metricFindOptions] : [];
    });
  }

  getDefaultQuery(): Partial<VariableQuery> {
    return DEFAULT_VARIABLE_QUERY;
  }
}

function selectableValueToMetricFindOption({ label, value }: SelectableValue<string>): MetricFindValue {
  return { text: label ?? value ?? '', value: value, expandable: true };
}
