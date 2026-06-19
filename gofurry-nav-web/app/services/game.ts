import type {
  AnonymousReviewModel,
  CommentReq,
  GameBaseInfoResponse,
  GameGroupRecord,
  GamePanelRecord,
  PriceRecord,
  GameV2DetailRecord,
  GameV2AssetView,
  GameV2ListItem,
  GameV2Movie,
  GameV2NewsItem,
  GameV2PanelRecord,
  GameTagRecord,
  GameHomeApiResponse,
  GameViewTouchResponse,
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

export interface GameHomeData {
  mainInfo: GameGroupRecord
  panelData: GamePanelRecord
  latestNews: LatestNewsRecord
  latestReviews: AnonymousReviewModel[]
}

export function getGameList() {
  return useApi('gameV2')<GameV2ListItem[]>('/game/list')
}

export async function getGameHomeData(lang = 'zh'): Promise<GameHomeData> {
  const payload = await useApi('gameV2')<GameHomeApiResponse>('/game/home', {
    query: {
      lang: normalizeGameLang(lang),
      region: 'CN',
    }
  })

  return {
    mainInfo: mapV2PanelToGameGroup(payload.panel),
    panelData: mapV2PanelToGamePanel(payload.panel),
    latestNews: {
      news_zh: payload.latest_news.news_zh.map(mapV2News),
      news_en: payload.latest_news.news_en.map(mapV2News),
    },
    latestReviews: payload.latest_reviews,
  }
}

export async function getGameMainInfo(lang = 'zh'): Promise<GameGroupRecord> {
  return mapV2PanelToGameGroup(await getGameV2Panel(lang))
}

export async function getGameMainPanel(lang = 'zh'): Promise<GamePanelRecord> {
  return mapV2PanelToGamePanel(await getGameV2Panel(lang))
}

export async function getLatestGameNews(): Promise<LatestNewsRecord> {
  const [newsZh, newsEn] = await Promise.all([
    getMoreLatestGameNews('zh', 12),
    getMoreLatestGameNews('en', 12),
  ])
  return {
    news_zh: newsZh,
    news_en: newsEn,
  }
}

export async function getMoreLatestGameNews(lang: string, limit = 100): Promise<NewsBaseModel[]> {
  const rows = await useApi('gameV2')<GameV2NewsItem[]>('/game/news/latest', {
    query: { lang: normalizeGameLang(lang), limit }
  })
  return rows.map(mapV2News)
}

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

export function getGameBaseInfo(id: string, lang: string): Promise<GameBaseInfoResponse> {
  return useApi('gameV2')<GameV2DetailRecord>('/game/info', {
    query: { id, lang: normalizeGameLang(lang), region: 'CN', news_limit: 20 }
  }).then(mapV2Detail)
}

export function getGameRemark(id: string, page = 1, limit = 5): Promise<RemarkResponse> {
  return useApi('gameV2')('/game/reviews', { query: { id, page, limit } })
}

export function getRecommendedGame(id: string, lang: string): Promise<RecommendedModel[]> {
  return useApi('gameV2')('/game/recommend/similar', {
    query: { id, lang: normalizeGameLang(lang), region: 'CN', limit: 8 }
  })
}

export function touchGameView(id: string): Promise<GameViewTouchResponse> {
  return useApi('gameV2')(`/game/games/${id}/view`, {
    method: 'POST'
  })
}

export function commitComment(query: CommentReq): Promise<ApiResult<string>> {
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

export function getLotteryParticipation(query: LotteryReq): Promise<ApiResult<string>> {
  return $fetch('/game/prizes/participation', {
    baseURL: useRuntimeConfig().public.gameV2ApiBase,
    credentials: 'include',
    method: 'POST',
    body: query
  })
}

function getGameV2Panel(lang: string): Promise<GameV2PanelRecord> {
  return useApi('gameV2')('/game/panel/main', {
    query: {
      lang: normalizeGameLang(lang),
      region: 'CN',
      limit: 24,
      top_online_limit: 60,
      price_limit: 120,
      news_limit: 12,
    }
  })
}

function normalizeGameLang(lang: string) {
  return lang === 'en' ? 'en' : 'zh'
}

function mapV2PanelToGameGroup(panel: GameV2PanelRecord): GameGroupRecord {
  return {
    latest: panel.latest_games.map(mapV2ListItemToBase),
    recent: panel.updated_games.map(mapV2ListItemToBase),
    hot: panel.top_online.slice(0, 24).map(mapV2ListItemToBase),
    free: panel.free_games.map(mapV2ListItemToBase),
  }
}

function mapV2PanelToGamePanel(panel: GameV2PanelRecord): GamePanelRecord {
  const topOnline = panel.top_online.map(mapV2ListItemToTopPlayer)
  const topPrice = panel.top_price.map(mapV2ListItemToPrice)
  const highestDiscount = panel.highest_discount.map(mapV2ListItemToPrice)
  const lowPrice = panel.low_price.map(mapV2ListItemToPrice)

  return {
    top_count: {
      one: topOnline.slice(0, 15),
      two: topOnline.slice(15, 30),
      three: topOnline.slice(30, 45),
      four: topOnline.slice(45, 60),
    },
    top_discount_vo: highestDiscount
      .sort((a, b) => b.discount - a.discount || a.global_price - b.global_price)
      .slice(0, 15),
    top_price_vo: topPrice
      .sort((a, b) => b.global_price - a.global_price || b.discount - a.discount)
      .slice(0, 15),
    bottom_price: {
      one: buildUsdPriceZone(lowPrice, 10),
      two: buildUsdPriceZone(lowPrice, 15),
      three: buildUsdPriceZone(lowPrice, 20),
      four: buildUsdPriceZone(lowPrice, 25),
    },
  }
}

function buildUsdPriceZone(items: PriceRecord[], maxDollars: number) {
  const maxCents = maxDollars * 100
  return items
    .filter((item) => item.global_price > 0 && item.global_price <= maxCents)
    .sort((a, b) => b.global_price - a.global_price || b.discount - a.discount)
    .slice(0, 15)
}

function mapV2ListItemToBase(game: GameV2ListItem) {
  return {
    game_id: game.id,
    avg_score: game.avg_score ?? 0,
    comment_count: game.comment_count ?? 0,
    name: game.name,
    name_en: game.name,
    info: game.summary,
    info_en: game.summary,
    header: bestV2Cover(game),
  }
}

function mapV2ListItemToTopPlayer(game: GameV2ListItem) {
  return {
    id: game.id,
    name: game.name,
    desc: game.summary,
    count_peak: game.online_count?.count ?? 0,
    count_recent: game.online_count?.count ?? 0,
    collect_time: toUnixSeconds(game.online_count?.collected_at),
    header: bestV2Cover(game),
  }
}

function mapV2ListItemToPrice(game: GameV2ListItem) {
  const price = game.price
  const globalPrice = selectGlobalPrice(game.prices ?? [], price)
  return {
    id: game.id,
    name: game.name,
    desc: game.summary,
    global_price: globalPrice?.available ? globalPrice.final_amount : 0,
    china_price: price?.available ? price.final_amount : 0,
    discount: globalPrice?.discount_percent ?? price?.discount_percent ?? 0,
    header: bestV2Cover(game),
  }
}

function selectGlobalPrice(prices: GameV2ListItem['prices'], regionPrice?: GameV2ListItem['price']) {
  return prices.find(price => price.available && price.region === 'US')
    || prices.find(price => price.available && price.currency === 'USD')
    || prices.find(price => price.available && price.region !== 'CN')
    || regionPrice
}

function bestV2Cover(game: GameV2ListItem) {
  return game.header_url || ''
}

function mapV2Detail(detail: GameV2DetailRecord): GameBaseInfoResponse {
  const requirements = {
    pc: {
      id: Number(detail.id || 0),
      minimum: detail.requirements?.pc?.minimum ?? '',
      recommended: detail.requirements?.pc?.recommended ?? '',
    },
    mac: {
      id: Number(detail.id || 0),
      minimum: detail.requirements?.mac?.minimum ?? '',
      recommended: detail.requirements?.mac?.recommended ?? '',
    },
    linux: {
      id: Number(detail.id || 0),
      minimum: detail.requirements?.linux?.minimum ?? '',
      recommended: detail.requirements?.linux?.recommended ?? '',
    },
  }

  return {
    name: detail.name,
    info: detail.summary || detail.short_description,
    type: detail.type ?? '',
    is_free: Boolean(detail.is_free),
    short_description: detail.short_description || detail.summary || '',
    create_time: detail.site?.create_time || detail.collected_at || '',
    update_time: detail.site?.update_time || detail.updated_at || '',
    resources: detail.site?.resources ?? [],
    groups: detail.site?.groups ?? [],
    links: detail.site?.links ?? [],
    release_date: detail.release?.date ?? '',
    developers: detail.developers ?? [],
    publishers: detail.publishers ?? [],
    appid: Number(detail.appid || 0),
    cover: bestV2DetailCover(detail),
    platform: platformsToText(detail.platforms),
    price_list: (detail.prices ?? []).map((price) => ({
      country: price.region,
      price: price.available ? (price.final_formatted || String(price.final_amount)) : '',
    })),
    news: (detail.news ?? []).map(mapV2DetailNews),
    tags: detail.tags ?? [],
    supported_languages: detail.supported_languages ?? '',
    required_age: '',
    website: detail.website ?? '',
    detailed_description: detail.detailed_description ?? '',
    about_the_game: detail.about_the_game ?? '',
    support: {
      url: detail.support_info?.url ?? '',
      email: detail.support_info?.email ?? '',
    },
    support_info: detail.support_info ?? {},
    screenshots: (detail.media?.screenshots ?? []).map((item) => ({
      id: Number(item.id || 0),
      path_thumbnail: item.thumbnail_url,
      path_full: item.url,
    })),
    movies: (detail.media?.movies ?? []).map(mapV2Movie),
    requirements,
    pc_requirements: requirements.pc,
    content_descriptors: detail.extra?.content_descriptors ?? null,
    ratings: detail.extra?.ratings ?? null,
    media: detail.media,
    online_count: detail.online_count?.count ?? 0,
    count_collect_time: detail.online_count?.collected_at ?? '',
    view_count: detail.site?.view_count ?? 0,
  }
}

function bestV2DetailCover(detail: GameV2DetailRecord) {
  return detail.media?.library_cover_2x_url
    || detail.media?.library_cover_url
    || firstStoreBrowseAssetURL(detail.media?.assets, [
    'library_capsule_2x',
    'library_capsule',
  ])
}

function firstStoreBrowseAssetURL(assets: GameV2AssetView[] | undefined, preferredTypes: string[]) {
  if (!assets?.length) {
    return ''
  }

  for (const type of preferredTypes) {
    const asset = assets.find((item) =>
      item.type === type
      && item.source === 'store_browse'
      && item.exists !== false
      && Boolean(item.url)
    )
    if (asset?.url) {
      return asset.url
    }
  }

  return ''
}

function mapV2DetailNews(news: GameV2NewsItem) {
  return {
    headline: news.headline,
    content: news.html || news.plain_text || news.summary,
    post_time: news.published_at,
    author: '',
    url: news.url,
  }
}

function mapV2Movie(movie: GameV2Movie) {
  const extra = asRecord(movie.extra)
  const dashH264 = stringField(extra, 'dash_h264_url') || movie.url || ''
  const hlsH264 = stringField(extra, 'hls_h264_url')
  const mp4Url = stringField(extra, 'mp4_max_url') || stringField(extra, 'mp4_480_url')
  const webmUrl = stringField(extra, 'webm_max_url') || stringField(extra, 'webm_480_url')

  return {
    id: Number(movie.id || 0),
    name: movie.name,
    thumbnail: movie.thumbnail_url,
    dash_av1: stringField(extra, 'dash_av1_url'),
    dash_h264: dashH264,
    hls_h264: hlsH264,
    mp4_url: mp4Url,
    webm_url: webmUrl,
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' ? value as Record<string, unknown> : {}
}

function stringField(record: Record<string, unknown>, key: string) {
  const value = record[key]
  return typeof value === 'string' ? value : ''
}

function mapV2News(news: GameV2NewsItem): NewsBaseModel {
  return {
    id: news.id || news.event_gid,
    name: news.game_name,
    post_time: news.published_at,
    headline: news.headline,
    header: news.header_url,
    author: '',
    content: news.html || news.plain_text || news.summary,
    url: news.url,
  }
}

function platformsToText(platforms: Record<string, boolean> = {}) {
  return ['windows', 'mac', 'linux'].filter((key) => platforms[key]).join(', ')
}

function toUnixSeconds(value?: string) {
  if (!value) {
    return 0
  }
  const time = new Date(value).getTime()
  return Number.isFinite(time) ? Math.floor(time / 1000) : 0
}
