<template>
  <div class="hero-catalog" :data-hero-catalog="variant" :aria-busy="busy">
    <p v-if="error" role="alert" class="gf-modal__help">{{ error }} <button type="button" class="gf-button gf-button--ghost" @click="retry">{{ label('重试', 'Retry') }}</button></p>
    <p v-if="unavailable" role="status" class="gf-modal__help">{{ label('已保存的背景当前不可用，可浏览并保存其他背景。', 'The saved background is unavailable. Browse and save another background.') }}</p>
    <div class="preferences-carousel" @keydown.left.prevent="move(-1)" @keydown.right.prevent="move(1)">
      <button type="button" class="preferences-arrow" :disabled="busy || position <= 1" :aria-label="label('上一张背景', 'Previous background')" @click="move(-1)"><PhCaretLeft :size="18" /></button>
      <div ref="previewBox" class="hero-catalog__preview">
        <ManagedAssetImage v-if="visible && active && preview" :key="preview.object_key" :object-key="preview.object_key" :alt="preview.name" fallback="" />
        <p v-else-if="!preview" class="gf-modal__help">{{ busy ? label('加载中…', 'Loading…') : label('暂无可用背景', 'No backgrounds available') }}</p>
      </div>
      <button type="button" class="preferences-arrow" :disabled="busy || !preview || position >= (catalog?.total ?? 0)" :aria-label="label('下一张背景', 'Next background')" @click="move(1)"><PhCaretRight :size="18" /></button>
    </div>
    <div class="hero-catalog__caption gf-modal__help" aria-live="polite" aria-atomic="true">
      <span :title="preview?.name">{{ preview?.name }}</span>
      <span data-hero-position>{{ position }} / {{ catalog?.total ?? 0 }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { PhCaretLeft, PhCaretRight } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import ManagedAssetImage from './ManagedAssetImage.vue'
import type { HeroCatalog } from '~/types/nav'
const props = defineProps<{ variant: 'desktop' | 'mobile'; selectedId: string | null; active: boolean }>()
const emit = defineEmits<{ select: [id: string] }>()
const { locale } = useI18n()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const api = useApi('navV2')
const busy = ref(false), error = ref(''), unavailable = ref(false)
const catalog = ref<HeroCatalog | null>(null), index = ref(0)
const preview = computed(() => catalog.value?.items[index.value] ?? null)
const position = computed(() => preview.value && catalog.value ? (catalog.value.page_num - 1) * catalog.value.page_size + index.value + 1 : 0)
const pages = new Map<number, HeroCatalog>()
const previewBox = ref<HTMLElement | null>(null), visible = ref(false)
let observer: IntersectionObserver | undefined, version = 0, initialized = false
let retryAction = initialize
function retry() { void retryAction() }
async function readPage(page: number) {
  const cached = pages.get(page)
  if (cached) return cached
  const result = await api<HeroCatalog>('/nav/appearance/heroes', { query: { variant: props.variant, page_num: page, page_size: 12, selected_id: props.selectedId ?? undefined } })
  pages.set(page, result)
  return result
}
function showPage(page: HeroCatalog, nextIndex: number) {
  catalog.value = page
  index.value = nextIndex
  if (props.active && preview.value) emit('select', preview.value.id)
}
async function initialize() {
  if (busy.value) return
  const current = ++version, selectedId = props.selectedId
  busy.value = true; error.value = ''; retryAction = initialize
  try {
    const first = await readPage(1)
    let selectedPage = first, selectedIndex = first.items.findIndex(item => item.id === selectedId)
    unavailable.value = Boolean(selectedId && !first.selected)
    // selected_id resolves metadata but has no ordinal. Locate an out-of-page
    // selection using the API's stable ID order, without walking every page or
    // mounting any image from those metadata-only reads. Never coerce bigint IDs.
    if (selectedId && first.selected && selectedIndex < 0) {
      let low = 2, high = Math.ceil(first.total / first.page_size)
      while (low <= high && current === version) {
        const page = Math.floor((low + high) / 2)
        const candidate = await readPage(page)
        const found = candidate.items.findIndex(item => item.id === selectedId)
        if (found >= 0) { selectedPage = candidate; selectedIndex = found; break }
        const last = candidate.items.at(-1)
        if (last && BigInt(last.id) < BigInt(selectedId)) low = page + 1
        else high = page - 1
      }
    }
    if (current !== version) return
    initialized = true
    showPage(selectedPage, Math.max(0, selectedIndex))
  } catch { if (current === version) error.value = label('云端背景目录暂不可用，已保存设置不受影响。', 'The cloud catalog is unavailable. Your saved selection is unchanged.') }
  finally { if (current === version) busy.value = false }
}
async function move(direction: -1 | 1) {
  const page = catalog.value
  if (busy.value || !page || !preview.value || (direction < 0 ? position.value <= 1 : position.value >= page.total)) return
  error.value = ''
  const nextIndex = index.value + direction
  if (nextIndex >= 0 && nextIndex < page.items.length) { showPage(page, nextIndex); return }
  const current = ++version
  busy.value = true; retryAction = () => move(direction)
  try {
    const next = await readPage(page.page_num + direction)
    if (current === version && next.items.length) showPage(next, direction > 0 ? 0 : next.items.length - 1)
  } catch { if (current === version) error.value = label('无法加载下一组背景，请重试。', 'Cannot load more backgrounds. Please retry.') }
  finally { if (current === version) busy.value = false }
}
watch(() => props.active, active => {
  if (!active) return
  if (!initialized) void initialize()
  else if (preview.value) emit('select', preview.value.id)
})
onMounted(() => {
  observer = new IntersectionObserver(entries => { visible.value = Boolean(entries[0]?.isIntersecting) })
  if (previewBox.value) observer.observe(previewBox.value)
  if (props.active) void initialize()
})
onUnmounted(() => { version++; observer?.disconnect() })
</script>

<style scoped>
.hero-catalog { display: grid; gap: .5rem; min-width: 0; }
.hero-catalog__preview { display: grid; place-items: center; height: clamp(9rem, 28vw, 12rem); overflow: hidden; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); background: var(--gf-page-background); }
.hero-catalog__preview img { width: 100%; height: 100%; min-height: 0; object-fit: contain; }
.hero-catalog__caption { display: flex; justify-content: space-between; gap: .75rem; padding: 0 1.9rem; }
.hero-catalog__caption > :first-child { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.hero-catalog__caption > :last-child { flex-shrink: 0; font-variant-numeric: tabular-nums; }
</style>
