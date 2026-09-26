export type Kind = 'game' | 'site' | 'other'
export type Status = 'idea' | 'researching' | 'landed' | 'shelved'
export type IdeaInput = { kind: Kind; title: string; source: string; note: string; priority: 'normal' | 'high'; version?: number }
export type Idea = Omit<IdeaInput, 'title' | 'source' | 'version'> & {
  id: number; title: string | null; source: string | null; source_key: string | null; version: number; status: Status
  creator_name: string; researcher_name: string; created_by_account_id: number; researching_by_account_id: number | null
  linked_kind: 'game' | 'site' | null; linked_resource_id: number | null; created_at: string
}
export type PreviewRow = IdeaInput & { index: number; source_key: string | null; display_hint: string; valid: boolean; errors: string[]; warnings: string[]; idea_match: { id: number; title: string } | null; resource_match: { id: number; kind: Kind; title: string } | null; batch_duplicate_of: number | null }
export type Inventory = { reserve_count: number; researching_count: number; landed_30d: number }
export type BoardNote = { id: number; body: string; x: number; y: number; width: number; height: number; z_index: number; version: number; updated_by_account_id: number }
export const kindOptions = [{ value: 'game', label: 'Game' }, { value: 'site', label: 'Site' }, { value: 'other', label: 'Other' }]
export const statusLabels: Record<Status, string> = { idea: '储备', researching: '整理中', landed: '已落地', shelved: '搁置' }
export const priorityOptions = [{ value: 'normal', label: '普通' }, { value: 'high', label: '优先' }]
export const warningLabels: Record<string, string> = { batch_duplicate: '本批次重复', idea_duplicate: '内容池已有相似记录', resource_exists: '正式 Game / Site 已存在' }
export function ideaInput(idea: Idea): IdeaInput { return { kind: idea.kind, title: idea.title ?? '', source: idea.source ?? '', note: idea.note, priority: idea.priority, version: idea.version } }
