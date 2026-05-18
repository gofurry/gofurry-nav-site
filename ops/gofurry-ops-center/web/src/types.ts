export type ApiResult<T> = {
  code: number
  message: string
  data: T
}

export type AuthState = {
  authenticated: boolean
  initialized: boolean
}

export type OpsNode = {
  node_id: string
  region: string
  role: string
  display_name: string
  status: string
  agent_version: string
  last_seen_at?: string
  updated_at: string
}

export type ServiceStatus = {
  key: string
  node_id: string
  service_type: string
  name: string
  status: string
  message?: string
  failure_count: number
  latency_ms?: number
  last_ok_at?: string
  updated_at: string
}

export type AlertState = {
  key: string
  region: string
  node_id?: string
  level: string
  type: string
  title: string
  message?: string
  status: string
  first_seen_at: string
  last_seen_at: string
  resolved_at?: string
}

export type PeerSummary = {
  region: string
  center_id: string
  status: string
  last_heartbeat_at?: string
  nodes_total: number
  nodes_down: number
  critical_alerts: number
  warning_alerts: number
  last_sync_status?: string
  updated_at: string
}

export type SyncRun = {
  id: number
  region: string
  sync_name: string
  version?: string
  status: string
  items_total: number
  checksum_ok?: boolean
  error_message?: string
  started_at?: string
  finished_at?: string
  created_at: string
}

export type DeployEvent = {
  id: number
  region: string
  node_id?: string
  service_name: string
  version?: string
  status: string
  message?: string
  created_at: string
}

export type Overview = {
  center_id: string
  region: string
  status: string
  nodes_total: number
  nodes_down: number
  critical_alerts: number
  warning_alerts: number
  last_heartbeat_at?: string
  last_sync?: SyncRun
  peer?: PeerSummary
  services: ServiceStatus[]
  alerts: AlertState[]
}
