<template>
  <section class="gf-modal__section background-preferences" data-background-preferences>
    <div class="gf-modal__copy"><h3 class="gf-modal__label">{{ label('页面背景', 'Page background') }}</h3><p class="gf-modal__help">{{ label('预览后点击“保存”应用。本地图片只保存在此浏览器。', 'Preview, then save to apply. Local images stay in this browser.') }}</p></div>
    <div class="background-preferences__field">
      <span :id="`${fieldId}-source`">{{ label('背景来源', 'Source') }}</span>
      <div class="background-preferences__sources" role="group" :aria-labelledby="`${fieldId}-source`">
        <button v-for="source in (['default', 'server', 'local'] as const)" :key="source" type="button" :aria-pressed="draft.source === source" @click="changeSource(source)">
          {{ source === 'default' ? label('默认图案', 'Default') : source === 'server' ? label('服务端图案', 'Server patterns') : label('本地图片', 'Local image') }}
        </button>
      </div>
    </div>
    <div v-if="draft.source === 'server' && patterns.length" class="background-preferences__field">
      <span :id="`${fieldId}-patterns`">{{ label('选择图案', 'Choose a pattern') }}</span>
      <div class="background-preferences__carousel">
        <button type="button" class="background-preferences__arrow" :disabled="!canPrevious" :aria-label="label('上一组图案', 'Previous patterns')" @click="scrollPatterns(-1)"><PhCaretLeft :size="18" /></button>
        <div ref="strip" class="background-preferences__strip" role="group" :aria-labelledby="`${fieldId}-patterns`" @scroll="syncStrip">
          <BackgroundPatternOption v-for="pattern in patterns" :key="pattern.id" :pattern="pattern" :selected="draft.pattern_id === pattern.id" @select="selectPattern(pattern.id)" />
        </div>
        <button type="button" class="background-preferences__arrow" :disabled="!canNext" :aria-label="label('下一组图案', 'Next patterns')" @click="scrollPatterns(1)"><PhCaretRight :size="18" /></button>
      </div>
    </div>
    <p v-if="draft.source === 'server' && !patterns.length" class="gf-modal__help">{{ label('暂无可用图案，将使用默认背景。', 'No patterns available. The bundled background will be used.') }}</p>
    <div v-if="draft.source === 'local'" class="background-preferences__field"><label>{{ label('选择本地图片', 'Choose local image') }}<input class="gf-input" type="file" accept="image/svg+xml,image/png,image/jpeg,image/webp,image/avif,image/gif,image/bmp" @change="pickFile" /></label><p class="gf-modal__help">{{ local?.name || label('SVG 可调整颜色；照片等格式可调整透明度和尺寸。最多 10 MiB。', 'SVG supports color controls; raster images support opacity and size. Up to 10 MiB.') }}</p><button type="button" class="gf-button gf-button--ghost" @click="clearLocal">{{ label('清除本地背景', 'Clear local background') }}</button></div>
    <div class="background-preferences__controls">
      <div v-if="!raster" class="background-preferences__field background-preferences__color" @focusout="closeColor" @keydown.esc.stop="colorOpen = false">
        <label :for="`${fieldId}-color`">{{ label('图案颜色', 'Pattern color') }}</label>
        <div class="background-preferences__color-input">
          <button type="button" :aria-label="label('选择图案颜色', 'Choose pattern color')" :aria-expanded="colorOpen" :aria-controls="`${fieldId}-palette`" @click="colorOpen = !colorOpen"><span :style="{ backgroundColor: appearance.color }" /><PhCaretDown :size="12" /></button>
          <input :id="`${fieldId}-color`" :value="colorText" type="text" maxlength="7" spellcheck="false" autocomplete="off" @input="setColor(($event.target as HTMLInputElement).value)" />
        </div>
        <div v-if="colorOpen" :id="`${fieldId}-palette`" class="background-preferences__palette" role="group" :aria-label="label('常用颜色', 'Color presets')">
          <button v-for="color in colors" :key="color" type="button" :style="{ backgroundColor: color }" :aria-label="color" :aria-pressed="appearance.color.toLowerCase() === color" @click="setColor(color); colorOpen = false" />
        </div>
      </div>
      <label class="background-preferences__field"><span class="background-preferences__label-row">{{ label('透明度', 'Opacity') }}<output>{{ appearance.opacity }}</output></span><span class="background-preferences__slider-box"><input class="background-preferences__slider" type="range" min="0" max="1" step="0.001" :aria-label="label('透明度', 'Opacity')" :value="appearance.opacity" :style="{ '--range-progress': `${appearance.opacity * 100}%` }" @input="draft.overrides.opacity = ($event.target as HTMLInputElement).valueAsNumber" /></span></label>
      <label class="background-preferences__field"><span>{{ label('平铺尺寸', 'Tile size') }}</span><span class="background-preferences__size"><input class="gf-input" type="text" inputmode="numeric" :aria-label="label('平铺尺寸', 'Tile size')" :value="Number.isFinite(appearance.size) ? appearance.size : ''" @input="setSize(($event.target as HTMLInputElement).value)" /><span>px</span></span></label>
    </div>
    <div class="background-preferences__preview" aria-label="Background preview"><div :style="previewStyle" /></div>
    <div class="background-preferences__footer"><button type="button" class="gf-button gf-button--surface" @click="resetAppearance">{{ label('使用图案默认外观', 'Use pattern defaults') }}</button><span class="gf-modal__help">{{ label('仅保存你主动调整的参数。', 'Only your explicit overrides are saved.') }}</span></div>
    <p v-if="error" role="alert" class="background-preferences__error">{{ error }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, useId, watch } from 'vue'
import BackgroundPatternOption from './BackgroundPatternOption.vue'
import { PhCaretLeft, PhCaretRight, PhCaretDown } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '~/stores/theme'
import type { BackgroundPattern } from '~/types/nav'
import { DEFAULT_PATTERN_URL, defaultBackgroundPreference, patternAppearance, prepareLocalBackground, readBackgroundPreference, readLocalBackground, saveBackgroundPreference, type BackgroundPreference, type LocalBackground, type PatternDefaults } from '~/utils/backgroundPreferences'

const { locale } = useI18n()
const theme = useThemeStore()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const api = useApi('navV2')
const fieldId = useId()
const colorOpen = ref(false)
const colorText = ref('')
const colors = ['#9c846a', '#ac9680', '#7c2d12', '#c17c54', '#6b7c62', '#627e8f', '#8b7a96', '#333333', '#777777', '#aaaaaa', '#dddddd', '#ffffff']
const strip = ref<HTMLElement | null>(null)
const canPrevious = ref(false)
const canNext = ref(false)
let stripObserver: ResizeObserver | undefined
function syncStrip() {
  const element = strip.value
  canPrevious.value = Boolean(element && element.scrollLeft > 2)
  canNext.value = Boolean(element && element.scrollLeft + element.clientWidth < element.scrollWidth - 2)
}
watch(strip, element => {
  stripObserver?.disconnect()
  if (element) { stripObserver = new ResizeObserver(syncStrip); stripObserver.observe(element); syncStrip() }
})
function scrollPatterns(direction: number) {
  const element = strip.value
  if (element) element.scrollBy({ left: direction * element.clientWidth, behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth' })
}
function setColor(value: string) {
  error.value = ''
  colorText.value = value
  if (/^#[a-f\d]{6}$/i.test(value)) draft.value.overrides.color = value
}
function closeColor(event: FocusEvent) {
  if (!(event.currentTarget as HTMLElement).contains(event.relatedTarget as Node | null)) colorOpen.value = false
}
function setSize(value: string) { error.value = ''; draft.value.overrides.size_px = /^\d+$/.test(value) ? Number(value) : NaN }
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
watch(() => appearance.value.color, value => { colorText.value = value }, { immediate: true })
const server = useManagedAsset(() => selected.value?.object_key, DEFAULT_PATTERN_URL, () => Boolean(selected.value))
const previewStyle = computed(() => {
  const image = draft.value.source === 'local' ? localURL.value || DEFAULT_PATTERN_URL : selected.value ? server.src.value : DEFAULT_PATTERN_URL
  const url = image ? `url("${image}")` : 'none'
  return { position: 'absolute' as const, inset: '0', opacity: appearance.value.opacity, backgroundColor: raster.value ? 'transparent' : appearance.value.color, maskImage: raster.value ? 'none' : url, WebkitMaskImage: raster.value ? 'none' : url, maskRepeat: 'repeat', WebkitMaskRepeat: 'repeat', maskSize: `${appearance.value.size}px`, WebkitMaskSize: `${appearance.value.size}px`, backgroundImage: raster.value ? url : 'none', backgroundSize: `${appearance.value.size}px`, backgroundRepeat: 'repeat' }
})
function setLocal(value: LocalBackground | null) { if (localURL.value) URL.revokeObjectURL(localURL.value); local.value = value; localURL.value = value ? URL.createObjectURL(value.blob) : '' }
function changeSource(value: string) { error.value = ''; if (value === 'default' || value === 'server' || value === 'local') draft.value = { version: 1, source: value, overrides: {}, ...(value === 'server' && patterns.value[0] ? { pattern_id: patterns.value[0].id } : {}) } }
function selectPattern(id: string) {
  error.value = ''
  draft.value = { version: 1, source: 'server', pattern_id: id, overrides: {} }
  nextTick(() => strip.value?.querySelector<HTMLElement>(`[data-pattern-id="${id}"]`)?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'instant' }))
}
function clearLocal() { selectionVersion++; busy.value = false; setLocal(null); pendingFile.value = null; removeLocal.value = true; draft.value = defaultBackgroundPreference() }
async function pickFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]; if (!file) return
  const version = ++selectionVersion; busy.value = true; error.value = ''
  try { const prepared = await prepareLocalBackground(file); if (version === selectionVersion) { setLocal(prepared); pendingFile.value = prepared; removeLocal.value = false; draft.value = { version: 1, source: 'local', overrides: {} } } } catch (err) { error.value = err instanceof Error ? err.message : label('无法读取图片', 'Cannot read image') } finally { if (version === selectionVersion) busy.value = false }
}
function resetAppearance() { draft.value.overrides = {}; colorText.value = appearance.value.color; error.value = '' }
async function save() {
  if (!raster.value && !/^#[a-f\d]{6}$/i.test(colorText.value)) { error.value = label('请输入有效的六位十六进制颜色，例如 #9c846a', 'Enter a six-digit hex color, such as #9c846a'); return false }
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
onUnmounted(() => { stripObserver?.disconnect(); disposed = true; selectionVersion++; if (localURL.value) URL.revokeObjectURL(localURL.value) })
</script>

<style scoped>
.background-preferences { display: grid; gap: 1.1rem; }
.background-preferences__field { display: grid; min-width: 0; gap: .5rem; font-size: .82rem; line-height: 1.4; color: var(--gf-text-main); }
.background-preferences__sources { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .4rem; }
.background-preferences__sources button { padding: .65rem .3rem; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); background: transparent; color: var(--gf-text-muted); cursor: pointer; transition: background 160ms, border-color 160ms; }
.background-preferences__sources button:hover { background: var(--gf-surface-hover); }
.background-preferences__sources button[aria-pressed='true'] { border-color: var(--gf-accent); background: var(--gf-accent-soft); color: var(--gf-accent); }
.background-preferences__carousel { display: grid; grid-template-columns: 1.5rem minmax(0, 1fr) 1.5rem; gap: .4rem; align-items: center; }
.background-preferences__strip { display: flex; gap: .6rem; overflow-x: auto; scroll-snap-type: x mandatory; scrollbar-width: none; overscroll-behavior-x: contain; padding: 3px; }
.background-preferences__strip::-webkit-scrollbar { display: none; }
.background-preferences__arrow { display: grid; place-items: center; height: 2.5rem; border: 0; border-radius: 6px; background: transparent; color: var(--gf-text-main); cursor: pointer; }
.background-preferences__arrow:hover:not(:disabled) { background: var(--gf-accent-soft); color: var(--gf-accent); }
.background-preferences__arrow:disabled { opacity: .25; cursor: default; }
.background-preferences__controls { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .85rem; align-items: start; }
.background-preferences__controls:has(> :nth-child(2):last-child) { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.background-preferences__label-row { display: flex; align-items: baseline; justify-content: space-between; gap: .25rem; }
.background-preferences__label-row output { color: var(--gf-text-muted); font-size: .75rem; font-variant-numeric: tabular-nums; }
.background-preferences__color { position: relative; }
.background-preferences__color-input { display: flex; align-items: center; height: 42px; min-width: 0; border: 1px solid var(--gf-border-strong); border-radius: var(--gf-radius-sm); background: var(--gf-input-bg); }
.background-preferences__color-input button { display: flex; flex-shrink: 0; align-items: center; gap: 2px; padding: 8px 4px 8px 8px; border: 0; background: transparent; color: var(--gf-text-muted); cursor: pointer; }
.background-preferences__color-input button span { width: 19px; height: 19px; border-radius: 4px; box-shadow: inset 0 0 0 1px rgb(0 0 0 / .1); }
.background-preferences__color-input input { width: 100%; min-width: 0; padding: 0 5px; border: 0; background: transparent; color: var(--gf-text-main); font-size: .74rem; font-variant-numeric: tabular-nums; outline: none; }
.background-preferences__color-input:focus-within { outline: 2px solid var(--gf-focus-ring); }
.background-preferences__palette { position: absolute; z-index: 5; top: calc(100% + 6px); left: 0; display: grid; grid-template-columns: repeat(6, 24px); gap: 8px; padding: 12px; border: 1px solid var(--gf-border-strong); border-radius: var(--gf-radius-sm); background: var(--gf-modal-bg); box-shadow: var(--gf-shadow-soft); }
.background-preferences__palette button { width: 24px; height: 24px; border: 1px solid var(--gf-border-strong); border-radius: 50%; cursor: pointer; }
.background-preferences__palette button:hover, .background-preferences__palette button[aria-pressed='true'] { outline: 2px solid var(--gf-accent); outline-offset: 2px; }
.background-preferences__slider-box { display: flex; align-items: center; height: 42px; padding: 0 5px; }
.background-preferences__slider { appearance: none; width: 100%; height: 5px; margin: 0; border-radius: 10px; background: linear-gradient(to right, var(--gf-accent) var(--range-progress), var(--gf-border-strong) var(--range-progress)); cursor: pointer; }
.background-preferences__slider::-webkit-slider-thumb { appearance: none; width: 15px; height: 15px; border: 3px solid var(--gf-modal-bg); border-radius: 50%; background: var(--gf-accent); box-shadow: 0 0 0 1px var(--gf-accent); }
.background-preferences__slider::-moz-range-thumb { width: 10px; height: 10px; border: 3px solid var(--gf-modal-bg); border-radius: 50%; background: var(--gf-accent); box-shadow: 0 0 0 1px var(--gf-accent); }
.background-preferences__slider:hover { filter: brightness(1.15); }
.background-preferences__size { position: relative; height: 42px; }
.background-preferences__size .gf-input { width: 100%; height: 42px; padding: 0 30px 0 12px; font-size: .85rem; }
.background-preferences__size > span { position: absolute; right: 10px; top: 50%; transform: translateY(-50%); font-size: .72rem; color: var(--gf-text-muted); pointer-events: none; }
.background-preferences__preview { position: relative; height: 130px; overflow: hidden; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); background: var(--gf-page-background); }
.background-preferences__footer { display: flex; flex-wrap: wrap; gap: .5rem; align-items: flex-end; }
.background-preferences__footer .gf-modal__help { padding-bottom: .15rem; }
.background-preferences__error { color: var(--gf-danger); font-size: .8rem; }
.background-preferences button:focus-visible, .background-preferences__slider:focus-visible { outline: 2px solid var(--gf-accent); outline-offset: 3px; }
@media (max-width: 400px) { .background-preferences__controls { gap: .55rem; } .background-preferences__color-input button { padding: 4px; } .background-preferences__color-input button svg { display: none; } .background-preferences__color-input button span { width: 15px; height: 15px; } .background-preferences__color-input input { font-size: .68rem; } }
@media (prefers-reduced-motion: reduce) { .background-preferences__sources button { transition: none; } }
</style>
