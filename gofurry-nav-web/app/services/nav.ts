import type {
  DnsRecord,
  Group,
  HttpRecord,
  PingRecord,
  SayingModel,
  Site,
  SiteInfo,
  changelogResp
} from '~/types/nav'

export function getSites(lang: string): Promise<Site[]> {
  return useApi('nav')('/nav/page/site/list', { query: { lang } })
}

export function getGroups(lang: string): Promise<Group[]> {
  return useApi('nav')('/nav/page/group/list', { query: { lang } })
}

export function getPing(): Promise<Record<string, string>> {
  return useApi('nav')('/nav/page/ping/list')
}

export function addCount(): Promise<unknown> {
  return useApi('nav')('/nav/stat/add/count')
}

export function getSearchSuggestion(engine: 'baidu' | 'bing' | 'google' | 'bilibili', keyword: string): Promise<string[]> {
  return useApi('nav')(`/nav/page/search/${engine}`, { query: { q: keyword } })
}

export function getSaying(): Promise<SayingModel> {
  return useApi('nav')('/nav/page/header/getSaying')
}

export function getSiteDetail(id: string, lang: string): Promise<SiteInfo> {
  return useApi('nav')('/nav/site/getSiteDetail', { query: { id, lang } })
}

export function getSitePingRecord(domain: string): Promise<PingRecord> {
  return useApi('nav')('/nav/site/getSitePingRecord', { query: { domain } })
}

export function getSiteHttpRecord(domain: string): Promise<HttpRecord> {
  return useApi('nav')('/nav/site/getSiteHttpRecord', { query: { domain } })
}

export function getSiteDnsRecord(domain: string): Promise<DnsRecord> {
  return useApi('nav')('/nav/site/getSiteDnsRecord', { query: { domain } })
}

export function getImageUrl(type: string): Promise<string> {
  return useApi('nav')('/nav/page/header/image/url', { query: { type } })
}

export function getChangeLog(): Promise<changelogResp[]> {
  return useApi('nav')('/nav/site/changelog')
}
