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

type PublicCitation = {
  title?: string
  url?: string
  source_type?: string
  snippet?: string
  score?: number
  chunk_index?: number
}

function normalizePublicCitations(value: unknown): PublicCitation[] {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((item) => {
    const source = item && typeof item === 'object' ? item as Record<string, unknown> : {}
    return {
      title: typeof source.title === 'string' ? source.title : '',
      url: typeof source.url === 'string' ? source.url : '',
      source_type: typeof source.source_type === 'string' ? source.source_type : '',
      snippet: typeof source.snippet === 'string'
        ? source.snippet
        : typeof source.content === 'string'
          ? source.content
          : '',
      score: typeof source.score === 'number' ? source.score : undefined,
      chunk_index: typeof source.chunk_index === 'number' ? source.chunk_index : undefined
    }
  })
}

function rewriteArchivePayload(event: string, payload: unknown) {
  const source = payload && typeof payload === 'object' ? { ...(payload as Record<string, unknown>) } : null
  if (!source) {
    return { event, payload }
  }

  if (event === 'sources' && Array.isArray(source.sources)) {
    const citations = normalizePublicCitations(source.sources)
    delete source.sources
    return {
      event: 'citations',
      payload: {
        ...source,
        citations
      }
    }
  }

  if (event === 'done') {
    if (Array.isArray(source.citations)) {
      return {
        event,
        payload: {
          ...source,
          citations: normalizePublicCitations(source.citations)
        }
      }
    }
    if (Array.isArray(source.sources)) {
      const citations = normalizePublicCitations(source.sources)
      delete source.sources
      return {
        event,
        payload: {
          ...source,
          citations
        }
      }
    }
  }

  return { event, payload }
}

function rewriteArchiveFrame(frame: string) {
  const lines = frame.split(/\r?\n/)
  const event = lines.find(line => line.startsWith('event:'))?.slice(6).trim() || 'message'
  const data = lines
    .filter(line => line.startsWith('data:'))
    .map(line => line.slice(5).trim())
    .join('\n')

  if (!data || (event !== 'sources' && event !== 'done')) {
    return `${frame}\n\n`
  }

  try {
    const parsed = JSON.parse(data)
    const rewritten = rewriteArchivePayload(event, parsed)
    return `event: ${rewritten.event}\ndata: ${JSON.stringify(rewritten.payload)}\n\n`
  } catch {
    return `${frame}\n\n`
  }
}

function rewriteArchiveStream(body: ReadableStream<Uint8Array>) {
  const decoder = new TextDecoder()
  const encoder = new TextEncoder()

  return new ReadableStream<Uint8Array>({
    async start(controller) {
      const reader = body.getReader()
      let buffer = ''

      try {
        while (true) {
          const { value, done } = await reader.read()
          if (done) {
            break
          }

          buffer += decoder.decode(value, { stream: true })
          const frames = buffer.split(/\r?\n\r?\n/)
          buffer = frames.pop() || ''

          frames.forEach((frame) => {
            controller.enqueue(encoder.encode(rewriteArchiveFrame(frame)))
          })
        }

        buffer += decoder.decode()
        if (buffer.trim()) {
          controller.enqueue(encoder.encode(rewriteArchiveFrame(buffer)))
        }
        controller.close()
      } catch (error) {
        controller.error(error)
      } finally {
        reader.releaseLock()
      }
    }
  })
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

    return rewriteArchiveStream(response.body)
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
