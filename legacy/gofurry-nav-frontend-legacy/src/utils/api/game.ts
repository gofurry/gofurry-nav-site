import { createRequest } from '@/utils/request.ts'
import type {
    AnonymousReviewModel, CommentReq, CreatorResponse, GameBaseInfoResponse,
    GameGroupRecord,
    GamePanelRecord, GameTagRecord, LatestNewsRecord,
    LotteryReq, LotteryResp, NewsBaseModel, RecommendedModel, RemarkResponse,
    SearchItemModel, SearchPageQueryRequest, SearchPageResponse,
} from '@/types/game.ts'
import type { ApiResult } from '@/types/common.ts'
import axios from "axios";

const gameRequest = createRequest(import.meta.env.VITE_GAME_API_BASE_URL)

export function getGameList() {
    return gameRequest.get("/game/list")
}

export function getGameMainInfo(): Promise<GameGroupRecord> {
    return gameRequest.get(`/game/info/main`)
}

export function getGameMainPanel(): Promise<GamePanelRecord> {
    return gameRequest.get("/game/panel/main")
}

export function getLatestGameNews(): Promise<LatestNewsRecord> {
    return gameRequest.get("/game/update/latest")
}

export function getMoreLatestGameNews(lang :string): Promise<NewsBaseModel[]> {
    return gameRequest.get("/game/update/latest/more", { params: { lang: lang } })
}

export function getRandomGame(): Promise<string> {
    return gameRequest.get("/recommend/game/random")
}

export function getSearchSimple(lang: string, txt: string): Promise<SearchItemModel[]> {
    return gameRequest.post("/search/game/simple", { txt, lang });
}

export function getLatestReview(): Promise<AnonymousReviewModel[]> {
    return gameRequest.get("/review/latest")
}

export function getTagList(lang: string): Promise<GameTagRecord[]> {
    return gameRequest.get("/game/tag/list", { params: { lang: lang } })
}

export function searchGameAdvanced(query: SearchPageQueryRequest, lang: string): Promise<SearchPageResponse> {
    return gameRequest.post("/search/game/page", {...query,lang})
}

export function getGameBaseInfo(id: string, lang: string): Promise<GameBaseInfoResponse> {
    return gameRequest.get("/game/info", { params: { id: id, lang: lang } })
}

export function getGameRemark(id: string): Promise<RemarkResponse> {
    return gameRequest.get("/game/remark", { params: { id: id } })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
    return gameRequest.get("/recommend/game/CBF", { params: { id: id, lang: lang } })
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
            `${import.meta.env.VITE_GAME_API_BASE_URL}/review/anonymous`,
            query
        )
        .then(res => res.data)
}

export function getLottery(): Promise<LotteryResp> {
    return gameRequest.get("/prize/info")
}

export function getLotteryParticipation(
    query: LotteryReq
): Promise<ApiResult<string>> {
    return axios
        .post(
            `${import.meta.env.VITE_GAME_API_BASE_URL}/prize/participation`,
            query
        )
        .then(res => res.data)
}