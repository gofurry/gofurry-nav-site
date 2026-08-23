<template>
  <img
    :src="currentSrc"
    :alt="alt"
    @error="handleError"
  >
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ensureSteamSharedCdnPreference, steamSharedAssetCandidates } from '@/utils/steamAssets'

const props = withDefaults(defineProps<{
  src?: string | null
  alt?: string
}>(), {
  alt: '',
})

const emit = defineEmits<{
  (event: 'error', value: Event): void
}>()

const { locale } = useI18n()
const failedIndex = ref(0)
const preferenceVersion = ref(0)
const candidates = computed(() => {
  preferenceVersion.value
  return steamSharedAssetCandidates(props.src, locale.value)
})
const currentSrc = computed(() => candidates.value[failedIndex.value] ?? '')

watch(
  [() => props.src, () => locale.value],
  () => {
    failedIndex.value = 0
    ensureSteamSharedCdnPreference(props.src, locale.value)
  }
)

onMounted(() => {
  ensureSteamSharedCdnPreference(props.src, locale.value)
  window.addEventListener('gofurry:steam-shared-cdn-preference-updated', handlePreferenceUpdated)
})

onBeforeUnmount(() => {
  window.removeEventListener('gofurry:steam-shared-cdn-preference-updated', handlePreferenceUpdated)
})

function handleError(event: Event) {
  if (failedIndex.value < candidates.value.length - 1) {
    failedIndex.value += 1
    return
  }

  emit('error', event)
}

function handlePreferenceUpdated() {
  failedIndex.value = 0
  preferenceVersion.value += 1
}
</script>
