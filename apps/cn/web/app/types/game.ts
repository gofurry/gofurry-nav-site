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
    release: GameV2ReleaseState | null
    first_available: GameV2FirstAvailable | null
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
    code: string
    category_id: string
    category_code: string
    category_name: string
    game_count: number
}

// 查询请求结构
export interface SearchPageQueryRequest {
    pageNum: number
    pageSize: number
    content?: string
    availability?: GameSearchAvailability
    pub_start_time?: string // 格式: "2025-12-29 22:56:00"
    pub_end_time?: string
    planned_start_time?: string
    planned_end_time?: string
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
    release: GameV2ReleaseState | null
    first_available: GameV2FirstAvailable | null
}

export type GameSearchAvailability = 'available' | 'upcoming'

export interface SearchPageResponse {
    total: number
    list: SearchPageResponseItem[]
}

export interface GameBaseInfoResponse {
    name: string
    info: string
    type: string
    is_free: boolean
    short_description: string
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
    release: GameV2Release
    first_available: GameV2FirstAvailable | null
    languages: GameV2Language[]
    required_age: string
    website: string
    detailed_description: string
    about_the_game: string
    support: SupportModel
    support_info: Record<string, string>
    screenshots: ScreenshotsModel[]
    movies: MoviesModel[]
    requirements: GameRequirementsGroupModel
    pc_requirements: RequirementsModel
    content_descriptors: unknown
    ratings: unknown
    media: GameV2MediaView
    online_count: number
    count_collect_time: string
    view_count: number
}

export interface GameRequirementsGroupModel {
    pc: RequirementsModel
    mac: RequirementsModel
    linux: RequirementsModel
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
    code: string
    category_code: string
    role: 'normal' | 'primary' | 'secondary'
    id: string
    name: string
    desc: string
}

export interface RemarkResponse {
    total: number
    avg_score: number
    page_num: number
    page_size: number
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
    library_cover_url: string
    library_cover_2x_url: string
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

// Business text is localized by Game Backend's lang query, with per-field zh fallback.
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
    code: string
    category_code: string
    role: 'normal' | 'primary' | 'secondary'
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
    peak_count?: number
    status: string
    collected_at: string
}

export interface GameV2ListItem {
    id: string
    appid: string
    name: string
    name_zh?: string
    name_en?: string
    summary: string
    summary_zh?: string
    summary_en?: string
    header_url: string
    capsule_url: string
    release_date: string
    release: GameV2ReleaseState | null
    first_available: GameV2FirstAvailable | null
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
    popular_games?: GameV2ListItem[]
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

// Independent optional Home slice. IDs and signed tokens stay opaque strings.
export interface GameShowcaseAction {
    type: 'game' | 'project' | 'product' | 'website'
    game_id?: string
    target?: string
}

export interface GameShowcaseSecondaryAction {
    type: 'steam' | 'kickstarter' | 'website' | 'other'
    target?: string
}

export interface GameShowcaseArtwork {
    kind: 'managed' | 'steam'
    url?: string
    desktop_object_key?: string
    mobile_object_key?: string
    focal_x?: number
    focal_y?: number
}

export type GameShowcaseRelease = Pick<GameV2Release, 'availability' | 'precision' | 'exact_date'
    | 'year' | 'month' | 'quarter' | 'window_start' | 'window_end'>

export interface GameShowcaseItem {
    key: string
    source: 'managed' | 'automatic'
    reason: 'editorial' | 'sponsored' | 'upcoming' | 'new_release' | 'trending'
    content_type: 'game' | 'crowdfunding' | 'tabletop' | 'merchandise' | 'other'
    sponsored: boolean
    campaign_id?: string
    game_id?: string
    title: string
    summary: string
    tags: string[]
    editorial_note?: string
    artwork: GameShowcaseArtwork
    primary_action: GameShowcaseAction
    secondary_action?: GameShowcaseSecondaryAction
    release?: GameShowcaseRelease
    position: number
    tracking_token: string
}

export interface GameShowcaseSnapshot {
    schema_version: 1
    snapshot_id: string
    generated_at: string
    valid_until: string
    items: GameShowcaseItem[]
}

export type GameShowcaseClickSource = 'artwork' | 'title' | 'primary' | 'secondary'
export type GameShowcaseEvent = {
    tracking_token: string
    session_id: string
} & ({ event: 'impression'; source?: never } | { event: 'click'; source: GameShowcaseClickSource })

export interface GameViewTouchResponse {
    game_id: number
    view_count: number
}

export interface GameV2Release {
    coming_soon: boolean
    date: string
    availability: 'upcoming' | 'available' | 'unknown'
    precision: GameReleasePrecision
    exact_date: string | null
    year: number | null
    month: number | null
    quarter: number | null
    window_start: string | null
    window_end: string | null
    raw_text: string
    observed_at: string | null
}

export type GameV2ReleaseState = Omit<GameV2Release, 'coming_soon' | 'date'>

export type GameReleasePrecision = 'day' | 'month' | 'quarter' | 'year' | 'tba' | 'none' | 'unknown'

export interface GameV2FirstAvailable {
    precision: Exclude<GameReleasePrecision, 'tba' | 'none' | 'unknown'>
    exact_date: string | null
    year: number
    month: number | null
    quarter: number | null
    window_start: string
    window_end: string
    source: 'legacy_manual' | 'observed_transition' | 'steam_backfill'
    inferred: boolean
}

export interface GameV2Language {
    code: string | null
    steam_name: string
    steam_api_code: string | null
    steam_web_code: string | null
    tier: 'platform' | 'game_only' | 'unknown'
    interface_supported: boolean | null
    subtitles_supported: boolean | null
    full_audio_supported: boolean | null
}

export interface GameV2MediaView {
    header_url: string
    capsule_url: string
    capsule_v5_url: string
    capsule_small_url: string
    capsule_main_url: string
    library_cover_url: string
    library_cover_2x_url: string
    library_hero_url: string
    library_logo_url: string
    library_logo_2x_url: string
    background_url: string
    background_raw_url: string
    screenshots: GameV2Screenshot[]
    movies: GameV2Movie[]
    assets: GameV2AssetView[]
}

export interface GameV2AssetView {
    type: string
    family: string
    source: string
    lang: string
    key: string
    title: string
    url: string
    thumbnail_url: string
    format: string
    exists?: boolean
    status_code: number
    content_type: string
    content_length: number
    extra?: Record<string, unknown> | unknown
    sort_order: number
    checked_at?: string
    collected_at: string
    updated_at: string
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

export interface GameV2DetailExtraPayload {
    content_descriptors?: unknown
    ratings?: unknown
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
    first_available: GameV2FirstAvailable | null
    languages: GameV2Language[]
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
    extra?: GameV2DetailExtraPayload
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

export interface GameTagCategory {
    id: string
    code: string
    name: string
    tags: GameTagRecord[]
}
// Collection schema v1: order, visibility and chronology are Backend-owned.
export type GameCollectionMode = 'sfw' | 'nsfw'
export type GameCollectionPhase = 'released' | 'released_unknown' | 'upcoming_overdue' | 'upcoming' | 'upcoming_tba' | 'unknown'
export interface GameCollectionMetadata { schema_version: 1; generated_at: string; as_of_date: string }
export interface GameCollectionInfo { code: string; name: string; info: string; visible_game_count: number; published_at: string }
export interface GameCollectionPreviewGame { game_id: string; name: string; header_url: string }
export interface GameCollectionSummary extends GameCollectionInfo { preview_games: GameCollectionPreviewGame[] }
export interface GameCollectionHome extends GameCollectionMetadata { slots: { slot: number; collection: GameCollectionSummary }[] }
export interface GameCollectionIndex extends GameCollectionMetadata { page: number; page_size: number; total: number; has_more: boolean; items: GameCollectionSummary[] }
export interface GameCollectionChronology {
  source: 'first_available' | 'release'
  precision: 'day' | 'month' | 'quarter' | 'year'
  window_start: string
  window_end: string
  inferred: boolean
}
export interface GameCollectionTimelineItem {
  game_id: string; name: string; summary: string; header_url: string
  phase: GameCollectionPhase; chronology: GameCollectionChronology | null
  // Additive schema v1 fields may be absent in pre-upgrade cached responses.
  primary_tag?: { code: string; name: string } | null
  secondary_tag?: { code: string; name: string } | null
  rating?: { average: number; count: number } | null
  online?: { count: number; collected_at: string } | null
  community_count?: number
}
export interface GameCollectionDetail extends GameCollectionMetadata { collection: GameCollectionInfo; items: GameCollectionTimelineItem[] }

export type GameCollectionPhaseFilter = 'all' | 'released' | 'upcoming' | 'mixed'
export type GameCollectionSort = 'published_desc' | 'count_desc' | 'count_asc' | 'name_asc' | 'name_desc'
export interface GameCollectionCriteria {
  q: string
  phase: GameCollectionPhaseFilter
  sort: GameCollectionSort
}
