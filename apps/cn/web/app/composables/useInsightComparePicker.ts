import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavSiteDirectory } from '@/services/nav'
import { getSearchSimple } from '@/utils/api/game'
import type { InsightDomain } from '@/types/insights'
import type { ComparePickerEntity } from '@/types/insightCompare'

export function useInsightComparePicker(domain: InsightDomain) {
  const { locale } = useI18n()
  const keyword = ref('')
  const items = ref<ComparePickerEntity[]>([])
  const cache = ref<ComparePickerEntity[]>([])
  const loading = ref(false)
  const failed = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined
  let controller: AbortController | undefined
  let generation = 0
  let mounted = false
  const remember = (entities: ComparePickerEntity[]) => {
    const merged = new Map(cache.value.map(item => [item.id, item]))
    for (const item of entities) merged.set(item.id, item)
    cache.value = [...merged.values()]
  }
  async function directory() {
    const token = ++generation
    loading.value = true
    failed.value = false
    try {
      const sites = await getNavSiteDirectory(locale.value)
      if (token !== generation) return
      items.value = sites.map(site => ({ id: Number(site.id), name: site.name, subtitle: site.domain,
        visual: site.icon ? { kind: 'site_icon' as const, asset: site.icon } : undefined,
      })).filter(item => Number.isSafeInteger(item.id) && item.id > 0)
      remember(items.value)
    } catch { if (token === generation) failed.value = true }
    finally { if (token === generation) loading.value = false }
  }
  watch(keyword, value => {
    if (domain !== 'game') return
    clearTimeout(timer)
    controller?.abort()
    const token = ++generation
    items.value = []
    failed.value = false
    loading.value = false
    if (!value.trim()) return
    loading.value = true
    timer = setTimeout(async () => {
      controller = new AbortController()
      try {
        const result = await getSearchSimple(locale.value, value.trim(), { signal: controller.signal })
        if (token !== generation) return
        items.value = result.map(game => ({ id: Number(game.id), name: game.name, subtitle: game.info,
          visual: game.cover ? { kind: 'game_header' as const, asset: game.cover } : undefined,
        })).filter(item => Number.isSafeInteger(item.id) && item.id > 0)
        remember(items.value)
      } catch { if (token === generation) failed.value = true }
      finally { if (token === generation) loading.value = false }
    }, 350)
  })
  watch(locale, () => {
    keyword.value = ''
    items.value = []
    if (mounted && domain === 'site') void directory()
  })
  onMounted(() => { mounted = true; if (domain === 'site') void directory() })
  onBeforeUnmount(() => { generation++; controller?.abort(); clearTimeout(timer) })
  return { keyword, items, cache, loading, failed }
}
