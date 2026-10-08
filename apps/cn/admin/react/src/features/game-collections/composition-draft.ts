import type { CollectionComposition, CollectionMember, CompositionDraft, EffectiveCollectionMember } from './types'

const canonicalIDs = (ids: number[]) => [...new Set(ids)].sort((a, b) => a - b)
export const compositionDraftOf = (value: CollectionComposition): CompositionDraft => ({ rule_tags: value.rule_tags, manual_members: value.manual_members, excluded_members: value.excluded_members })
export const compositionIDs = (value: CompositionDraft) => ({
  tag_ids: canonicalIDs(value.rule_tags.map(tag => tag.tag_id)),
  manual_game_ids: canonicalIDs(value.manual_members.map(game => game.game_id)),
  excluded_game_ids: canonicalIDs(value.excluded_members.map(game => game.game_id)),
})
export const compositionChanged = (value: CompositionDraft, base: CompositionDraft) => JSON.stringify(compositionIDs(value)) !== JSON.stringify(compositionIDs(base))

export function changeMemberOverride(draft: CompositionDraft, member: CollectionMember, action: 'pin' | 'unpin' | 'exclude' | 'restore'): CompositionDraft {
  const manual = draft.manual_members.filter(game => game.game_id !== member.game_id)
  const excluded = draft.excluded_members.filter(game => game.game_id !== member.game_id)
  if (action === 'pin') return { ...draft, manual_members: [...manual, member], excluded_members: excluded }
  if (action === 'unpin') return { ...draft, manual_members: manual }
  if (action === 'exclude') return { ...draft, manual_members: manual, excluded_members: [...excluded, member] }
  return { ...draft, excluded_members: excluded }
}

// Only preview overrides on known server matches. New/removed Tag rules and
// restored server exclusions must be resolved by the Backend after saving.
export function previewComposition(base: CollectionComposition, draft: CompositionDraft): EffectiveCollectionMember[] {
  const rows = new Map<number, EffectiveCollectionMember>()
  for (const member of base.effective_members) {
    if (member.source !== 'manual') rows.set(member.game_id, { ...member, source: 'automatic' })
  }
  for (const member of draft.manual_members) rows.set(member.game_id, { ...member, source: rows.has(member.game_id) ? 'both' : 'manual' })
  for (const member of draft.excluded_members) rows.delete(member.game_id)
  return [...rows.values()]
}
