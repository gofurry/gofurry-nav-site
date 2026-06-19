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