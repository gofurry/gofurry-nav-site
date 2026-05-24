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
    observed_at: string;
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

export interface changelogResp {
    title: string;
    url: string;
    create_time: string;
    update_time: string;
}
