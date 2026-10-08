import { z } from 'zod'
import { getJSON, listJSON } from '../../lib/api'
import type { CollectionComposition, CollectionContent, CollectionWorkspace, GameCollection } from './types'

export const collectionEndpoint = '/api/v1/game/collections'
export const eligibleCollectionEndpoint = `${collectionEndpoint}?status=published&home_eligible=true`
export const workspaceKey = (id: string | number) => ['game-collection', String(id)]
export const homeKey = ['game-collection-home']
const boundedText = (max: number) => z.string().trim().refine(value => Array.from(value).length <= max, `最多 ${max} 个字符`)
export const collectionContentSchema = z.object({
  name: boundedText(160).refine(Boolean, '请输入中文名称'), name_en: boundedText(160).refine(Boolean, '请输入英文名称'),
  info: boundedText(500), info_en: boundedText(500),
})
export const createCollectionSchema = collectionContentSchema.extend({ code: z.string().min(1, '请输入 Code').max(64).regex(/^[a-z0-9]+(-[a-z0-9]+)*$/, '使用小写字母、数字及单个连字符') })
export function contentOf(c: CollectionContent): CollectionContent { return { name: c.name, name_en: c.name_en, info: c.info, info_en: c.info_en } }

// Both read endpoints own their version. Do not combine composition from a newer
// version with an older content snapshot; retry only reads if a writer races us.
export async function loadCollectionWorkspace(id: string | number): Promise<CollectionWorkspace> {
  for (let attempt = 0; attempt < 2; attempt++) {
    const [collection, composition] = await Promise.all([
      getJSON<GameCollection>(`${collectionEndpoint}/${id}`),
      getJSON<CollectionComposition>(`${collectionEndpoint}/${id}/composition`),
    ])
    if (collection.version === composition.version) return { collection: collectionWithComposition(collection, composition), composition }
  }
  throw new Error('游戏分区正在更新，请重新加载后重试。')
}
export async function loadEligibleCollections(keyword: string, pageSize: number) {
  const result = await listJSON<GameCollection>(eligibleCollectionEndpoint, 1, pageSize, keyword)
  return { ...result, list: result.list.map(c => ({ id: String(c.id), label: c.name, extra: `${c.code} · SFW 可见 ${c.sfw_member_count}` })) }
}

// Composition owns the membership snapshot; independent Tag edits need not change version.
export function collectionWithComposition(collection: GameCollection, composition: CollectionComposition): GameCollection {
  return { ...collection, version: composition.version, home_slot: composition.home_slot, member_count: composition.counts.effective, sfw_member_count: composition.counts.sfw_visible }
}
