export type ApiResult<T> = { code: number; message: string; data: T }
export type PageResult<T> = { total: number; list: T[] }

export type AuthIdentity = {
  account_id: number
  username: string
  display_name: string
  role: 'owner' | 'developer' | 'operator'
  status: 'active' | 'disabled'
  session_version: number
  capabilities: string[]
}

export type AuthState = { initialized: boolean; authenticated: boolean; identity?: AuthIdentity }
export type OptionItem = { id: string; label: string; extra?: string }
export type KeyValue = { key: string; value: string }

export type Site = {
  id: number; name: string; name_en: string; info: string; info_en: string
  create_time: string; update_time: string; country: string | null
  nsfw: string; welfare: string; icon: string | null; deleted: boolean
}

export type CollectorTarget = {
  id: number; site_id: number; name: string; proxy: string; prefix: string | null; tls: string; primary: boolean
}

export type SiteSummary = { id: number; name: string; name_en: string; update_time: string; primary_target: string; group_names: string[]; featured: boolean }

export type SiteGroupRelation = {
  id: number; site_id: number; group_id: number; group_name: string; weight: number
}

export type FeaturedSite = { id: number; site_id: number; weight: number }
export type SiteWorkspace = { site: Site; targets: CollectorTarget[]; groups: SiteGroupRelation[]; featured: FeaturedSite | null }

export type Game = {
  id: number; name: string; name_en: string; info: string; info_en: string
  create_time: string; update_time: string; resources: KeyValue[]; groups: KeyValue[]
  developers: string[]; publishers: string[]; appid: number; header: string; links: KeyValue[]
  weight: number; primary_tag: number; secondary_tag: number
}

export type GameTagRelation = { id: number; game_id: number; tag_id: number; tag_name: string }
export type GameWorkspace = { game: Game; tags: GameTagRelation[] }
