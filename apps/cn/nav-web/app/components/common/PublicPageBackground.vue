<template>
  <div class="gf-public-background" data-public-background :data-pattern-status="effectiveSource" aria-hidden="true">
    <div class="gf-public-background__pattern" :style="patternStyle" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { BackgroundPattern } from '~/types/nav'
import { BACKGROUND_CHANGE_EVENT, BACKGROUND_PREFERENCE_KEY, DEFAULT_PATTERN_URL, defaultBackgroundPreference, readBackgroundPreference, readLocalBackground, type LocalBackground } from '~/utils/backgroundPreferences'
import { useThemeStore } from '~/stores/theme'

const preference = ref(defaultBackgroundPreference())
const patterns = ref<BackgroundPattern[]>([])
const local = ref<LocalBackground | null>(null)
const localURL = ref('')
const theme = useThemeStore()
const api = useApi('navV2')
let revision = 0
const selected = computed(() => preference.value.source === 'server' ? patterns.value.find((pattern) => pattern.id === preference.value.pattern_id) : undefined)
const server = useManagedAsset(() => selected.value?.object_key, DEFAULT_PATTERN_URL, () => Boolean(selected.value))
const effectiveSource = computed(() => preference.value.source === 'server' ? (selected.value && server.src.value !== DEFAULT_PATTERN_URL ? 'server' : 'default') : preference.value.source === 'local' ? (local.value ? 'local' : 'default') : 'default')
const patternStyle = computed(() => {
  const source = effectiveSource.value
  const overrides = source === preference.value.source ? preference.value.overrides : {}
  const dark = theme.theme === 'dark'
  const raster = source === 'local' && local.value?.kind === 'raster'
  const image = source === 'local' ? localURL.value : source === 'server' ? server.src.value : DEFAULT_PATTERN_URL
  const color = overrides.color ?? (source === 'server' ? selected.value![dark ? 'dark_color' : 'light_color'] : 'var(--gf-page-pattern-color)')
  const opacity = overrides.opacity ?? (source === 'server' ? selected.value![dark ? 'dark_opacity' : 'light_opacity'] : 'var(--gf-page-pattern-opacity)')
  const size = overrides.size_px ? `${overrides.size_px}px` : source === 'server' ? `${selected.value!.default_size_px}px` : 'var(--gf-page-pattern-size)'
  const imageValue = image ? `url("${image}")` : 'none'
  return { backgroundColor: raster ? 'transparent' : color, opacity, maskImage: raster ? 'none' : imageValue, WebkitMaskImage: raster ? 'none' : imageValue, maskSize: size, WebkitMaskSize: size, backgroundImage: raster ? imageValue : 'none', backgroundRepeat: 'repeat', backgroundSize: size }
})

function clearLocalURL() { if (localURL.value) URL.revokeObjectURL(localURL.value); localURL.value = '' }
async function reload() {
  const current = ++revision
  const next = readBackgroundPreference()
  preference.value = next
  try {
    if (next.source === 'server') {
      const result = await api<{ patterns: BackgroundPattern[] }>('/nav/appearance/patterns')
      if (current === revision) patterns.value = result.patterns
    } else if (next.source === 'local') {
      const result = await readLocalBackground()
      if (current === revision) { clearLocalURL(); local.value = result; if (result) localURL.value = URL.createObjectURL(result.blob) }
    }
  } catch { if (current === revision) { patterns.value = []; local.value = null; clearLocalURL() } }
  if (current === revision && next.source !== 'local') { local.value = null; clearLocalURL() }
}
const storageChanged = (event: StorageEvent) => { if (event.key === BACKGROUND_PREFERENCE_KEY || event.key === null) void reload() }
onMounted(() => { void reload(); window.addEventListener(BACKGROUND_CHANGE_EVENT, reload); window.addEventListener('storage', storageChanged); window.addEventListener('focus', reload) })
onUnmounted(() => { revision++; window.removeEventListener(BACKGROUND_CHANGE_EVENT, reload); window.removeEventListener('storage', storageChanged); window.removeEventListener('focus', reload); clearLocalURL() })
</script>

<style scoped>
.gf-public-background {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: var(--gf-page-background);
}

.gf-public-background__pattern {
  position: absolute;
  inset: 0;
  background-color: var(--gf-page-pattern-color);
  -webkit-mask-image: var(--gf-page-pattern);
  mask-image: var(--gf-page-pattern);
  -webkit-mask-position: top left;
  mask-position: top left;
  -webkit-mask-repeat: repeat;
  mask-repeat: repeat;
  -webkit-mask-size: var(--gf-page-pattern-size);
  mask-size: var(--gf-page-pattern-size);
  opacity: var(--gf-page-pattern-opacity);
}
</style>
