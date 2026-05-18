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

export type MetricsRange = '1h' | '6h' | '24h'

export type MetricPoint = {
  timestamp: string
  value: number
}

export type MetricSeries = {
  name: string
  unit?: string
  points: MetricPoint[]
}

export type StatusCount = {
  name: string
  count: number
}

export type TopResource = {
  node_id: string
  name?: string
  value: number
  unit?: string
  updated_at: string
}

export type SystemAggregate = {
  nodes_reported: number
  cpu_avg: number
  memory_avg: number
  load1_avg: number
  reported_at?: string
}

export type OverviewMetrics = {
  center_id: string
  region: string
  range: MetricsRange
  generated_at: string
  last_sample_at?: string
  sample_freshness_seconds: number
  latest_system: SystemAggregate
  highest_disk?: TopResource
  service_status_counts: StatusCount[]
  alert_level_counts: StatusCount[]
  top_cpu: TopResource[]
  top_memory: TopResource[]
  top_disk: TopResource[]
  cpu_trend: MetricPoint[]
  memory_trend: MetricPoint[]
  load_trend: MetricPoint[]
  disk_trend: MetricPoint[]
  latency_trend: MetricPoint[]
}

export type SystemSnapshot = {
  cpu_usage: number
  memory_usage: number
  memory_used: number
  memory_total: number
  load1: number
  load5: number
  load15: number
  uptime_seconds: number
  reported_at: string
}

export type DiskSnapshot = {
  mount: string
  usage: number
  inode_usage: number
  used: number
  total: number
  reported_at: string
}

export type NetworkSnapshot = {
  name: string
  bytes_sent: number
  bytes_recv: number
  packets_sent: number
  packets_recv: number
  tx_bytes_per_sec: number
  rx_bytes_per_sec: number
  reported_at: string
}

export type DockerSnapshot = {
  name: string
  running: boolean
  status: string
  health_status?: string
  restart_count: number
  error_message?: string
  reported_at: string
}

export type HTTPCheckSnapshot = {
  name: string
  url?: string
  status: string
  status_code?: number
  latency_ms: number
  error_message?: string
  reported_at: string
}

export type ServiceCheckSnapshot = {
  service_type: string
  name: string
  status: string
  latency_ms: number
  error_message?: string
  database_size?: number
  connections?: number
  memory_used?: number
  key_count?: number
  reported_at: string
}

export type CertSnapshot = {
  name: string
  host: string
  status: string
  expires_at?: string
  days_remaining: number
  matched_name: boolean
  error_message?: string
  reported_at: string
}

export type NodeLatestMetrics = {
  system?: SystemSnapshot
  disks: DiskSnapshot[]
  networks: NetworkSnapshot[]
  docker: DockerSnapshot[]
  http_checks: HTTPCheckSnapshot[]
  service_checks: ServiceCheckSnapshot[]
  certs: CertSnapshot[]
}

export type NodeTrendMetrics = {
  cpu: MetricPoint[]
  memory: MetricPoint[]
  load: MetricPoint[]
  network_rx: MetricPoint[]
  network_tx: MetricPoint[]
  disk_usage: MetricSeries[]
  service_latency: MetricSeries[]
}

export type NodeMetrics = {
  node: OpsNode
  range: MetricsRange
  generated_at: string
  last_sample_at?: string
  sample_freshness_seconds: number
  latest: NodeLatestMetrics
  trends: NodeTrendMetrics
}
