// /types/game.ts

export interface GameGroupRecord {
    latest: BaseGameInfoRecord[]
    recent: BaseGameInfoRecord[]
    hot: BaseGameInfoRecord[]
    free: BaseGameInfoRecord[]
}

export interface BaseGameInfoRecord {
    game_id: string
    avg_score: number
    comment_count: number
    name: string
    name_en: string
    info: string
    info_en: string
    header: string
}

export interface GamePanelRecord {
    top_count: TopCountVo
    top_discount_vo: PriceRecord[]
    top_price_vo: PriceRecord[]
    bottom_price: BottomPriceVo
}

export interface TopCountVo {
    one: TopPlayerCountRecord[]
    two: TopPlayerCountRecord[]
    three: TopPlayerCountRecord[]
    four: TopPlayerCountRecord[]
}

export interface TopPlayerCountRecord {
    id: string
    name: string
    count_peak: number
    count_recent: number
    collect_time: number
    header: string
}

export interface BottomPriceVo {
    one: PriceRecord[]
    two: PriceRecord[]
    three: PriceRecord[]
    four: PriceRecord[]
}

export interface PriceRecord {
    id: string
    name: string
    global_price: number
    china_price: number
    discount: number
    header: string
}

export interface LatestNewsRecord {
    "news_zh": NewsBaseModel[]
    "news_en": NewsBaseModel[]
}

export interface NewsBaseModel {
    id: string
    name: string
    post_time: string
    headline: string
    header: string
    author: string
    content: string
    url: string
}

export interface SearchItemModel {
    id: string
    name: string
    info: string
    cover: string
}

export interface AnonymousReviewModel {
    region: string
    score: number
    content: string
    ip: string
    time: string
    game_name: string
    game_cover: string
}

export interface GameTagRecord {
    id: string
    name: string
    prefix: string
    game_count: number
}

// 查询请求结构
export interface SearchPageQueryRequest {
    pageNum: number
    pageSize: number
    content?: string
    pub_start_time?: string // 格式: "2025-12-29 22:56:00"
    pub_end_time?: string
    update_start_time?: string
    update_end_time?: string
    score?: boolean
    remark_order?: boolean
    time_order?: boolean
    tag_list?: number[]
}

// 分页响应类型
export interface SearchPageResponseItem {
    id: string
    name: string
    info: string
    cover: string
    update_time: string
    release_date: string
    remark_count: number
    avg_score: number
    appid: number
    primary_tag: string
    secondary_tag: string
}

export interface SearchPageResponse {
    total: number
    list: SearchPageResponseItem[]
}

export interface GameBaseInfoResponse {
    name: string
    info: string
    create_time: string
    update_time: string
    resources: KvModel[]
    groups: KvModel[]
    links: KvModel[]
    release_date: string
    developers: string[]
    publishers: string[]
    appid: number
    cover: string
    platform: string
    price_list: PriceModel[]
    news: NewsModel[]
    tags: TagModel[]
    supported_languages: string
    required_age: string
    website: string
    detailed_description: string
    about_the_game: string
    support: SupportModel
    screenshots: ScreenshotsModel[]
    movies: MoviesModel[]
    pc_requirements: RequirementsModel
    online_count: number
    count_collect_time: string
}

export interface RequirementsModel {
    id: number
    minimum: string
    recommended: string
}

export interface MoviesModel {
    id: number
    name: string
    thumbnail: string
    dash_av1: string
    dash_h264: string
    hls_h264: string
}

export interface ScreenshotsModel {
    id: number
    path_thumbnail: string
    path_full: string
}

export interface SupportModel {
    url: string
    email: string
}

export interface KvModel {
    key: string
    value: string
}

export interface PriceModel {
    price: string
    country: string
}

export interface NewsModel {
    headline: string
    content: string
    post_time: string
    author: string
    url: string
}

export interface TagModel {
    id: string
    name: string
    desc: string
}

export interface RemarkResponse {
    total: number
    avg_score: number
    remarks: RemarkModel[]
}

export interface RemarkModel {
    region: string
    content: string
    score: number
    create_time: string
    ip: string
    name: string
}

export interface RecommendedModel {
    id: string
    name: string
    info: string
    similarity: number
    appid: string
}

export interface CommentReq {
    id: string
    name: string
    content: string
    score: number
}

export interface CreatorResponse {
    id: string
    name: string
    info: string
    url: string
    avatar: string
    links: KvModel[]
    contact: KvModel[]
    type: number
    create_time: string
    update_time: string
}

// 抽奖

export interface LotteryResp {
    history: LotteryHistoryModel
    active: LotteryActiveModel[]
}

export interface LotteryHistoryModel {
    prize: HistoryPrizeModel[]
    prize_count: number
}

export interface HistoryPrizeModel {
    name: string
    desc: string
    end_time: string
    prize: PrizeModel
    winner: MemberModel[]
    count: number
}

export interface PrizeModel {
    title: string
    platform: string
    count: number
}

export interface MemberModel {
    name: string
    email: string
}

export interface LotteryActiveModel {
    lottery: LotteryActiveVo
    member: MemberModel[]
    count: number
}

export interface LotteryActiveVo {
    id: string
    title: string
    desc: string
    start_time: string
    end_time: string
    prize: PrizeModel
}

export interface LotteryReq {
    id: number
    name: string
    email: string
    key: string
}