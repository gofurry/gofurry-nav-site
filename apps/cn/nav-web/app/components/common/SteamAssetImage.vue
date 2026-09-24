<template>
  <img
    ref="image"
    :src="currentSrc"
    :alt="alt"
    @error="handleError"
  >
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  src?: string | null
  alt?: string
}>(), {
  alt: '',
})

const emit = defineEmits<{
  (event: 'error', value: Event): void
}>()

const image = ref<HTMLImageElement | null>(null)
const { src: currentSrc, onError } = useSteamAsset(() => props.src)

onMounted(() => {
  if (currentSrc.value && image.value?.complete && image.value.naturalWidth === 0) handleError(new Event('error'))
})

function handleError(event: Event) {
  if (!onError()) emit('error', event)
}

</script>
