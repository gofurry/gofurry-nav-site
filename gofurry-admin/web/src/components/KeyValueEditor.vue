<script setup lang="ts">
import type { KeyValue } from '../types'

const props = defineProps<{ modelValue: KeyValue[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: KeyValue[]] }>()

function patch(index: number, key: 'key' | 'value', value: string) {
  const next = [...(props.modelValue ?? [])]
  next[index] = { ...(next[index] ?? { key: '', value: '' }), [key]: value }
  emit('update:modelValue', next)
}

function addRow() {
  emit('update:modelValue', [...(props.modelValue ?? []), { key: '', value: '' }])
}

function removeRow(index: number) {
  const next = [...(props.modelValue ?? [])]
  next.splice(index, 1)
  emit('update:modelValue', next.length ? next : [{ key: '', value: '' }])
}
</script>

<template>
  <div class="space-y-2">
    <div v-for="(item, index) in modelValue" :key="index" class="grid gap-2 md:grid-cols-[180px_1fr_auto]">
      <input
        class="ui-control px-3 py-2 text-sm"
        :value="item.key"
        placeholder="key"
        @input="patch(index, 'key', ($event.target as HTMLInputElement).value)"
      />
      <input
        class="ui-control px-3 py-2 text-sm"
        :value="item.value"
        placeholder="value"
        @input="patch(index, 'value', ($event.target as HTMLInputElement).value)"
      />
      <button type="button" class="ui-button ui-button--quiet px-3 text-sm" @click="removeRow(index)">删</button>
    </div>
    <button type="button" class="ui-button px-3 py-2 text-sm" @click="addRow">添加一项</button>
  </div>
</template>
