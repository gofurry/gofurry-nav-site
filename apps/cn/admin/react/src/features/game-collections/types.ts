export type CollectionStatus = 'draft' | 'published' | 'archived'
export type CollectionContent = { name: string; name_en: string; info: string; info_en: string }
export type GameCollection = CollectionContent & {
  id: number; code: string; status: CollectionStatus; version: number
  published_at: string | null; archived_at: string | null; created_at: string; updated_at: string
  member_count: number; sfw_member_count: number; home_slot: number | null
}
export type CollectionMember = { game_id: number; name: string; name_en: string; appid: number; adult: boolean }
export type CollectionMembers = { collection_id: number; version: number; members: CollectionMember[] }
export type CollectionRuleTag = { tag_id: number; code: string; name: string; name_en: string; active: boolean }
export type EffectiveCollectionMember = CollectionMember & { source: 'automatic' | 'manual' | 'both' }
export type CollectionComposition = {
  collection_id: number; version: number; home_slot: number | null
  rule_tags: CollectionRuleTag[]; manual_members: CollectionMember[]; excluded_members: CollectionMember[]
  effective_members: EffectiveCollectionMember[]
  counts: { auto_matched: number; manual_pinned: number; excluded: number; effective: number; sfw_visible: number }
}
export type CompositionDraft = Pick<CollectionComposition, 'rule_tags' | 'manual_members' | 'excluded_members'>
export type CollectionWorkspace = { collection: GameCollection; composition: CollectionComposition }
export type CollectionHome = { revision: string; slots: { slot: number; collection: GameCollection | null }[] }
export const collectionStatusLabels: Record<CollectionStatus, string> = { draft: '草稿', published: '已发布', archived: '已归档' }
