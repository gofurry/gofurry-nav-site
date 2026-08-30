import type { OperationTab } from '../../components/admin/operations'

export const metricsTabs: OperationTab[] = [{ key: 'overview', label: '概览' }, { key: 'results', label: '指标结果' }, { key: 'entities', label: '实体状态' }, { key: 'technical', label: '技术契约', capability: 'metrics.technical' }]
export const changesTabs: OperationTab[] = [{ key: 'recent', label: '最近变化' }, { key: 'entities', label: '按实体' }, { key: 'technical', label: '技术契约', capability: 'changes.technical' }]

export function collectionActionVisibility(can: (capability: string) => boolean) {
  return { runNow: can('collection.execute'), manual: can('collection.execute'), control: can('collection.control') }
}
