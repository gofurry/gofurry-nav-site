<template>
  <section class="gf-modal__section background-preferences" data-background-preferences>
    <div class="gf-modal__copy"><h3 class="gf-modal__label">{{ label('页面背景', 'Page background') }}</h3><p class="gf-modal__help">{{ label('预览后点击偏好设置的“保存”应用。本地图片只保存在此浏览器。', 'Preview, then save preferences to apply. Local images stay in this browser.') }}</p></div>
    <label class="background-preferences__field">{{ label('背景来源', 'Source') }}<select class="gf-input" :value="draft.source" @change="changeSource(($event.target as HTMLSelectElement).value)"><option value="default">{{ label('默认图案', 'Bundled default') }}</option><option value="server">{{ label('服务端图案', 'Server pattern') }}</option><option value="local">{{ label('本地图片', 'Local image') }}</option></select></label>
    <label v-if="draft.source === 'server'" class="background-preferences__field">{{ label('选择图案', 'Pattern') }}<select class="gf-input" :value="draft.pattern_id || ''" @change="selectPattern(($event.target as HTMLSelectElement).value)"><option value="">{{ label('请选择', 'Select a pattern') }}</option><option v-for="pattern in patterns" :key="pattern.id" :value="pattern.id">{{ locale === 'en' ? pattern.name_en : pattern.name }}</option></select></label>
    <p v-if="draft.source === 'server' && !patterns.length" class="gf-modal__help">{{ label('暂无可用图案，将使用默认背景。', 'No patterns available. The bundled background will be used.') }}</p>
    <div v-if="draft.source === 'local'" class="background-preferences__field"><label>{{ label('选择本地图片', 'Choose local image') }}<input class="gf-input" type="file" accept="image/svg+xml,image/png,image/jpeg,image/webp,image/avif,image/gif,image/bmp" @change="pickFile" /></label><p class="gf-modal__help">{{ local?.name || label('SVG 可调整颜色；照片等格式可调整透明度和尺寸。最多 10 MiB。', 'SVG supports color controls; raster images support opacity and size. Up to 10 MiB.') }}</p><button type="button" class="gf-button gf-button--ghost" @click="clearLocal">{{ label('清除本地背景', 'Clear local background') }}</button></div>
    <div class="background-preferences__controls">
      <label v-if="!raster" class="background-preferences__field">{{ label('图案颜色', 'Pattern color') }}<input type="color" :value="appearance.color" @input="draft.overrides.color = ($event.target as HTMLInputElement).value" /></label>
      <label class="background-preferences__field">{{ label('透明度', 'Opacity') }} · {{ appearance.opacity }}<input type="range" min="0" max="1" step="0.01" :value="appearance.opacity" @input="draft.overrides.opacity = ($event.target as HTMLInputElement).valueAsNumber" /></label>
      <label class="background-preferences__field">{{ label('平铺尺寸（px）', 'Tile size (px)') }}<input class="gf-input" type="number" min="1" max="10000" :value="appearance.size" @input="draft.overrides.size_px = ($event.target as HTMLInputElement).valueAsNumber" /></label>
    </div>
    <div class="background-preferences__preview" aria-label="Background preview"><div :style="previewStyle" /></div>
    <div class="gf-modal__actions"><button type="button" class="gf-button gf-button--surface" @click="draft.overrides = {}">{{ label('使用图案默认外观', 'Use pattern defaults') }}</button><span class="gf-modal__help">{{ label('仅保存你主动调整的参数。', 'Only your explicit overrides are saved.') }}</span></div>
    <p v-if="error" role="alert" class="background-preferences__error">{{ error }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '~/stores/theme'
import type { BackgroundPattern } from '~/types/nav'
import { DEFAULT_PATTERN_URL, defaultBackgroundPreference, patternAppearance, prepareLocalBackground, readBackgroundPreference, readLocalBackground, saveBackgroundPreference, type BackgroundPreference, type LocalBackground, type PatternDefaults } from '~/utils/backgroundPreferences'

const { locale } = useI18n()
const theme = useThemeStore()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const api = useApi('navV2')
const draft = ref<BackgroundPreference>(defaultBackgroundPreference())
const patterns = ref<BackgroundPattern[]>([])
const local = ref<LocalBackground | null>(null)
const localURL = ref('')
const pendingFile = ref<LocalBackground | null>(null)
const removeLocal = ref(false)
const busy = ref(true)
const error = ref('')
const defaults = ref<PatternDefaults>({ light_color: '#000000', dark_color: '#000000', light_opacity: 0, dark_opacity: 0, default_size_px: 1 })
let selectionVersion = 0
let disposed = false
function syncDefaults() {
  const css = getComputedStyle(document.documentElement)
  const color = css.getPropertyValue('--gf-page-pattern-color').trim()
  const opacity = Number(css.getPropertyValue('--gf-page-pattern-opacity'))
  const size = Number.parseFloat(css.getPropertyValue('--gf-page-pattern-size'))
  defaults.value = { light_color: color, dark_color: color, light_opacity: opacity, dark_opacity: opacity, default_size_px: size }
}
watch(() => theme.theme, () => { if (import.meta.client) syncDefaults() }, { flush: 'post' })
const selected = computed(() => draft.value.source === 'server' ? patterns.value.find((pattern) => pattern.id === draft.value.pattern_id) : undefined)
const raster = computed(() => draft.value.source === 'local' && local.value?.kind === 'raster')
const appearance = computed(() => patternAppearance(selected.value ?? defaults.value, theme.theme, draft.value.overrides))
const server = useManagedAsset(() => selected.value?.object_key, DEFAULT_PATTERN_URL, () => Boolean(selected.value))
const previewStyle = computed(() => {
  const image = draft.value.source === 'local' ? localURL.value || DEFAULT_PATTERN_URL : selected.value ? server.src.value : DEFAULT_PATTERN_URL
  const url = image ? `url("${image}")` : 'none'
  return { position: 'absolute' as const, inset: '0', opacity: appearance.value.opacity, backgroundColor: raster.value ? 'transparent' : appearance.value.color, maskImage: raster.value ? 'none' : url, WebkitMaskImage: raster.value ? 'none' : url, maskRepeat: 'repeat', WebkitMaskRepeat: 'repeat', maskSize: `${appearance.value.size}px`, WebkitMaskSize: `${appearance.value.size}px`, backgroundImage: raster.value ? url : 'none', backgroundSize: `${appearance.value.size}px`, backgroundRepeat: 'repeat' }
})
function setLocal(value: LocalBackground | null) { if (localURL.value) URL.revokeObjectURL(localURL.value); local.value = value; localURL.value = value ? URL.createObjectURL(value.blob) : '' }
function changeSource(value: string) { if (value === 'default' || value === 'server' || value === 'local') draft.value = { version: 1, source: value, overrides: {}, ...(value === 'server' && patterns.value[0] ? { pattern_id: patterns.value[0].id } : {}) } }
function selectPattern(id: string) { draft.value = { version: 1, source: 'server', pattern_id: id, overrides: {} } }
function clearLocal() { selectionVersion++; busy.value = false; setLocal(null); pendingFile.value = null; removeLocal.value = true; draft.value = defaultBackgroundPreference() }
async function pickFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]; if (!file) return
  const version = ++selectionVersion; busy.value = true; error.value = ''
  try { const prepared = await prepareLocalBackground(file); if (version === selectionVersion) { setLocal(prepared); pendingFile.value = prepared; removeLocal.value = false; draft.value = { version: 1, source: 'local', overrides: {} } } } catch (err) { error.value = err instanceof Error ? err.message : label('无法读取图片', 'Cannot read image') } finally { if (version === selectionVersion) busy.value = false }
}
async function save() {
  if (busy.value) { error.value = label('请等待图片读取完成', 'Please wait for the image to load'); return false }
  if (draft.value.source === 'local' && !local.value) { error.value = label('请先选择本地图片', 'Choose a local image first'); return false }
  if (draft.value.source === 'server' && !selected.value) draft.value = defaultBackgroundPreference()
  if (draft.value.overrides.size_px !== undefined && (!Number.isInteger(draft.value.overrides.size_px) || draft.value.overrides.size_px < 1 || draft.value.overrides.size_px > 10000)) { error.value = label('尺寸必须为 1–10000 的整数', 'Size must be an integer from 1 to 10000'); return false }
  try { await saveBackgroundPreference(draft.value, pendingFile.value, removeLocal.value); return true } catch (err) { error.value = err instanceof Error ? err.message : label('无法保存背景', 'Cannot save background'); return false }
}
defineExpose({ save })
onMounted(async () => {
  syncDefaults()
  const version = ++selectionVersion
  draft.value = readBackgroundPreference()
  try { const stored = await readLocalBackground(); if (version === selectionVersion) setLocal(stored) } catch { /* Local storage may be unavailable; default/server still work. */ }
  if (version === selectionVersion) busy.value = false
  if (disposed) return
  try { const result = await api<{ patterns: BackgroundPattern[] }>('/nav/appearance/patterns'); if (!disposed) patterns.value = result.patterns } catch { if (!disposed) error.value = label('服务端图案暂不可用，默认和本地背景仍可使用。', 'Server patterns are temporarily unavailable. Default and local backgrounds remain available.') }
})
onUnmounted(() => { disposed = true; selectionVersion++; if (localURL.value) URL.revokeObjectURL(localURL.value) })
</script>

<style scoped>
.background-preferences { display: grid; gap: 1rem; }
.background-preferences__field { display: grid; gap: .45rem; font-size: .85rem; color: var(--gf-text-main); }
.background-preferences__controls { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap: 1rem; align-items: end; }
.background-preferences__preview { position: relative; height: 140px; overflow: hidden; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); background: var(--gf-page-background); }
.background-preferences__error { color: var(--gf-danger); font-size: .8rem; }
</style>
