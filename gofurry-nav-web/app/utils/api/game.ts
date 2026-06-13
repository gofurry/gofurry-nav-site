import type {
    AnonymousReviewModel, CommentReq, CreatorResponse,
    GameTagRecord,
    LotteryReq, LotteryResp, RecommendedModel,
    SearchItemModel, SearchPageQueryRequest, SearchPageResponse,
} from '@/types/game'
import type { ApiResult } from '@/types/common'

export function getRandomGame(): Promise<string> {
    return useApi('gameV2')('/game/recommend/random')
}

export function getSearchSimple(lang: string, txt: string): Promise<SearchItemModel[]> {
    return useApi('gameV2')('/game/search/simple', { method: 'POST', body: { txt, lang } })
}

export function getLatestReview(limit = 15): Promise<AnonymousReviewModel[]> {
    return useApi('gameV2')('/game/reviews/latest', { query: { limit } })
}

export function getTagList(lang: string): Promise<GameTagRecord[]> {
    return useApi('gameV2')('/game/tags', { query: { lang } })
}

export function searchGameAdvanced(query: SearchPageQueryRequest, lang: string): Promise<SearchPageResponse> {
    return useApi('gameV2')('/game/search/page', { method: 'POST', body: { ...query, lang } })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
    return useApi('gameV2')('/game/recommend/similar', {
        query: { id, lang, region: 'CN', limit: 8 }
    })
}

export function getGameCreator(lang: string): Promise<CreatorResponse[]> {
    return useApi('gameV2')('/game/creators', { query: { lang } })
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
