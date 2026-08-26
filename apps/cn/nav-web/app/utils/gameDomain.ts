import type { GameReleasePrecision, GameV2FirstAvailable, GameV2Language, GameV2ReleaseState } from '~/types/game'

export type GameDisplayLocale = 'zh' | 'en'

export interface GameCalendarValue {
  precision: GameReleasePrecision
  exact_date: string | null
  year: number | null
  month: number | null
  quarter: number | null
  raw_text?: string
}

export interface GameReleaseDisplay {
  kind: 'released' | 'planned' | 'unknown'
  value: string
}

export function formatGameCalendar(value: GameCalendarValue, locale: GameDisplayLocale): string {
  const year = value.year
  const month = value.month
  switch (value.precision) {
    case 'day': {
      const parts = parseCalendarDate(value.exact_date)
      if (!parts) return fallbackCalendar(value, locale)
      if (locale === 'zh') return `${parts.year}年${parts.month}月${parts.day}日`
      return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC' })
        .format(new Date(Date.UTC(parts.year, parts.month - 1, parts.day)))
    }
    case 'month':
      if (!year || !month) return fallbackCalendar(value, locale)
      if (locale === 'zh') return `${year}年${month}月`
      return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'long', timeZone: 'UTC' })
        .format(new Date(Date.UTC(year, month - 1, 1)))
    case 'quarter':
      if (!year || !value.quarter) return fallbackCalendar(value, locale)
      return locale === 'zh' ? `${year}年第${value.quarter}季度` : `Q${value.quarter} ${year}`
    case 'year':
      return year ? (locale === 'zh' ? `${year}年` : String(year)) : fallbackCalendar(value, locale)
    case 'tba':
      return locale === 'zh' ? '敬请期待' : 'Coming Soon'
    default:
      return fallbackCalendar(value, locale)
  }
}

export function resolveGameReleaseDisplay(
  release: GameV2ReleaseState | null | undefined,
  firstAvailable: GameV2FirstAvailable | null | undefined,
  locale: GameDisplayLocale,
): GameReleaseDisplay {
  if (firstAvailable) {
    return {
      kind: 'released',
      value: formatGameCalendar({ ...firstAvailable, raw_text: '' }, locale),
    }
  }
  if (release?.availability === 'upcoming') {
    return { kind: 'planned', value: formatGameCalendar(release, locale) }
  }
  return { kind: 'unknown', value: locale === 'zh' ? '未知' : 'Unknown' }
}

export function formatGameLanguageName(language: GameV2Language, locale: GameDisplayLocale): string {
  let name = language.steam_name
  if (language.code) {
    try {
      name = new Intl.DisplayNames([locale === 'zh' ? 'zh-CN' : 'en-US'], { type: 'language' }).of(language.code) || name
    } catch {
      // Preserve Steam's official name when a runtime does not recognize the code.
    }
  }
  if (language.full_audio_supported === true) {
    return `${name}${locale === 'zh' ? '（完整音频）' : ' (Full Audio)'}`
  }
  return name
}

export function formatGameLanguages(languages: GameV2Language[], locale: GameDisplayLocale): string {
  return languages.map(language => formatGameLanguageName(language, locale)).filter(Boolean).join(locale === 'zh' ? '、' : ', ')
}

function fallbackCalendar(value: GameCalendarValue, locale: GameDisplayLocale): string {
  const raw = value.raw_text?.trim()
  return raw || (locale === 'zh' ? '未知' : 'Unknown')
}

function parseCalendarDate(value: string | null | undefined) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value || '')
  if (!match) return null
  return { year: Number(match[1]), month: Number(match[2]), day: Number(match[3]) }
}
