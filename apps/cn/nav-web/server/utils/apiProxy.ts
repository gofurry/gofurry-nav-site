import {
  createError,
  getHeaders,
  getMethod,
  getQuery,
  readRawBody,
  setHeader,
  setResponseStatus,
  type H3Event
} from 'h3'

type ApiService = 'nav' | 'navV2' | 'game' | 'gameV2'

const hopByHopHeaders = new Set([
  'connection',
  'content-length',
  'host',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailer',
  'transfer-encoding',
  'upgrade'
])

function normalizeCatchAll(value: string | string[] | undefined) {
  if (!value) {
    return ''
  }

  return Array.isArray(value) ? value.join('/') : value
}

function sanitizeRequestHeaders(headers: Record<string, string | undefined>) {
  return Object.entries(headers).filter(
    (entry): entry is [string, string] => Boolean(entry[1]) && !hopByHopHeaders.has(entry[0].toLowerCase())
  )
}

function resolveTargetBase(event: H3Event, service: ApiService) {
  const config = useRuntimeConfig(event)
  if (service === 'game') {
    return config.gameApiInternalBase
  }
  if (service === 'gameV2') {
    return config.gameV2ApiInternalBase
  }
  if (service === 'navV2') {
    return config.navV2ApiInternalBase
  }
  return config.navApiInternalBase
}

export async function proxyApiNamespace(event: H3Event, service: ApiService, namespace: string) {
  const method = getMethod(event)
  const suffix = normalizeCatchAll(event.context.params?.path)
  const path = `/${namespace}${suffix ? `/${suffix}` : ''}`
  const body = method === 'GET' || method === 'HEAD' ? undefined : await readRawBody(event)

  try {
    const response = await $fetch.raw(path, {
      baseURL: resolveTargetBase(event, service),
      method,
      query: getQuery(event),
      body,
      headers: sanitizeRequestHeaders(getHeaders(event)),
      responseType: 'text',
      redirect: 'manual'
    })

    setResponseStatus(event, response.status)

    const location = response.headers.get('location')
    if (location) {
      setHeader(event, 'location', location)
    }

    const contentType = response.headers.get('content-type')
    if (contentType) {
      setHeader(event, 'content-type', contentType)
    }

    return response._data
  } catch (error: any) {
    if (error?.response) {
      setResponseStatus(event, error.response.status)

      const contentType = error.response.headers?.get?.('content-type')
      if (contentType) {
        setHeader(event, 'content-type', contentType)
      }

      return error.response._data
    }

    throw createError({
      statusCode: 502,
      statusMessage: `Unable to reach ${service} API service`
    })
  }
}
