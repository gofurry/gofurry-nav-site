import type {
  NavHomeBackgrounds,
  NavHomePingResponse,
  NavHomeResponse,
  NavHomeSayingResponse,
  NavSearchSuggestionEngine,
  NavSearchSuggestionsResponse,
  NavSiteIndexResponse,
  NavUpdatesResponse,
  SayingModel,
} from '~/types/nav'

export function getNavHome(lang: string): Promise<NavHomeResponse> {
  return useApi('navV2')('/nav/home', { query: { lang } })
}

export function getNavHomePing(): Promise<NavHomePingResponse> {
  return useApi('navV2')('/nav/home/ping')
}

export function getNavHomeSaying(): Promise<NavHomeSayingResponse> {
  return useApi('navV2')('/nav/home/saying')
}

export function getNavHomeBackgrounds(): Promise<NavHomeResponse['backgrounds']> {
  return useApi('navV2')<{ backgrounds: NavHomeBackgrounds }>('/nav/home/backgrounds').then((response) => response.backgrounds)
}

export function getNavSiteIndex(): Promise<NavSiteIndexResponse> {
  return useApi('navV2')('/nav/sites/index')
}

export function getSearchSuggestion(
  engine: NavSearchSuggestionEngine,
  keyword: string,
  signal?: AbortSignal
): Promise<NavSearchSuggestionsResponse> {
  return useApi('navV2')('/nav/search/suggestions', { query: { engine, q: keyword }, signal })
}

export function getNavUpdates(lang: 'zh' | 'en'): Promise<NavUpdatesResponse> {
  return useApi('navV2')('/nav/updates', { query: { lang } })
}
