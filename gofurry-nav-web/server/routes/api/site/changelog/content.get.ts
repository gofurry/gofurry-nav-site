import { createError, getQuery } from 'h3'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const rawUrl = typeof query.url === 'string' ? query.url.trim() : ''

  if (!rawUrl) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Missing changelog url',
    })
  }

  let target: URL
  try {
    target = new URL(rawUrl)
  } catch {
    throw createError({
      statusCode: 400,
      statusMessage: 'Invalid changelog url',
    })
  }

  if (!['http:', 'https:'].includes(target.protocol)) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Unsupported changelog protocol',
    })
  }

  try {
    const response = await fetch(target, {
      headers: {
        accept: 'text/markdown,text/plain;q=0.9,*/*;q=0.8',
      },
    })

    if (!response.ok) {
      throw createError({
        statusCode: response.status,
        statusMessage: 'Unable to load changelog content',
      })
    }

    const buffer = await response.arrayBuffer()
    return new TextDecoder('utf-8').decode(buffer)
  } catch (error: any) {
    if (error?.statusCode) {
      throw error
    }

    throw createError({
      statusCode: 502,
      statusMessage: 'Unable to load changelog content',
    })
  }
})
