import { createRequest } from '@/utils/request'
import type {
    AnonymousReviewModel, CommentReq, CreatorResponse,
    GameTagRecord,
    LotteryReq, LotteryResp, RecommendedModel,
    SearchItemModel, SearchPageQueryRequest, SearchPageResponse,
} from '@/types/game'
import type { ApiResult } from '@/types/common'
import axios from "axios";

const gameRequest = createRequest(import.meta.env.VITE_GAME_API_BASE_URL)

export function getRandomGame(): Promise<string> {
    return gameRequest.get("/game/recommend/random")
}

export function getSearchSimple(lang: string, txt: string): Promise<SearchItemModel[]> {
    return useApi('gameV2')('/game/search/simple', { method: 'POST', body: { txt, lang } })
}

export function getLatestReview(): Promise<AnonymousReviewModel[]> {
    return gameRequest.get("/game/review/latest")
}

export function getTagList(lang: string): Promise<GameTagRecord[]> {
    return useApi('gameV2')('/game/tags', { query: { lang } })
}

export function searchGameAdvanced(query: SearchPageQueryRequest, lang: string): Promise<SearchPageResponse> {
    return useApi('gameV2')('/game/search/page', { method: 'POST', body: { ...query, lang } })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
    return gameRequest.get("/game/recommend/CBF", { params: { id: id, lang: lang } })
}

export function getGameCreator(lang: string): Promise<CreatorResponse[]> {
    return gameRequest.get("/game/creator", { params: { lang: lang } })
}

// export function commitComment(query: CommentReq): Promise<ApiResult<string>> {
//     return gameRequest.post("/review/anonymous", {...query})
// }

export function commitComment(
    query: CommentReq
): Promise<ApiResult<string>> {
    return axios
        .post(
            `${import.meta.env.VITE_GAME_API_BASE_URL}/game/review/anonymous`,
            query
        )
        .then(res => res.data)
}

export function getLottery(): Promise<LotteryResp> {
    return gameRequest.get("/game/prize/info")
}

export function getLotteryParticipation(
    query: LotteryReq
): Promise<ApiResult<string>> {
    return axios
        .post(
            `${import.meta.env.VITE_GAME_API_BASE_URL}/game/prize/participation`,
            query
        )
        .then(res => res.data)
}
