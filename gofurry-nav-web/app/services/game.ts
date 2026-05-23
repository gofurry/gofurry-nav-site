import type {
  AnonymousReviewModel,
  CommentReq,
  CreatorResponse,
  GameBaseInfoResponse,
  GameGroupRecord,
  GamePanelRecord,
  GameTagRecord,
  LatestNewsRecord,
  LotteryReq,
  LotteryResp,
  NewsBaseModel,
  RecommendedModel,
  RemarkResponse,
  SearchItemModel,
  SearchPageQueryRequest,
  SearchPageResponse
} from '~/types/game'
import type { ApiResult } from '~/types/api'

export function getGameList() {
  return useApi('game')('/game/info/list')
}

export function getGameMainInfo(): Promise<GameGroupRecord> {
  return useApi('game')('/game/info/main')
}

export function getGameMainPanel(): Promise<GamePanelRecord> {
  return useApi('game')('/game/panel/main')
}

export function getLatestGameNews(): Promise<LatestNewsRecord> {
  return useApi('game')('/game/update/latest')
}

export function getMoreLatestGameNews(lang: string): Promise<NewsBaseModel[]> {
  return useApi('game')('/game/update/latest/more', { query: { lang } })
}

export function getRandomGame(): Promise<string> {
  return useApi('game')('/game/recommend/random')
}

export function getSearchSimple(lang: string, txt: string): Promise<SearchItemModel[]> {
  return useApi('game')('/game/search/simple', { method: 'POST', body: { txt, lang } })
}

export function getLatestReview(): Promise<AnonymousReviewModel[]> {
  return useApi('game')('/game/review/latest')
}

export function getTagList(lang: string): Promise<GameTagRecord[]> {
  return useApi('game')('/game/tag/list', { query: { lang } })
}

export function searchGameAdvanced(query: SearchPageQueryRequest, lang: string): Promise<SearchPageResponse> {
  return useApi('game')('/game/search/page', { method: 'POST', body: { ...query, lang } })
}

export function getGameBaseInfo(id: string, lang: string): Promise<GameBaseInfoResponse> {
  return useApi('game')('/game/info', { query: { id, lang } })
}

export function getGameRemark(id: string): Promise<RemarkResponse> {
  return useApi('game')('/game/remark', { query: { id } })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
  return useApi('game')('/game/recommend/CBF', { query: { id, lang } })
}

export function getGameCreator(lang: string): Promise<CreatorResponse[]> {
  return useApi('game')('/game/creator', { query: { lang } })
}

export function commitComment(query: CommentReq): Promise<ApiResult<string>> {
  return $fetch('/game/review/anonymous', {
    baseURL: useRuntimeConfig().public.gameApiBase,
    credentials: 'include',
    method: 'POST',
    body: query
  })
}

export function getLottery(): Promise<LotteryResp> {
  return useApi('game')('/game/prize/info')
}

export function getLotteryParticipation(query: LotteryReq): Promise<ApiResult<string>> {
  return $fetch('/game/prize/participation', {
    baseURL: useRuntimeConfig().public.gameApiBase,
    credentials: 'include',
    method: 'POST',
    body: query
  })
}
