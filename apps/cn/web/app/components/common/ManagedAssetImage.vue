<template>
  <picture v-if="mobileObjectKey" class="contents">
    <source :media="mobileMedia" :srcset="mobile.src.value || undefined" />
    <img v-if="selected.src.value" ref="image" v-bind="$attrs" :src="desktop.src.value || undefined" :alt="alt" @error="onError" />
  </picture>
  <img v-else-if="desktop.src.value" ref="image" v-bind="$attrs" :src="desktop.src.value" :alt="alt" @error="onError" />
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ objectKey?: string | null; mobileObjectKey?: string | null; mobileMedia?: string; alt?: string; fallback?: string }>(), {
  objectKey: null, mobileObjectKey: null, alt: '', fallback: '/defaultLogo.svg', mobileMedia: '(max-width: 1023px)',
})
const emit = defineEmits<{ exhausted: []; pending: [] }>()
const desktop = useManagedAsset(() => props.objectKey, props.fallback)
const mobile = useManagedAsset(() => props.mobileObjectKey, props.fallback)
const isMobile = ref(false)
const selected = computed(() => props.mobileObjectKey && isMobile.value ? mobile : desktop)
const image = ref<HTMLImageElement | null>(null)
let query: MediaQueryList | undefined
function updateViewport() { isMobile.value = query?.matches ?? false }
function onError() {
  // The picture chooses its source before hydration; consult its media query
  // even when the first resource fails before the mounted hook runs.
  if (props.mobileObjectKey) isMobile.value = window.matchMedia(props.mobileMedia).matches
  const route = selected.value
  route.onError()
}
onMounted(() => {
  if (props.mobileObjectKey) {
    query = window.matchMedia(props.mobileMedia)
    updateViewport()
    query.addEventListener('change', updateViewport)
  }
  if (selected.value.src.value && image.value?.complete && image.value.naturalWidth === 0
    && image.value.currentSrc === new URL(selected.value.src.value, document.baseURI).href) onError()
  if (!selected.value.src.value) emit('exhausted')
})
watch(() => selected.value.src.value, value => { if (value) emit('pending'); else emit('exhausted') })
onBeforeUnmount(() => query?.removeEventListener('change', updateViewport))
</script>
