import { createError, setResponseStatus } from 'h3'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  try {
    const response = await $fetch.raw('/api/v1/chat/status', {
      baseURL: config.ragApiInternalBase,
      responseType: 'json'
    })
    setResponseStatus(event, response.status)
    return response._data
  } catch (error: any) {
    if (error?.response) {
      setResponseStatus(event, error.response.status)
      return error.response._data
    }

    throw createError({
      statusCode: 502,
      statusMessage: 'Unable to reach RAG service'
    })
  }
})
