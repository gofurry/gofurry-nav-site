// types/stat.ts

// 分组统计
export interface GroupCount {
    group_id: string
    count: number
}

// 访问统计
export interface ViewsCount {
    total: number
    year_count: number
    month_count: number
    date: string[]
    count: number[]
}

// 地区统计
export interface RegionStat {
    region_map: Record<string, number>
}

export interface CommonStat {
    site_count: number
    domain_count: number,
    dns_count: number,
    http_count: number,
    ping_count: number,
    site_reach_rate: number,
    non_profit_business_ratio: number,
    sfw_nsfw_ratio: number
}

export interface SiteModel {
    name: string,
    country: string,
    create_time: string
}

export interface PingModel {
    name: string,
    status: string,
    createTime: string,
    loss: string,
    delay: string
}

// 节点监控数据接口
export interface NodeMetrics {
    cpu_usage: string; // CPU 使用率
    disk_usage: string; // 磁盘使用率
    mem_usage: string; // 内存使用量
    net_rx_1d: string; // 1天内网络接收量
    net_tx_1d: string; // 1天内网络发送量
    tcp_connections: string; // 连接数量
    uptime: string; // 服务器运行时长
}

// 汇总监控数据接口
export interface SummaryMetrics {
    avg_response_1h: string; // 1小时内平均响应时间
    fail_rate_1h: string; // 1小时内请求失败率
    http_requests_1d: string; // 1天内HTTP请求总量
    http_requests_7d: string; // 7天内HTTP请求总量
    p95_response_1h: string; // 1小时内P95响应时间
    p99_response_1h: string; // 1小时内P99响应时间
}

export interface PathAvgResponse {
    [path: string]: string;
}

export interface NavPathMetrics {
    avg_response_1h: PathAvgResponse;
}


export interface GamePathMetrics {
    avg_response_1h: PathAvgResponse;
}

export interface PromMetricsModel {
    node: NodeMetrics; // 节点服务器监控数据
    nav: SummaryMetrics; // nav 模块汇总监控数据
    game: SummaryMetrics; // game 模块汇总监控数据
    nav_path: NavPathMetrics; // nav 模块各接口路径监控数据
    game_path: GamePathMetrics; // game 模块各接口路径监控数据
}

export interface PromMetricsHistoryModel {
    cpu: HistoryMetricsModel
    connect: HistoryMetricsModel
    memory: HistoryMetricsModel
}

export interface HistoryMetricsModel {
    twenty_minutes: MetricsModel[];
    one_hour: MetricsModel[];
    twenty_hours: MetricsModel[];
}

export interface MetricsModel {
    time: number;
    usage: number;
}