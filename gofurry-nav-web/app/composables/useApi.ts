import type { NitroFetchOptions } from 'nitropack'
import { ApiError, type ApiResult, type ApiService } from '~/types/api'

type ApiRequestOptions = NitroFetchOptions<string>
type ApiRequest = <T>(path: string, options?: ApiRequestOptions) => Promise<T>

function resolveBaseURL(service: ApiService) {
  const config = useRuntimeConfig()

  if (service === 'game') {
    return import.meta.server
      ? config.gameApiInternalBase
      : config.public.gameApiBase
  }

  if (service === 'navV2') {
    return import.meta.server
      ? config.navV2ApiInternalBase
      : config.public.navV2ApiBase
  }

  return import.meta.server
    ? config.navApiInternalBase
    : config.public.navApiBase
}

export const useApi = (service: ApiService = 'nav'): ApiRequest => {
  const baseURL = resolveBaseURL(service)

  return async <T>(path: string, options: ApiRequestOptions = {}) => {
    const response = await $fetch<ApiResult<T>>(path, {
      baseURL,
      credentials: 'include',
      ...options
    })

    if (!response || response.code !== 1) {
      throw new ApiError(response?.message || response?.msg || '接口返回错误', undefined, response?.code)
    }

    return response.data
  }
}
