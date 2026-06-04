import type {
  NavHomePingResponse,
  NavHomeResponse,
  NavSearchSuggestionEngine,
  NavSearchSuggestionsResponse,
  NavUpdatesResponse,
  SayingModel,
} from '~/types/nav'

export function getNavHome(lang: string): Promise<NavHomeResponse> {
  return useApi('navV2')('/nav/home', { query: { lang } })
}

export function getNavHomePing(): Promise<NavHomePingResponse> {
  return useApi('navV2')('/nav/home/ping')
}

export function getSearchSuggestion(
  engine: NavSearchSuggestionEngine,
  keyword: string,
  signal?: AbortSignal
): Promise<NavSearchSuggestionsResponse> {
  return useApi('navV2')('/nav/search/suggestions', { query: { engine, q: keyword }, signal })
}

export function getSaying(): Promise<SayingModel> {
  return useApi('nav')('/nav/page/header/getSaying')
}

export function getImageUrl(type: string): Promise<string> {
  return useApi('nav')('/nav/page/header/image/url', { query: { type } })
}

export function getNavUpdates(lang: 'zh' | 'en'): Promise<NavUpdatesResponse> {
  return useApi('navV2')('/nav/updates', { query: { lang } })
}
