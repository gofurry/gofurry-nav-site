import type {
    AnonymousReviewModel, CommentReq,
    GameTagRecord,
    LotteryReq, LotteryResp, RecommendedModel,
    SearchItemModel, SearchPageQueryRequest, SearchPageResponse,
} from '@/types/game'
import type { ApiResult } from '@/types/common'
import type { NitroFetchOptions } from 'nitropack'

type ApiRequestOptions = NitroFetchOptions<string>

export function getRandomGame(): Promise<string> {
    return useApi('gameV2')('/game/recommend/random')
}

export function getSearchSimple(
    lang: string,
    txt: string,
    options: ApiRequestOptions = {}
): Promise<SearchItemModel[]> {
    return useApi('gameV2')('/game/search/simple', { method: 'POST', body: { txt, lang }, ...options })
}

export function getLatestReview(limit = 15): Promise<AnonymousReviewModel[]> {
    return useApi('gameV2')('/game/reviews/latest', { query: { limit } })
}

export function getTagList(lang: string, options: ApiRequestOptions = {}): Promise<GameTagRecord[]> {
    return useApi('gameV2')('/game/tags', { query: { lang }, ...options })
}

export function searchGameAdvanced(
    query: SearchPageQueryRequest,
    lang: string,
    options: ApiRequestOptions = {}
): Promise<SearchPageResponse> {
    return useApi('gameV2')('/game/search/page', { method: 'POST', body: { ...query, lang }, ...options })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
    return useApi('gameV2')('/game/recommend/similar', {
        query: { id, lang, region: 'CN', limit: 8 }
    })
}

// export function commitComment(query: CommentReq): Promise<ApiResult<string>> {
//     return gameRequest.post("/review/anonymous", {...query})
// }

export function commitComment(
    query: CommentReq
): Promise<ApiResult<string>> {
    return $fetch('/game/reviews/anonymous', {
        baseURL: useRuntimeConfig().public.gameV2ApiBase,
        credentials: 'include',
        method: 'POST',
        body: query
    })
}

export function getLottery(): Promise<LotteryResp> {
    return useApi('gameV2')('/game/prizes')
}

export function getLotteryParticipation(
    query: LotteryReq
): Promise<ApiResult<string>> {
    return $fetch('/game/prizes/participation', {
        baseURL: useRuntimeConfig().public.gameV2ApiBase,
        credentials: 'include',
        method: 'POST',
        body: query
    })
}
