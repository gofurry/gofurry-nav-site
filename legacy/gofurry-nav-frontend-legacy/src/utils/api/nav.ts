import { createRequest } from '@/utils/request.ts'
import type {Site, Group, SiteInfo, HttpRecord, DnsRecord, PingRecord, SayingModel, changelogResp} from '@/types/nav.ts'

// 响应拦截器
const navRequest = createRequest(import.meta.env.VITE_NAV_API_BASE_URL)

// 获取站点列表
export function getSites(lang: string): Promise<Site[]> {
    return navRequest.get(`/nav/page/site/list`, { params: { lang } })
}

// 获取分组
export function getGroups(lang: string): Promise<Group[]> {
    return navRequest.get(`/nav/page/group/list`, { params: { lang } })
}

// 获取延迟信息
export function getPing(): Promise<Record<string, string>> {
    return navRequest.get(`/nav/page/ping/list`)
}

// 增加浏览量
export function addCount() {
    navRequest.get(`/stat/add/count`)
}

// 搜索建议
export function getBaiduSuggestion(keyword: string): Promise<string[]> {
    return navRequest.get('/nav/page/search/baidu', { params: { q: keyword } })
}

export function getBingSuggestion(keyword: string): Promise<string[]> {
    return navRequest.get('/nav/page/search/bing', { params: { q: keyword } })
}

export function getGoogleSuggestion(keyword: string): Promise<string[]> {
    return navRequest.get('/nav/page/search/google', { params: { q: keyword } })
}

export function getBiliBiliSuggestion(keyword: string): Promise<string[]> {
    return navRequest.get('/nav/page/search/bilibili', { params: { q: keyword } })
}

// 随机金句
export function getSaying(): Promise<SayingModel> {
    return navRequest.get('/nav/page/header/getSaying')
}

// 详情页
export function getSiteDetail(id: string, lang: string): Promise<SiteInfo> {
    return navRequest.get('/nav/site/getSiteDetail', { params: { id, lang } })
}

export function getSitePingRecord(domain: string): Promise<PingRecord> {
    return navRequest.get('/nav/site/getSitePingRecord', { params: { domain } })
}

export function getSiteHttpRecord(domain: string): Promise<HttpRecord> {
    return navRequest.get('/nav/site/getSiteHttpRecord', { params: { domain } })
}

export function getSiteDnsRecord(domain: string): Promise<DnsRecord> {
    return navRequest.get('/nav/site/getSiteDnsRecord', { params: { domain } })
}

export function getImageUrl(type: string): Promise<string> {
    return navRequest.get('/nav/page/header/image/url', { params: { type } })
}

export function getChangeLog(): Promise<changelogResp[]> {
    return navRequest.get('/site/changelog')
}