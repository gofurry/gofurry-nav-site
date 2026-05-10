import { createError, getHeaders, readRawBody, setHeader, setResponseStatus } from 'h3'

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

function sanitizeHeaders(headers: Record<string, string | undefined>) {
  return Object.entries(headers).filter(
    (entry): entry is [string, string] => Boolean(entry[1]) && !hopByHopHeaders.has(entry[0].toLowerCase())
  )
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readRawBody(event)

  try {
    const response = await fetch(`${String(config.ragApiInternalBase).replace(/\/$/, '')}/api/v1/chat/stream`, {
      method: 'POST',
      headers: {
        ...Object.fromEntries(sanitizeHeaders(getHeaders(event))),
        'content-type': 'application/json',
        accept: 'text/event-stream'
      },
      body
    })

    setResponseStatus(event, response.status)
    setHeader(event, 'content-type', response.headers.get('content-type') || 'text/event-stream; charset=utf-8')
    setHeader(event, 'cache-control', 'no-cache, no-transform')
    setHeader(event, 'x-accel-buffering', 'no')

    if (!response.body) {
      throw createError({
        statusCode: 502,
        statusMessage: 'RAG service returned an empty stream'
      })
    }

    return response.body
  } catch (error: any) {
    if (error?.statusCode) {
      throw error
    }

    throw createError({
      statusCode: 502,
      statusMessage: 'Unable to reach RAG service'
    })
  }
})
