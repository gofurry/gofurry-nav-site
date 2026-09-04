export type AuthoritativeEntity = 'game' | 'site'

const notFoundMessages: Record<AuthoritativeEntity, string[]> = {
  game: [
    '查询站内游戏主档案失败: game not found',
    '目标游戏不存在或缺少 v2 详情',
  ],
  site: ['站点不存在'],
}

export function authoritativePageStatus(error: unknown, entity: AuthoritativeEntity): 404 | 503 {
  if (errorStatus(error) === 404) {
    return 404
  }

  const message = errorMessage(error)
  if (notFoundMessages[entity].some(marker => message.includes(marker))) {
    return 404
  }

  return 503
}

function errorStatus(error: unknown): number | undefined {
  const record = asRecord(error)
  const response = asRecord(record.response)
  for (const value of [record.statusCode, record.status, response.statusCode, response.status]) {
    if (typeof value === 'number' && Number.isFinite(value)) {
      return value
    }
  }
  return undefined
}

function errorMessage(error: unknown): string {
  const record = asRecord(error)
  const response = asRecord(record.response)
  const responseData = asRecord(response._data)
  const data = asRecord(record.data)
  return [
    record.message,
    record.statusMessage,
    typeof record.data === 'string' ? record.data : '',
    data.message,
    typeof response._data === 'string' ? response._data : '',
    responseData.message,
    typeof responseData.data === 'string' ? responseData.data : '',
  ].filter(value => typeof value === 'string').join(' ')
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' ? value as Record<string, any> : {}
}
