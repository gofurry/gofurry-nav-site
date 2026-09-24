<template>
  <HeroBackgroundFrame v-for="frame in frames" :key="frame.id" :desktop-key="frame.desktopKey" :mobile-key="frame.mobileKey" :local-src="frame.localURL"
    :ref="element => rememberFrame(frame.id, element)" :previous="frame.previous"
    :data-hero-pending="frame.id !== displayedId || undefined" :style="frame.id !== displayedId ? { visibility: 'hidden' } : undefined"
    @ready="promote(frame)" @failed="failed(frame)" />
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch, type ComponentPublicInstance } from 'vue'
import HeroBackgroundFrame from './HeroBackgroundFrame.vue'
import type { NavHomeHero } from '~/types/nav'
import { heroQuery } from '~/utils/heroPreferences'
import { loadRandomCustomNavHeaderBackground } from '~/utils/customNavHeaderBackground'

const props = defineProps<{ desktopObjectKey?: string | null; mobileObjectKey?: string | null }>()
const settings = useHeroPreferences()
const api = useApi('navV2')
type FrameInstance = InstanceType<typeof HeroBackgroundFrame>
interface Frame { id: number; desktopKey: string | null; mobileKey: string | null; localURL: string | null; previous?: ReturnType<FrameInstance['snapshot']> }
const frameInstances = new Map<number, FrameInstance>()
function rememberFrame(id: number, element: Element | ComponentPublicInstance | null) {
  if (element) frameInstances.set(id, element as FrameInstance)
  else frameInstances.delete(id)
}
let sequence = 0, request = 0, disposed = false, migrating = false
const initialLocal = settings.preference.value.mode === 'local'
const frames = ref<Frame[]>([{ id: sequence++, desktopKey: initialLocal ? null : props.desktopObjectKey ?? null, mobileKey: initialLocal ? null : props.mobileObjectKey ?? null, localURL: null }])
const displayedId = ref(0)
function release(frame: Frame) { if (frame.localURL) URL.revokeObjectURL(frame.localURL) }
function promote(frame: Frame) {
  if (!frames.value.some(item => item.id === frame.id) || frame.id === displayedId.value) return
  const old = frames.value.filter(item => item.id !== frame.id)
  displayedId.value = frame.id
  frames.value = [frame]
  old.forEach(release)
}
function stage(desktopKey: string | null, mobileKey: string | null, localURL: string | null = null) {
  const current = discardPending()
  if (current.desktopKey === desktopKey && current.mobileKey === mobileKey && current.localURL === localURL) return
  // Keep the old renderer while the new, real DOM image loads. Promote the
  // same element on load; do not replace it with another image after preload.
  frames.value.push({ id: sequence++, desktopKey, mobileKey, localURL, previous: frameInstances.get(current.id)?.snapshot() })
}
function discardPending() {
  const current = frames.value.find(frame => frame.id === displayedId.value)!
  frames.value.filter(frame => frame !== current).forEach(release)
  frames.value = [current]
  return current
}
function failed(frame: Frame) {
  if (!frames.value.some(item => item.id === frame.id)) return
  if (frame.id !== displayedId.value) {
    frames.value = frames.value.filter(item => item.id !== frame.id)
    release(frame)
  }
  if (frame.localURL && settings.preference.value.mode === 'local') {
    settings.save({ ...settings.preference.value, mode: 'random' })
  }
}
async function apply() {
  const version = ++request
  discardPending()
  const preference = { ...settings.preference.value }
  try {
    if (preference.mode === 'local') {
      const url = await loadRandomCustomNavHeaderBackground()
      if (disposed || version !== request) { if (url) URL.revokeObjectURL(url); return }
      if (url) stage(null, null, url)
      else settings.save({ ...preference, mode: 'random' }) // one focused retry via revision
      return
    }
    const response = await api<{ hero: NavHomeHero }>('/nav/home/hero', { query: heroQuery(preference) })
    if (!disposed && version === request) stage(response.hero.desktop?.object_key ?? null, response.hero.mobile?.object_key ?? null)
  } catch {
    if (!disposed && version === request && preference.mode === 'local') settings.save({ ...preference, mode: 'random' })
    // Optional cloud failure retains the displayed frame. Future reload retries.
  }
}
watch(settings.revision, () => { if (!migrating) void apply() }, { flush: 'sync' })
watch(() => [props.desktopObjectKey, props.mobileObjectKey], () => {
  if (settings.preference.value.mode !== 'local') stage(props.desktopObjectKey ?? null, props.mobileObjectKey ?? null)
})
onMounted(async () => {
  if (settings.legacyPending.value) {
    const version = request
    let url: string | null = null
    try { url = await loadRandomCustomNavHeaderBackground() } catch { /* unavailable legacy storage */ }
    if (disposed || version !== request || !settings.legacyPending.value) { if (url) URL.revokeObjectURL(url); return }
    // This one-time upgrade is the sole permitted SSR random → local handoff.
    // Reuse the loaded Blob rather than reading the folder a second time.
    if (url) {
      migrating = true
      settings.save({ ...settings.preference.value, mode: 'local' })
      migrating = false
      stage(null, null, url)
    } else settings.save(settings.preference.value)
  } else if (initialLocal || (props.desktopObjectKey === undefined && props.mobileObjectKey === undefined)) {
    await apply()
  }
})
onUnmounted(() => { disposed = true; request++; frames.value.forEach(release) })
</script>
