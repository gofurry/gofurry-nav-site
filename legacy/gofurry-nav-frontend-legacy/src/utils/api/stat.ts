// /utils/api/stat.ts
import { createRequest } from '@/utils/request.ts'
import type {
    GroupCount,
    ViewsCount,
    RegionStat,
    CommonStat,
    SiteModel,
    PingModel,
    PromMetricsModel, PromMetricsHistoryModel
} from '@/types/stat.ts'

// 响应拦截器
const navRequest = createRequest(import.meta.env.VITE_NAV_API_BASE_URL)

// 获取分组统计
export function getGroupCount(lang :string): Promise<GroupCount[]> {
    return navRequest.get('/stat/chart/group/count', { params: { lang } })
}

// 获取访问统计
export function getViewsCount(): Promise<ViewsCount> {
    return navRequest.get('/stat/chart/views/count')
}

// 获取城市访问统计
export function getCityStat(): Promise<RegionStat> {
    return navRequest.get('/stat/chart/views/region/city')
}

// 获取国家访问统计
export function getCountryStat(): Promise<RegionStat> {
    return navRequest.get('/stat/chart/views/region/country')
}

// 获取省份访问统计
export function getProvinceStat(): Promise<RegionStat> {
    return navRequest.get('/stat/chart/views/region/province')
}

// 获取导航站点的基本数据
export function getSiteCommonStat(): Promise<CommonStat> {
    return navRequest.get('/stat/nav/site/common')
}

// 获取近日收录的站点
export function getLatestSiteList(lang :string): Promise<SiteModel[]> {
    return navRequest.get('/stat/nav/site/list', { params: { lang } })
}

// 获取最近的最高延迟的 ping 记录
export function getLatestPingList(): Promise<PingModel[]> {
    return navRequest.get('/stat/nav/site/ping/list')
}

// 服务器数据
export function getPromMetrics(): Promise<PromMetricsModel> {
    return navRequest.get('/stat/prom/metrics')
}

// 时序数据
export function getPromMetricsHistory(): Promise<PromMetricsHistoryModel> {
    return navRequest.get('/stat/prom/metrics/history')
}

// 背景图
export function getImageUrl(): Promise<string> {
    return navRequest.get('/stat/image/url')
}