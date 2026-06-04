// types/nav.ts

// 站点结构
export interface Site {
    id: string;
    name: string;
    domain: string;
    info: string;
    country: string | null;
    nsfw: string;
    welfare: string;
    icon: string | null;
}

// 分组结构
export interface Group {
    id: string;
    name: string;
    info: string;
    priority: number;
    sites: string[];
}

// 延迟结构
export interface Delay {
    status: string;
    time: string;
    loss: string;
    delay: string;
}

export interface SiteInfo {
    name: string;
    info: string;
    icon: string | null;
    country: string | null;
    nsfw: string;
    welfare: string;
    view_count: number;
}

export type HealthSummaryState = 'ready' | 'missing' | 'stale';
export type HealthStatus = 'healthy' | 'warning' | 'degraded' | 'unknown' | 'down';

export interface ProtocolHealthSummary {
    protocol: 'ping' | 'http' | 'dns' | string;
    status: string;
    observed_at: string;
    duration_ms: number;
    stale: boolean;
    stale_after_seconds: number;
    error_code?: string;
}

export interface TargetHealthSummaryItem {
    target: string;
    status: HealthStatus;
    reason_codes: string[];
    reason_messages: string[];
    canonical_target_hint?: CanonicalTargetHint | null;
    target_relation_hints?: TargetRelationHint[];
    edge_provider_hints?: EdgeProviderHint[];
    observed_at: string;
}

export interface SiteHealthSummary {
    state: HealthSummaryState;
    site_id: number;
    status: HealthStatus;
    reason_codes: string[];
    reason_messages: string[];
    target_count: number;
    status_counts: Record<string, number>;
    targets: TargetHealthSummaryItem[];
    target_relation_hints?: SiteTargetRelationHint[];
    generated_at: string;
    schema_version: number;
}

export interface TargetHealthSummary {
    state: HealthSummaryState;
    site_id: number;
    target: string;
    status: HealthStatus;
    reason_codes: string[];
    reason_messages: string[];
    protocols: Record<string, ProtocolHealthSummary>;
    canonical_target_hint?: CanonicalTargetHint | null;
    target_relation_hints?: TargetRelationHint[];
    edge_provider_hints?: EdgeProviderHint[];
    observed_at: string;
    generated_at: string;
    schema_version: number;
}

export interface CanonicalTargetHint {
    target_host: string;
    final_host?: string;
    canonical_host?: string;
    preferred_host?: string;
    relation: string;
    source: string;
    final_url?: string;
    canonical_url?: string;
}

export interface TargetRelationHint {
    relation: string;
    source: string;
    target_host: string;
    related_host?: string;
    value?: string;
}

export interface SiteTargetRelationHint {
    relation: string;
    host: string;
    targets: string[];
}

export interface EdgeProviderHint {
    provider: string;
    hint_type: string;
    confidence: string;
    evidence: EdgeProviderEvidence[];
}

export interface EdgeProviderEvidence {
    source: string;
    field: string;
    value: string;
}

export interface SiteV2Info {
    id: number;
    name: string;
    info: string;
    icon: string | null;
    country: string | null;
    nsfw: string;
    welfare: string;
    view_count: number;
}

export interface SiteV2Target {
    target: string;
    domain_id: number;
    name: string;
    prefix?: string | null;
    tls: string;
    proxy: string;
    source: string;
    registered: boolean;
    summary_only: boolean;
    summary_state: string;
    status: string;
}

export interface CollectorEnvelope {
    site_id: number;
    target: string;
    protocol: string;
    status: string;
    observed_at: string;
    duration_ms: number;
    error_code?: string;
    error_message?: string;
    payload: unknown;
    payload_bytes?: number;
    payload_truncated?: boolean;
    payload_preview_max_bytes?: number;
    payload_preview_available?: boolean;
    schema_version: number;
    collector_id?: string;
    job_id?: string;
}

export interface TargetLatestResponse {
    state: HealthSummaryState;
    site_id: number;
    target: string;
    protocols: Record<string, CollectorEnvelope>;
    reason_codes?: string[];
    reason_messages?: string[];
    generated_at?: string;
    schema_version?: number;
}

export interface TargetObservationsResponse {
    state: HealthSummaryState;
    site_id: number;
    target: string;
    protocol: string;
    limit: number;
    items: CollectorEnvelope[];
    reason_codes?: string[];
    reason_messages?: string[];
    generated_at?: string;
    schema_version?: number;
}

export interface TargetTrendResponse {
    state: HealthSummaryState;
    site_id: number;
    target: string;
    windows: unknown;
    reason_codes?: string[];
    reason_messages?: string[];
    generated_at: string;
    schema_version: number;
}

export interface TargetChangesResponse {
    state: HealthSummaryState;
    site_id: number;
    target: string;
    events: unknown;
    reason_codes?: string[];
    reason_messages?: string[];
    generated_at: string;
    schema_version: number;
}

export interface SiteV2DetailResponse {
    site: SiteV2Info;
    targets: SiteV2Target[];
    selected_target: string;
    site_summary: SiteHealthSummary;
    target_summary: TargetHealthSummary;
    latest_core: TargetLatestResponse;
    derived: {
        trend: TargetTrendResponse;
        changes: TargetChangesResponse;
    };
    light_probe_state: TargetLatestResponse;
    generated_at: string;
    schema_version: number;
}

export interface HttpRecord {
    domain: string;
    url: string;
    statusCode: number;
    responseTime: string;
    contentLength: number;
    title: string;
    server: string;
    redirects: string[];
    headers: Record<string, string[]>;
    meta: {
        charset: string;
        description: string;
        keywords: string;
    };
    tlsVersion: string;
    cipherSuite: string;
    certExpiry: string;
    certDaysLeft: string;
    certIssuer: string;
    certIssuerOrg: string[];
    certDNSNames: string[];
    certPubKeyAlg: string;
    certSigAlg: string;
    certEmail: string | null;
    certIsCA: boolean;
}

export interface DnsItem {
    type: "A" | "AAAA" | "CNAME" | "TXT" | "MX" | "NS" | "SOA" | "RRSIG";
    value: string;
    ttl: number;
    dnssec: boolean;
    asn: string;
    country: string;
    city: string;
    provider_type: string;
    isp: string;
    duration: number;
    children: DnsItem[] | null;
    reverse_ptr: string;
    hijacked: boolean;
}

export interface DnsRecord {
    a: string;
    AAAA: string;
    CNAME: string;
    txt: string;
    MX: string;
    ns: string;
    SOA: string;
    caa: string;
}

// 延迟时序结构
export interface PingRecord {
    twenty: PingStats;
    sixty: PingStats;
    hundred: PingStats;
}

export interface PingStats {
    DelayModel: Delay[];
    avgDelay: string;
    avgLoss: string;
}

export interface SayingModel {
    author: string;
    content: string;
}

export interface NavHomeBackgrounds {
    desktop: string;
    mobile: string;
}

export interface NavHomeResponse {
    schema_version: number;
    generated_at: string;
    cache_state: Record<string, string>;
    reason_messages?: Record<string, string>;
    sites: Site[];
    groups: Group[];
    ping: Record<string, string>;
    saying: SayingModel | null;
    backgrounds: NavHomeBackgrounds;
}

export interface NavHomePingResponse {
    schema_version: number;
    generated_at: string;
    state: string;
    reason_messages?: string[];
    ping: Record<string, string>;
}

export interface NavHomeSayingResponse {
    schema_version: number;
    generated_at: string;
    state: string;
    reason_messages?: string[];
    saying: SayingModel | null;
}

export interface NavSiteIndexItem {
    id: number;
    domains: string[];
    updated_at: string | null;
}

export interface NavSiteIndexResponse {
    schema_version: number;
    generated_at: string;
    state: string;
    reason_messages?: string[];
    items: NavSiteIndexItem[];
}

export type NavUpdatesState = 'ready' | 'empty' | 'error';

export interface NavUpdateNotice {
    id: number;
    title: string;
    body: string;
    published_at: string;
    create_time: string;
    update_time: string;
}

export interface NavUpdatesResponse {
    schema_version: number;
    generated_at: string;
    state: NavUpdatesState;
    reason_messages?: string[];
    items: NavUpdateNotice[];
}

export type NavSearchSuggestionEngine = 'baidu' | 'bing' | 'google' | 'bilibili' | 'duckduckgo';
export type NavSearchSuggestionsState = 'ready' | 'empty' | 'error';

export interface NavSearchSuggestionsResponse {
    schema_version: number;
    generated_at: string;
    state: NavSearchSuggestionsState;
    engine: NavSearchSuggestionEngine | '';
    query: string;
    suggestions: string[];
    cache_state: 'hit' | 'miss';
    reason_messages?: string[];
}
