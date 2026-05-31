import type {
  Group,
  SayingModel,
  Site,
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

export function getImageUrl(type: string): Promise<string> {
  return useApi('nav')('/nav/page/header/image/url', { query: { type } })
}

export function getChangeLog(): Promise<changelogResp[]> {
  return useApi('nav')('/nav/site/changelog')
}
