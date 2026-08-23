type SeoLocale = 'zh' | 'en' | string

type DetailSeo = {
  title: string
  description: string
}

type SiteDetailSeoInput = {
  name?: string | null
  description?: string | null
  domain?: string | null
  locale?: SeoLocale
}

type GameDetailSeoInput = {
  name?: string | null
  description?: string | null
  locale?: SeoLocale
}

const MAX_TITLE_LENGTH = 78
const MIN_DESCRIPTION_LENGTH = 80
const MAX_DESCRIPTION_LENGTH = 180

function isEnglish(locale?: SeoLocale) {
  return locale === 'en' || locale === 'en-US'
}

function stripHtml(value: string) {
  return value
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
}

function normalizeText(value?: string | null) {
  return stripHtml(String(value || ''))
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/g, "'")
    .replace(/\s+/g, ' ')
    .trim()
}

function truncateText(value: string, maxLength: number) {
  if (value.length <= maxLength) {
    return value
  }

  const clipped = value.slice(0, maxLength - 1).trim()
  return `${clipped}…`
}

function uniqueParts(parts: string[]) {
  const seen = new Set<string>()
  const result: string[] = []

  for (const part of parts) {
    const normalized = normalizeText(part)
    if (!normalized || seen.has(normalized)) {
      continue
    }

    seen.add(normalized)
    result.push(normalized)
  }

  return result
}

function buildDescription(primary: string | null | undefined, fallbacks: string[]) {
  const parts = uniqueParts([normalizeText(primary), ...fallbacks])
  let description = ''

  for (const part of parts) {
    description = description ? `${description} ${part}` : part
    if (description.length >= MIN_DESCRIPTION_LENGTH) {
      break
    }
  }

  return truncateText(description, MAX_DESCRIPTION_LENGTH)
}

function buildTitle(primary: string, fallback: string) {
  return truncateText(normalizeText(primary) || fallback, MAX_TITLE_LENGTH)
}

export function buildSiteDetailSeo(input: SiteDetailSeoInput): DetailSeo {
  const en = isEnglish(input.locale)
  const name = normalizeText(input.name) || (en ? 'Furry Site' : '兽人站点')
  const domain = normalizeText(input.domain)
  const title = buildTitle(
    en
      ? `${name} - Furry Site Navigation, Status and Resource Details | GoFurry`
      : `${name} - 兽人站点导航、可用性监测与资源详情 | GoFurry`,
    en
      ? 'Furry Site Details - GoFurry Navigation'
      : '兽人站点详情 - GoFurry 导航'
  )

  const description = buildDescription(input.description, en
    ? [
        domain ? `Listed domain: ${domain}.` : '',
        'GoFurry provides the site entry, summary, visit signals, and user-perspective HTTP, DNS, TLS, and availability observations for furry culture resources.',
        'Use this page to discover communities, art, fiction, games, tools, and related navigation context.'
      ]
    : [
        domain ? `收录域名：${domain}。` : '',
        'GoFurry 提供站点入口、简介、访问量信号，以及用户视角的 HTTP、DNS、TLS 与可用性观测信息。',
        '本页帮助兽人文化爱好者发现社区、创作、小说、游戏、工具等相关资源，并判断站点当前访问状态。'
      ])

  return { title, description }
}

export function buildGameDetailSeo(input: GameDetailSeoInput): DetailSeo {
  const en = isEnglish(input.locale)
  const name = normalizeText(input.name) || (en ? 'Furry Game' : '兽人游戏')
  const title = buildTitle(
    en
      ? `${name} - Furry Game Info, Reviews and Related Resources | GoFurry`
      : `${name} - 兽人游戏资料、评价与相关资源 | GoFurry`,
    en
      ? 'Furry Game Details - GoFurry'
      : '兽人游戏详情 - GoFurry'
  )

  const description = buildDescription(input.description, en
    ? [
        'GoFurry collects furry and anthro game information, introductions, covers, review signals, related titles, and community resource links.',
        'Use this page to understand the game and continue discovering similar works.'
      ]
    : [
        'GoFurry 收录兽人、拟人和相关题材游戏资料，整理简介、封面、评价信号、相关作品与社区资源入口。',
        '本页帮助你了解游戏内容，并继续发现相近作品与更新信息。'
      ])

  return { title, description }
}
