import type { Game } from '../../lib/types'

type SteamPrefillFields = Partial<Pick<Game, 'name' | 'name_en' | 'info' | 'info_en' | 'groups' | 'developers' | 'publishers' | 'header' | 'links'>>
export type SteamPrefill = SteamPrefillFields & { warnings?: string[] }

export function steamPrefillValues(data: SteamPrefill): SteamPrefillFields {
  return Object.fromEntries(Object.entries(data).filter(([key, value]) =>
    ['name', 'name_en', 'info', 'info_en', 'groups', 'developers', 'publishers', 'header', 'links'].includes(key)
    && (typeof value === 'string' ? value.trim().length > 0 : Array.isArray(value) && value.length > 0),
  ))
}
