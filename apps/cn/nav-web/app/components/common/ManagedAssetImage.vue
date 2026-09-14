<template>
  <img v-if="src" ref="image" :src="src" :alt="alt" @error="onError" />
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
const props = withDefaults(defineProps<{ objectKey?: string | null; alt?: string; fallback?: string }>(), { alt: '', fallback: '/defaultLogo.svg' })
const { src, onError } = useManagedAsset(() => props.objectKey, props.fallback)
const image = ref<HTMLImageElement | null>(null)
onMounted(() => { if (image.value?.complete && image.value.naturalWidth === 0) onError() })
</script>
