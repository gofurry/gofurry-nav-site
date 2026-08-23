import type { CollectorEnvelope } from '@/types/nav'

export type DetailInfoItem = {
  label: string
  value: string | string[]
}

export type DetailSection = {
  title: string
  items: DetailInfoItem[]
}

export type LightProbeEntry = {
  protocol: string
  status: string
  payload: unknown
  observedAt: string
  durationMs: number
  errorCode: string
  errorMessage: string
  items: DetailInfoItem[]
}

export type SiteHeroBadge = {
  label: string
  class: string
}

export type SiteSignalCard = {
  eyebrow: string
  title: string
  value: string
  badge: string
  tone: string
  items: { label: string; value: string }[]
}

export type ObservationStripItem = {
  label: string
  value: string
  tone: string
}

export type ProtocolTrackEntry = {
  protocol: string
  label: string
  status: string
  duration: string
  observedAt: string
  staleAfter: string
  tone: string
}

export type SecurityHeaderItem = {
  label: string
  ok: boolean
}

export type ChangeEventItem = {
  key: string
  protocol: string
  field: string
  oldValue: string
  newValue: string
  detectedAt: string
}

export type ObservationProtocol = 'ping' | 'http' | 'dns'

export type ObservationHistoryItem = {
  protocol: ObservationProtocol
  title: string
  items: CollectorEnvelope[]
  visibleItems: CollectorEnvelope[]
  page: number
  totalPages: number
}

export type ObservationTone = 'good' | 'normal' | 'warn' | 'warm' | 'sky' | 'mint' | 'amber' | 'rose' | 'violet' | 'lime' | 'peach'

export type ObservationMetricItem = {
  label: string
  value: string
  accent?: boolean
  tone?: ObservationTone
}

export type ObservationInfoItem = {
  label: string
  value: string
  tone?: ObservationTone
}
