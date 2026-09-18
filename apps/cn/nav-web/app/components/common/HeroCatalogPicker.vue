<template>
  <div class="hero-catalog" :data-hero-catalog="variant">
    <p v-if="error" role="alert" class="gf-modal__help">{{ error }} <button type="button" class="gf-button gf-button--ghost" @click="load">{{ label('重试', 'Retry') }}</button></p>
    <p v-if="selectedId && catalog && !catalog.selected" role="status" class="gf-modal__help">{{ label('已保存的背景当前不可用，此设备将回退到云端随机。', 'The saved background is unavailable. This viewport will use a random cloud background.') }}</p>
    <div class="hero-catalog__controls">
      <button type="button" class="gf-button gf-button--ghost" :disabled="busy || page === 1" @click="page--">{{ label('上一页', 'Previous') }}</button>
      <span>{{ page }} / {{ Math.max(1, Math.ceil((catalog?.total ?? 0) / 12)) }}</span>
      <button type="button" class="gf-button gf-button--ghost" :disabled="busy || page * 12 >= (catalog?.total ?? 0)" @click="page++">{{ label('下一页', 'Next') }}</button>
    </div>
    <label class="gf-modal__label">{{ label('预览背景', 'Preview background') }}
      <select class="gf-input" :value="preview?.id ?? ''" :disabled="busy || !choices.length" @change="preview = choices.find(item => item.id === ($event.target as HTMLSelectElement).value) ?? null">
        <option v-if="!choices.length" value="">{{ label('暂无可用背景', 'No backgrounds available') }}</option>
        <option v-for="item in choices" :key="item.id" :value="item.id">{{ item.name }}{{ item.id === selectedId ? label('（已选）', ' (selected)') : '' }}</option>
      </select>
    </label>
    <div ref="previewBox" class="hero-catalog__preview" :class="{ 'hero-catalog__preview--mobile': variant === 'mobile' }">
      <ManagedAssetImage v-if="visible && active && preview" :key="preview.object_key" :object-key="preview.object_key" :alt="preview.name" fallback="" />
    </div>
    <div class="gf-modal__actions">
      <button type="button" class="gf-button gf-button--surface" :disabled="!preview || busy" :aria-pressed="Boolean(preview && preview.id === selectedId)" @click="choose(preview!)">{{ label('使用这张背景', 'Use this background') }}</button>
      <button type="button" class="gf-button gf-button--ghost" :aria-pressed="!selectedId" @click="emit('select', null)">{{ label('此设备随机', 'Random for this viewport') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import ManagedAssetImage from './ManagedAssetImage.vue'
import type { HeroCatalog, HeroCatalogItem } from '~/types/nav'
const props = defineProps<{ variant: 'desktop' | 'mobile'; selectedId: string | null; active: boolean }>()
const emit = defineEmits<{ select: [id: string | null] }>()
const { locale } = useI18n()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const api = useApi('navV2')
const page = ref(1), busy = ref(false), error = ref('')
const catalog = ref<HeroCatalog | null>(null), preview = ref<HeroCatalogItem | null>(null)
const choices = computed(() => {
  const data = catalog.value
  return data ? data.selected && !data.items.some(item => item.id === data.selected!.id) ? [data.selected, ...data.items] : data.items : []
})
const previewBox = ref<HTMLElement | null>(null), visible = ref(false)
let observer: IntersectionObserver | undefined, version = 0
async function load() {
  const current = ++version
  busy.value = true; error.value = ''
  try {
    const result = await api<HeroCatalog>('/nav/appearance/heroes', { query: { variant: props.variant, page_num: page.value, page_size: 12, selected_id: props.selectedId ?? undefined } })
    if (current !== version) return
    catalog.value = result
    preview.value = (page.value === 1 ? result.selected : null) ?? result.items[0] ?? null
  } catch { if (current === version) error.value = label('云端背景目录暂不可用，已保存设置不受影响。', 'The cloud catalog is unavailable. Your saved selection is unchanged.') }
  finally { if (current === version) busy.value = false }
}
function choose(item: HeroCatalogItem) {
  if (catalog.value) catalog.value.selected = item
  emit('select', item.id)
}
watch(page, load)
watch(() => props.active, active => { if (active && !catalog.value && !busy.value) void load() })
onMounted(() => {
  observer = new IntersectionObserver(entries => { visible.value = Boolean(entries[0]?.isIntersecting) })
  if (previewBox.value) observer.observe(previewBox.value)
  if (props.active) void load()
})
onUnmounted(() => { version++; observer?.disconnect() })
</script>

<style scoped>
.hero-catalog { display: grid; gap: .8rem; min-width: 0; }
.hero-catalog__controls { display: flex; align-items: center; justify-content: space-between; color: var(--gf-text-muted); font-size: .85rem; }
.hero-catalog__preview { height: 11rem; border-radius: .6rem; overflow: hidden; background: var(--gf-surface-muted); }
.hero-catalog__preview img { width: 100%; height: 100%; object-fit: contain; }
.hero-catalog__preview--mobile { height: 13rem; }
.hero-catalog select { display: block; width: 100%; margin-top: .5rem; }
</style>
