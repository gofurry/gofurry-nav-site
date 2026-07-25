<script setup lang="ts">
import type { ResourceField } from '../types'
import KeyValueEditor from './KeyValueEditor.vue'
import RemoteMultiSelect from './RemoteMultiSelect.vue'
import RemoteSelect from './RemoteSelect.vue'
import StringArrayEditor from './StringArrayEditor.vue'

defineProps<{ field: ResourceField; modelValue: unknown }>()
const emit = defineEmits<{ 'update:modelValue': [value: unknown] }>()

function dateTimeLocalValue(value: unknown) {
  const raw = String(value ?? '').trim()
  if (!raw) return ''
  const normalized = raw.includes('T') ? raw : raw.replace(' ', 'T')
  const match = normalized.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2})/)
  return match?.[1] ?? ''
}
</script>

<template>
  <div class="space-y-2">
    <label class="block text-sm font-medium tracking-wide text-[var(--text-muted)]">{{ field.label }}</label>

    <input
      v-if="field.type === 'text'"
      class="ui-control w-full px-3 py-2 text-sm"
      :placeholder="field.placeholder"
      :value="String(modelValue ?? '')"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />

    <textarea
      v-else-if="field.type === 'textarea'"
      class="ui-control min-h-28 w-full px-3 py-2 text-sm"
      :value="String(modelValue ?? '')"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />

    <input
      v-else-if="field.type === 'number'"
      type="number"
      class="ui-control w-full px-3 py-2 text-sm"
      :value="Number(modelValue ?? 0)"
      @input="emit('update:modelValue', Number(($event.target as HTMLInputElement).value))"
    />

    <input
      v-else-if="field.type === 'datetime'"
      type="datetime-local"
      class="ui-control w-full px-3 py-2 text-sm"
      :value="dateTimeLocalValue(modelValue)"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />

    <label v-else-if="field.type === 'bool'" class="ui-control flex items-center gap-3 px-3 py-2 text-sm">
      <input
        type="checkbox"
        :checked="Boolean(modelValue)"
        @change="emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
      />
      <span>{{ modelValue ? '启用' : '停用' }}</span>
    </label>

    <select
      v-else-if="field.type === 'select'"
      class="ui-control w-full px-3 py-2 text-sm"
      :value="String(modelValue ?? '')"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-for="option in field.options ?? []" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>

    <StringArrayEditor
      v-else-if="field.type === 'string-array'"
      :model-value="Array.isArray(modelValue) ? (modelValue as string[]) : ['']"
      @update:model-value="emit('update:modelValue', $event)"
    />

    <KeyValueEditor
      v-else-if="field.type === 'kv-array'"
      :model-value="Array.isArray(modelValue) ? (modelValue as never[]) : [{ key: '', value: '' }]"
      @update:model-value="emit('update:modelValue', $event)"
    />

    <RemoteSelect
      v-else-if="field.type === 'remote-select' && field.optionEndpoint"
      :endpoint="field.optionEndpoint"
      :model-value="String(modelValue ?? '')"
      @update:model-value="emit('update:modelValue', $event)"
    />

    <RemoteMultiSelect
      v-else-if="field.type === 'remote-multi' && field.optionEndpoint"
      :endpoint="field.optionEndpoint"
      :model-value="Array.isArray(modelValue) ? (modelValue as string[]) : []"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>
