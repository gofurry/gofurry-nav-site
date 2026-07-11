<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { listAllJSON } from '../api'
import type { OptionItem } from '../types'

const props = defineProps<{ endpoint: string; modelValue: string[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const keyword = ref('')
const loading = ref(false)
const error = ref('')
const options = ref<OptionItem[]>([])

const normalizedValue = computed(() => (props.modelValue ?? []).map((item) => String(item)))
const selectedSet = computed(() => new Set(normalizedValue.value))
const filteredOptions = computed(() => {
  const currentKeyword = keyword.value.trim().toLowerCase()
  if (!currentKeyword) {
    return options.value
  }

  return options.value.filter((item) => {
    const label = item.label?.toLowerCase() ?? ''
    const extra = item.extra?.toLowerCase() ?? ''
    const id = String(item.id).toLowerCase()
    return label.includes(currentKeyword) || extra.includes(currentKeyword) || id.includes(currentKeyword)
  })
})

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    const merged = new Map<string, OptionItem>()
    const result = await listAllJSON<OptionItem>(props.endpoint)

    for (const item of result.list) {
      merged.set(String(item.id), {
        ...item,
        id: String(item.id),
      })
    }

    options.value = Array.from(merged.values())
  } catch (loadError) {
    error.value = loadError instanceof Error ? loadError.message : '加载可选项失败'
  } finally {
    loading.value = false
  }
}

function updateValue(next: Iterable<string>) {
  emit('update:modelValue', Array.from(new Set(Array.from(next).map((item) => String(item)))))
}

function toggle(id: string, checked: boolean) {
  const next = new Set(normalizedValue.value)
  if (checked) {
    next.add(String(id))
  } else {
    next.delete(String(id))
  }
  updateValue(next)
}

function selectVisible() {
  const next = new Set(normalizedValue.value)
  for (const item of filteredOptions.value) {
    next.add(String(item.id))
  }
  updateValue(next)
}

function clearVisible() {
  const hidden = new Set(filteredOptions.value.map((item) => String(item.id)))
  updateValue(normalizedValue.value.filter((item) => !hidden.has(String(item))))
}

onMounted(loadAll)
watch(() => props.endpoint, loadAll)
</script>

<template>
  <div class="space-y-3">
    <div class="flex flex-col gap-3 md:flex-row md:items-center">
      <input
        v-model="keyword"
        class="w-full border border-[var(--line)] bg-black/20 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
        placeholder="搜索可选项"
      />
      <div class="flex shrink-0 items-center gap-2 text-xs">
        <button type="button" class="border border-[var(--line)] px-3 py-2" @click="selectVisible">勾选当前结果</button>
        <button type="button" class="border border-[var(--line)] px-3 py-2" @click="clearVisible">取消当前结果</button>
      </div>
    </div>

    <div class="flex items-center justify-between text-xs text-[var(--text-muted)]">
      <span>{{ loading ? '正在加载全部可选项...' : `已加载 ${options.length} 项，已选 ${normalizedValue.length} 项` }}</span>
      <button type="button" class="border border-[var(--line)] px-2 py-1" @click="loadAll">刷新</button>
    </div>

    <div v-if="error" class="text-xs text-[var(--danger)]">{{ error }}</div>

    <div class="max-h-80 overflow-auto border border-[var(--line)] bg-black/10 p-3">
      <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
        <label
          v-for="item in filteredOptions"
          :key="item.id"
          class="flex items-start gap-2 border border-[var(--line)]/70 bg-black/20 px-3 py-2 text-sm"
        >
          <input
            type="checkbox"
            class="mt-0.5 shrink-0"
            :checked="selectedSet.has(String(item.id))"
            @change="toggle(String(item.id), ($event.target as HTMLInputElement).checked)"
          />
          <span class="min-w-0">
            <span class="block break-words">{{ item.label }}</span>
            <span v-if="item.extra" class="block break-words text-xs text-[var(--text-muted)]">{{ item.extra }}</span>
          </span>
        </label>
      </div>

      <div v-if="!loading && filteredOptions.length === 0" class="py-6 text-center text-sm text-[var(--text-muted)]">
        没有匹配的可选项
      </div>
    </div>
  </div>
</template>
