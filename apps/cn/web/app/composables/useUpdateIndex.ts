import { computed, onScopeDispose, ref, watch, type Ref } from 'vue'
import { getNavUpdates } from '~/services/nav'
import type { NavUpdateIndexItem } from '~/types/nav'

// One Latest + twenty historical releases on SSR; later pages require a click.
export const updateIndexPageSize = 21

export async function useUpdateIndex(lang: Ref<'zh' | 'en'>) {
  const first = useAsyncData(() => `updates-index:${lang.value}`, () => getNavUpdates(lang.value, { page: 1, page_size: updateIndexPageSize }))
  const appended = ref<NavUpdateIndexItem[]>([])
  const page = ref(1), moreAvailable = ref(false), loadingMore = ref(false), moreError = ref(false)
  let generation = 0, controller: AbortController | undefined
  const reset = () => {
    generation++; controller?.abort(); controller = undefined
    appended.value = []; page.value = 1; loadingMore.value = false; moreError.value = false
  }
  watch(lang, reset, { flush: 'sync' })
  watch(first.data, reset, { flush: 'sync' })
  onScopeDispose(() => { generation++; controller?.abort() })
  const items = computed(() => [...(first.data.value?.items ?? []), ...appended.value])
  const hasMore = computed(() => page.value === 1 ? first.data.value?.has_more ?? false : moreAvailable.value)
  const state = computed(() => first.pending.value ? 'loading' : first.error.value || first.data.value?.state === 'error' ? 'error' : items.value.length ? 'ready' : 'empty')

  async function loadMore() {
    if (loadingMore.value || !hasMore.value || state.value !== 'ready') return
    const token = generation, requestedPage = page.value + 1, requestedLang = lang.value
    const request = new AbortController()
    controller = request; loadingMore.value = true; moreError.value = false
    try {
      const response = await getNavUpdates(requestedLang, { page: requestedPage, page_size: updateIndexPageSize }, request.signal)
      if (token !== generation || request.signal.aborted) return
      if (response.state === 'error' || response.page !== requestedPage) throw new Error('Release page unavailable')
      const ids = new Set(items.value.map(item => item.id))
      appended.value.push(...response.items.filter(item => !ids.has(item.id)))
      page.value = requestedPage; moreAvailable.value = response.has_more
    } catch {
      if (token === generation && !request.signal.aborted) moreError.value = true
    } finally {
      if (token === generation) { loadingMore.value = false; controller = undefined }
    }
  }
  await first
  return { items, total: computed(() => first.data.value?.total ?? 0), pending: first.pending, state,
    refresh: first.refresh, hasMore, loadingMore, moreError, loadMore }
}
