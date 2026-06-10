import type {
  AnonymousReviewModel,
  CommentReq,
  CreatorResponse,
  GameBaseInfoResponse,
  GameGroupRecord,
  GamePanelRecord,
  GameV2DetailRecord,
  GameV2ListItem,
  GameV2NewsItem,
  GameV2PanelRecord,
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

interface GameHomeData {
  mainInfo: GameGroupRecord
  panelData: GamePanelRecord
  latestNews: LatestNewsRecord
}

export function getGameList() {
  return useApi('game')('/game/info/list')
}

export async function getGameHomeData(lang = 'zh'): Promise<GameHomeData> {
  const [panel, newsPair] = await Promise.all([
    getGameV2Panel(lang),
    getLatestGameNews(),
  ])
  return {
    mainInfo: mapV2PanelToGameGroup(panel),
    panelData: mapV2PanelToGamePanel(panel),
    latestNews: newsPair,
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
  return useApi('gameV2')<GameV2DetailRecord>('/game/info', {
    query: { id, lang: normalizeGameLang(lang), region: 'CN', news_limit: 20 }
  }).then(mapV2Detail)
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

function getGameV2Panel(lang: string): Promise<GameV2PanelRecord> {
  return useApi('gameV2')('/game/panel/main', {
    query: {
      lang: normalizeGameLang(lang),
      region: 'CN',
      limit: 24,
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
    hot: panel.top_online.map(mapV2ListItemToBase),
    free: panel.free_games.map(mapV2ListItemToBase),
  }
}

function mapV2PanelToGamePanel(panel: GameV2PanelRecord): GamePanelRecord {
  const topOnline = panel.top_online.map(mapV2ListItemToTopPlayer)
  const lowPrice = panel.low_price.map(mapV2ListItemToPrice)

  return {
    top_count: {
      one: topOnline.slice(0, 15),
      two: topOnline.slice(0, 30),
      three: topOnline.slice(0, 45),
      four: topOnline.slice(0, 60),
    },
    top_discount_vo: panel.highest_discount.map(mapV2ListItemToPrice),
    top_price_vo: lowPrice,
    bottom_price: {
      one: lowPrice.filter((item) => item.china_price > 0 && item.china_price <= 1000),
      two: lowPrice.filter((item) => item.china_price > 1000 && item.china_price <= 1500),
      three: lowPrice.filter((item) => item.china_price > 1500 && item.china_price <= 2000),
      four: lowPrice.filter((item) => item.china_price > 2000 && item.china_price <= 2500),
    },
  }
}

function mapV2ListItemToBase(game: GameV2ListItem) {
  return {
    game_id: game.id,
    avg_score: 0,
    comment_count: 0,
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
    count_peak: game.online_count?.count ?? 0,
    count_recent: game.online_count?.count ?? 0,
    collect_time: toUnixSeconds(game.online_count?.collected_at),
    header: bestV2Cover(game),
  }
}

function mapV2ListItemToPrice(game: GameV2ListItem) {
  const price = game.price
  return {
    id: game.id,
    name: game.name,
    global_price: 0,
    china_price: price?.available ? price.final_amount : 0,
    discount: price?.discount_percent ?? 0,
    header: bestV2Cover(game),
  }
}

function bestV2Cover(game: GameV2ListItem) {
  return game.header_url || game.capsule_url || ''
}

function mapV2Detail(detail: GameV2DetailRecord): GameBaseInfoResponse {
  return {
    name: detail.name,
    info: detail.summary || detail.short_description,
    create_time: detail.site?.create_time || detail.collected_at || '',
    update_time: detail.site?.update_time || detail.updated_at || '',
    resources: detail.site?.resources ?? [],
    groups: detail.site?.groups ?? [],
    links: detail.site?.links ?? [],
    release_date: detail.release?.date ?? '',
    developers: detail.developers ?? [],
    publishers: detail.publishers ?? [],
    appid: Number(detail.appid || 0),
    cover: detail.media?.header_url || detail.header_url || detail.site?.header || '',
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
    screenshots: (detail.media?.screenshots ?? []).map((item) => ({
      id: Number(item.id || 0),
      path_thumbnail: item.thumbnail_url,
      path_full: item.url,
    })),
    movies: (detail.media?.movies ?? []).map((item) => ({
      id: Number(item.id || 0),
      name: item.name,
      thumbnail: item.thumbnail_url,
      dash_av1: '',
      dash_h264: item.url,
      hls_h264: item.url,
    })),
    pc_requirements: {
      id: Number(detail.id || 0),
      minimum: detail.requirements?.pc?.minimum ?? '',
      recommended: detail.requirements?.pc?.recommended ?? '',
    },
    online_count: detail.online_count?.count ?? 0,
    count_collect_time: detail.online_count?.collected_at ?? '',
    view_count: detail.site?.view_count ?? 0,
  }
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
