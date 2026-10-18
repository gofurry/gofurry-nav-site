<script setup lang="ts">
import { renderUpdateMarkdown } from '~/utils/updateMarkdown'
const props = defineProps<{ source: string }>()
const sanitizedHTML = computed(() => renderUpdateMarkdown(props.source))
</script>

<template>
  <!-- Only the shared-contract renderer's sanitized result enters this HTML boundary. -->
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div class="updates-markdown" data-update-markdown v-html="sanitizedHTML" />
</template>
