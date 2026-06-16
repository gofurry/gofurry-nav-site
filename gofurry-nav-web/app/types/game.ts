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
    desc: string
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
    desc: string
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
    view_count: number
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
    mp4_url?: string
    webm_url?: string
}

export interface GameV2MovieExtra {
    dash_av1_url?: string
    dash_h264_url?: string
    hls_h264_url?: string
    mp4_480_url?: string
    mp4_max_url?: string
    webm_480_url?: string
    webm_max_url?: string
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
    appid: string
    name: string
    summary: string
    header_url: string
    capsule_url: string
    score: number
    display_score: number
    rank: number
    reasons: RecommendationReason[]
    algorithm_version: string
    computed_at: string
    tags: GameV2Tag[]
    price: GameV2PriceView
    online_count: GameV2OnlineCount
}

export interface RecommendationReason {
    type: string
    label: string
    value: string
    weight: number
}

export interface CommentReq {
    id: string
    name: string
    content: string
    score: number
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

export interface GameV2Tag {
    id: string
    name: string
    desc: string
}

export interface GameV2PriceView {
    region: string
    available: boolean
    unavailable_reason?: string
    is_free: boolean
    currency: string
    initial_amount: number
    final_amount: number
    discount_percent: number
    initial_formatted: string
    final_formatted: string
    collected_at: string
    updated_at: string
}

export interface GameV2OnlineCount {
    count: number
    status: string
    collected_at: string
}

export interface GameV2ListItem {
    id: string
    appid: string
    name: string
    summary: string
    header_url: string
    capsule_url: string
    release_date: string
    developers: string[]
    publishers: string[]
    platforms: Record<string, boolean>
    prices: GameV2PriceView[]
    price: GameV2PriceView
    online_count: GameV2OnlineCount
    tags: GameV2Tag[]
    avg_score: number
    comment_count: number
    updated_at: string
}

export interface GameV2PanelRecord {
    latest_games: GameV2ListItem[]
    updated_games: GameV2ListItem[]
    top_online: GameV2ListItem[]
    free_games: GameV2ListItem[]
    top_price: GameV2ListItem[]
    highest_discount: GameV2ListItem[]
    low_price: GameV2ListItem[]
    latest_news: GameV2NewsItem[]
}

export interface GameHomeApiNewsRecord {
    news_zh: GameV2NewsItem[]
    news_en: GameV2NewsItem[]
}

export interface GameHomeApiResponse {
    panel: GameV2PanelRecord
    latest_news: GameHomeApiNewsRecord
    latest_reviews: AnonymousReviewModel[]
}

export interface GameV2Release {
    coming_soon: boolean
    date: string
}

export interface GameV2MediaView {
    header_url: string
    capsule_url: string
    capsule_v5_url: string
    background_url: string
    background_raw_url: string
    screenshots: GameV2Screenshot[]
    movies: GameV2Movie[]
}

export interface GameV2Screenshot {
    id: string
    url: string
    thumbnail_url: string
}

export interface GameV2Movie {
    id: string
    name: string
    url: string
    thumbnail_url: string
    extra?: GameV2MovieExtra | Record<string, unknown>
}

export interface GameV2RequirementsView {
    pc: Record<string, string>
    mac: Record<string, string>
    linux: Record<string, string>
}

export interface GameV2SiteInfo {
    id: string
    name: string
    info: string
    header: string
    view_count: number
    resources: KvModel[]
    groups: KvModel[]
    links: KvModel[]
    create_time: string
    update_time: string
}

export interface GameV2DetailRecord {
    id: string
    appid: string
    requested_lang: string
    lang: string
    name: string
    summary: string
    type: string
    is_free: boolean
    website: string
    header_url: string
    short_description: string
    detailed_description: string
    about_the_game: string
    release: GameV2Release
    developers: string[]
    publishers: string[]
    platforms: Record<string, boolean>
    supported_languages: string
    support_info: Record<string, string>
    prices: GameV2PriceView[]
    price: GameV2PriceView
    media: GameV2MediaView
    requirements: GameV2RequirementsView
    news: GameV2NewsItem[]
    online_count: GameV2OnlineCount
    site: GameV2SiteInfo
    tags: GameV2Tag[]
    collected_at: string
    updated_at: string
}

export interface GameV2NewsItem {
    id: string
    game_id: string
    appid: string
    lang: string
    game_name: string
    header_url: string
    event_gid: string
    headline: string
    summary: string
    plain_text: string
    html: string
    url: string
    tags: string[]
    published_at: string
    updated_at: string
    comment_count: number
    vote_up_count: number
    vote_down_count: number
}
