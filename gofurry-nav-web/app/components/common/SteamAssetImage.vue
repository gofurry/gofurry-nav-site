<template>
  <img
    :src="currentSrc"
    :alt="alt"
    @error="handleError"
  >
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { steamSharedAssetCandidates } from '@/utils/steamAssets'

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
const candidates = computed(() => steamSharedAssetCandidates(props.src, locale.value))
const currentSrc = computed(() => candidates.value[failedIndex.value] ?? '')

watch(
  [() => props.src, () => locale.value],
  () => {
    failedIndex.value = 0
  }
)

function handleError(event: Event) {
  if (failedIndex.value < candidates.value.length - 1) {
    failedIndex.value += 1
    return
  }

  emit('error', event)
}
</script>
