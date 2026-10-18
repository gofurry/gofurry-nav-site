import type { ComparePickerEntity } from '../types/insightCompare'

export function filterCompareSites(items: ComparePickerEntity[], keyword: string, selected: number[]) {
  const query = keyword.trim().toLowerCase()
  const relevance = (item: ComparePickerEntity) => {
    const values = [item.name, item.subtitle || ''].map(value => value.toLowerCase())
    if (!query || values.includes(query)) return 0
    if (values.some(value => value.startsWith(query))) return 1
    return values.some(value => value.includes(query)) ? 2 : 3
  }
  return items.map((item, index) => ({ item, index, match: relevance(item) }))
    .filter(({ item, match }) => !selected.includes(item.id) && match < 3)
    .sort((a, b) => a.match - b.match || a.index - b.index).slice(0, 10).map(({ item }) => item)
}

export function compareSelectedEntities(ids: number[], authoritative: ComparePickerEntity[], cached: ComparePickerEntity[]) {
  return ids.map(id => authoritative.find(item => item.id === id) || cached.find(item => item.id === id) || { id, name: '#' + id })
}
