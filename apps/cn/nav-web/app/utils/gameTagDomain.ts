import type { GameTagCategory, GameTagRecord } from '../types/game'

export type GameTagGroup = GameTagCategory & {
  children: (GameTagRecord & { selected: boolean })[]
  expanded: boolean
  limit: number
}

export function buildGameTagGroups(categories: GameTagCategory[], selectedIDs: number[], previous: GameTagGroup[] = []): GameTagGroup[] {
  return categories.map(category => ({
    ...category,
    expanded: previous.find(group => group.id === category.id)?.expanded ?? false,
    limit: 16,
    children: category.tags.map(tag => ({ ...tag, selected: selectedIDs.includes(Number(tag.id)) })),
  }))
}

export function hasAdultTag(tags: readonly { code: string }[] | null | undefined): boolean {
  return tags?.some(tag => tag.code === 'adult') ?? false
}
