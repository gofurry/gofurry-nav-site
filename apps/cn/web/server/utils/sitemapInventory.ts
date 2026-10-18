export interface SitemapEntity {
  id: string
}

export function parseUpdateInventory(response: unknown): SitemapEntity[] {
  const data = asRecord(successData(response, 'Release Notes'))
  if (data.schema_version !== 1 || !['ready', 'empty'].includes(data.state)
    || !Array.isArray(data.items) || (data.state === 'empty' && data.items.length > 0)
    || (data.state === 'ready' && data.items.length === 0)) {
    throw new Error('Release Notes inventory has an invalid data shape')
  }
  return parseEntities(data.items, 'Release Notes', item => {
    const id = asRecord(item).id
    if (typeof id !== 'number' || !Number.isSafeInteger(id) || id <= 0) {
      throw new Error('Release Notes inventory has an invalid identity')
    }
    return id
  })
}

export function parseSiteInventory(response: unknown): SitemapEntity[] {
  const data = successData(response, 'Nav Sites')
  const record = asRecord(data)
  if (!Array.isArray(record.items)) {
    throw new Error('Nav Sites inventory has an invalid data shape')
  }
  return parseEntities(record.items, 'Nav Sites', item => asRecord(item).id)
}

export function parseSiteGroupInventory(response: unknown): SitemapEntity[] {
  const data = successData(response, 'Site Groups')
  if (!Array.isArray(data)) {
    throw new Error('Site Groups inventory has an invalid data shape')
  }
  return parseEntities(data, 'Site Groups', item => asRecord(item).id)
}

export function parseGameInventory(response: unknown): SitemapEntity[] {
  const data = successData(response, 'Games')
  if (!Array.isArray(data)) {
    throw new Error('Games inventory has an invalid data shape')
  }
  return parseEntities(data, 'Games', (item) => {
    const record = asRecord(item)
    return record.id ?? record.game_id
  })
}

function successData(response: unknown, inventory: string): unknown {
  const record = asRecord(response)
  if (record.code !== 1 || !Object.prototype.hasOwnProperty.call(record, 'data')) {
    throw new Error(`${inventory} inventory returned a non-success API envelope`)
  }
  return record.data
}

function parseEntities(items: unknown[], inventory: string, idOf: (item: unknown) => unknown): SitemapEntity[] {
  return items.map((item, index) => {
    const id = idOf(item)
    if ((typeof id !== 'string' && typeof id !== 'number') || String(id).trim() === '') {
      throw new Error(`${inventory} inventory item ${index} has an invalid identity`)
    }
    return { id: String(id) }
  })
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}
