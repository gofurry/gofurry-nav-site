<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { listAllJSON, listJSON } from '../api'
import type { OptionItem } from '../types'

const props = withDefaults(defineProps<{ endpoint: string; modelValue: string; valueField?: 'id' | 'label' }>(), {
  valueField: 'id',
})
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const keyword = ref('')
const loading = ref(false)
const options = ref<OptionItem[]>([])

const normalizedValue = computed(() => props.modelValue ?? '')
const shouldLoadAllOptions = computed(() => props.endpoint.endsWith('/options/tags'))

async function load() {
  loading.value = true
  try {
    const result = shouldLoadAllOptions.value
      ? await listAllJSON<OptionItem>(props.endpoint, keyword.value)
      : await listJSON<OptionItem>(props.endpoint, 1, 50, keyword.value)
    options.value = result.list
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.endpoint, load)
watch(keyword, load)
</script>

<template>
  <div class="space-y-2">
    <input
      v-model="keyword"
      class="ui-control w-full px-3 py-2 text-sm"
      placeholder="搜索可选项"
    />
    <select
      class="ui-control w-full px-3 py-2 text-sm"
      :value="normalizedValue"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option value="">请选择</option>
      <option v-for="item in options" :key="item.id" :value="props.valueField === 'label' ? item.label : item.id">
        {{ item.label }}{{ item.extra ? ` / ${item.extra}` : '' }}
      </option>
    </select>
    <div class="text-xs text-[var(--text-muted)]">{{ loading ? '加载中…' : `共 ${options.length} 项` }}</div>
  </div>
</template>
