<template>
  <div class="insight-workspace-selector" role="group" :aria-label="label">
    <button v-for="item in items" :key="item.value" type="button" :data-workspace-option="item.value"
      :aria-pressed="selected === item.value" @click="$emit('select', item.value)" @focus="reveal">
      {{ item.label }}
    </button>
  </div>
</template>

<script setup lang="ts" generic="T extends string">
defineProps<{ items: { value: T; label: string }[]; selected: T; label: string }>()
defineEmits<{ select: [value: T] }>()
function reveal(event: FocusEvent) {
  (event.target as HTMLElement).scrollIntoView({ block: 'nearest', inline: 'nearest' })
}
</script>
