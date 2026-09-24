<template>
  <img v-if="localSrc" ref="image" class="nav-header__background nav-header__background--local" :src="localSrc" alt="" aria-hidden="true" @load="ready" @error="emit('failed')" />
  <picture v-else class="nav-header__background nav-header__background--managed" aria-hidden="true">
    <source media="(min-width: 768px)" :srcset="desktop.src.value || emptyImage" />
    <img ref="image" :src="mobile.src.value || emptyImage" alt="" @load="ready" @error="failed" />
  </picture>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { ManagedAssetSnapshot } from '~/composables/useManagedAsset'
const props = defineProps<{ desktopKey: string | null; mobileKey: string | null; localSrc: string | null; previous?: { desktop: ManagedAssetSnapshot; mobile: ManagedAssetSnapshot } }>()
const emit = defineEmits<{ ready: []; failed: [] }>()
// Each keyed frame owns immutable resources. Only its actual renderer can
// advance fallback; no hidden Image/preload gets to veto a painted frame.
const desktop = useManagedAsset(() => props.desktopKey, '', false, props.previous?.desktop)
const mobile = useManagedAsset(() => props.mobileKey, '', false, props.previous?.mobile)
defineExpose({ snapshot: () => ({ desktop: desktop.snapshot(), mobile: mobile.snapshot() }) })
const image = ref<HTMLImageElement | null>(null)
const emptyImage = 'data:image/svg+xml,%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20width=%221%22%20height=%221%22/%3E'
function ready() {
  const element = image.value
  if (!element?.complete || !element.naturalWidth) return
  const isDesktop = window.matchMedia('(min-width: 768px)').matches
  const key = isDesktop ? props.desktopKey : props.mobileKey
  if (!props.localSrc && key && element.currentSrc === emptyImage) emit('failed')
  else emit('ready')
}
function failed() {
  const element = image.value
  if (!element || element.naturalWidth > 0) return
  if (element.currentSrc === desktop.src.value) desktop.onError()
  else if (element.currentSrc === mobile.src.value) mobile.onError()
}
onMounted(() => {
  if (!image.value?.complete) return
  if (image.value.naturalWidth > 0) ready()
  else if (props.localSrc) emit('failed')
  else failed()
})
</script>

<style scoped>
.nav-header__background--managed img, .nav-header__background--local { width: 100%; height: 100%; display: block; object-fit: cover; object-position: center; }
</style>
