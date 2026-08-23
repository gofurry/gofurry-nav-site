<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { getJSON, sendJSON } from '../api'
import type { BulkReplaceConfig } from '../types'
import FieldEditor from './FieldEditor.vue'

const props = defineProps<{ config: BulkReplaceConfig }>()
const emit = defineEmits<{ saved: [] }>()

const state = reactive<{ owner_id: string; ids: string[] }>({ owner_id: '', ids: [] })
const saving = ref(false)
const loadingSelection = ref(false)
const selectionReady = ref(false)
const message = ref('')
let selectionToken = 0

async function loadSelection(ownerID: string) {
  const token = ++selectionToken
  state.ids = []
  selectionReady.value = false
  message.value = ''
  if (!ownerID) return
  if (!props.config.selectionEndpoint) {
    selectionReady.value = true
    return
  }

  loadingSelection.value = true
  try {
    const endpoint = props.config.selectionEndpoint.replace('{ownerId}', encodeURIComponent(ownerID))
    const ids = await getJSON<Array<string | number>>(endpoint)
    if (token !== selectionToken) return
    state.ids = ids.map((id) => String(id))
    selectionReady.value = true
  } catch (error) {
    if (token !== selectionToken) return
    message.value = error instanceof Error ? error.message : '加载已有映射失败'
  } finally {
    if (token === selectionToken) loadingSelection.value = false
  }
}

watch(() => state.owner_id, loadSelection)
watch(() => props.config.endpoint, () => {
  selectionToken += 1
  state.owner_id = ''
  state.ids = []
  loadingSelection.value = false
  selectionReady.value = false
  message.value = ''
})

async function submit() {
  saving.value = true
  message.value = ''
  try {
    await sendJSON(props.config.endpoint, 'PUT', {
      owner_id: Number(state.owner_id),
      ids: state.ids.map((item) => Number(item)),
    })
    message.value = '批量映射已保存'
    emit('saved')
  } catch (error) {
    message.value = error instanceof Error ? error.message : '保存失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="workspace-section space-y-4">
    <div>
      <div class="text-sm font-semibold">{{ config.title ?? '批量替换映射' }}</div>
      <div class="mt-1 text-xs text-[var(--text-muted)]">
        {{ config.description ?? '适合一次维护一个主体的整组映射。' }}
      </div>
    </div>
    <FieldEditor :field="config.ownerField" :model-value="state.owner_id" @update:model-value="state.owner_id = String($event ?? '')" />
    <FieldEditor :field="config.targetField" :model-value="state.ids" @update:model-value="state.ids = ($event as string[]) ?? []" />
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="ui-button ui-button--primary px-4 py-2 text-sm"
        :disabled="!state.owner_id || !selectionReady || saving || loadingSelection"
        @click="submit"
      >
        {{ saving ? '保存中…' : loadingSelection ? '加载中…' : '保存映射' }}
      </button>
      <span class="text-xs text-[var(--text-muted)]">{{ message }}</span>
    </div>
  </section>
</template>
