import type {
  AlertState,
  ApiResult,
  AuthState,
  DeployEvent,
  MetricsRange,
  NodeMetrics,
  OpsNode,
  Overview,
  OverviewMetrics,
  PeerSummary,
  ServiceStatus,
  SyncRun,
} from './types'

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(path, {
    ...init,
    headers,
    credentials: 'include',
  })
  const result = (await response.json()) as ApiResult<T>
  if (!response.ok || result.code !== 1) {
    throw new Error(result.message || '请求失败')
  }
  return result.data
}

export function authState() {
  return request<AuthState>('/api/v1/admin/auth/state')
}

export function login(passcode: string) {
  return request<AuthState>('/api/v1/admin/auth/login', {
    method: 'POST',
    body: JSON.stringify({ passcode }),
  })
}

export function logout() {
  return request<AuthState>('/api/v1/admin/auth/logout', { method: 'POST' })
}

export function overview() {
  return request<Overview>('/api/v1/dashboard/overview')
}

export function overviewMetrics(range: MetricsRange = '1h') {
  return request<OverviewMetrics>(`/api/v1/dashboard/metrics/overview?range=${encodeURIComponent(range)}`)
}

export function nodes() {
  return request<OpsNode[]>('/api/v1/dashboard/nodes')
}

export function node(id: string) {
  return request<OpsNode>(`/api/v1/dashboard/nodes/${encodeURIComponent(id)}`)
}

export function nodeMetrics(id: string, range: MetricsRange = '1h') {
  return request<NodeMetrics>(`/api/v1/dashboard/nodes/${encodeURIComponent(id)}/metrics?range=${encodeURIComponent(range)}`)
}

export function services() {
  return request<ServiceStatus[]>('/api/v1/dashboard/services')
}

export function alerts(active = true) {
  return request<AlertState[]>(`/api/v1/dashboard/alerts?active=${active}`)
}

export function syncRuns(limit = 50) {
  return request<SyncRun[]>(`/api/v1/dashboard/sync-runs?limit=${limit}`)
}

export function peerStatus() {
  return request<PeerSummary[]>('/api/v1/dashboard/peer/status')
}

export function deployments(limit = 50) {
  return request<DeployEvent[]>(`/api/v1/dashboard/deployments?limit=${limit}`)
}
