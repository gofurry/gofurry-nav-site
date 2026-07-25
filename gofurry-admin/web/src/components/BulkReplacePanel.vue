<script setup lang="ts">
import { reactive, ref } from 'vue'
import { sendJSON } from '../api'
import type { ResourceConfig } from '../types'
import FieldEditor from './FieldEditor.vue'

const props = defineProps<{ config: ResourceConfig }>()
const emit = defineEmits<{ saved: [] }>()

const state = reactive<{ owner_id: string; ids: string[] }>({ owner_id: '', ids: [] })
const saving = ref(false)
const message = ref('')

async function submit() {
  if (!props.config.bulkReplace) return
  saving.value = true
  message.value = ''
  try {
    await sendJSON(props.config.bulkReplace.endpoint, 'PUT', {
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
  <section v-if="config.bulkReplace" class="workspace-section space-y-4">
    <div>
      <div class="text-sm font-semibold">批量替换映射</div>
      <div class="mt-1 text-xs text-[var(--text-muted)]">适合一次维护一个网站或一个游戏的整组映射。</div>
    </div>
    <FieldEditor :field="config.bulkReplace.ownerField" :model-value="state.owner_id" @update:model-value="state.owner_id = String($event ?? '')" />
    <FieldEditor :field="config.bulkReplace.targetField" :model-value="state.ids" @update:model-value="state.ids = ($event as string[]) ?? []" />
    <div class="flex items-center gap-3">
      <button type="button" class="ui-button ui-button--primary px-4 py-2 text-sm" @click="submit">{{ saving ? '保存中…' : '保存映射' }}</button>
      <span class="text-xs text-[var(--text-muted)]">{{ message }}</span>
    </div>
  </section>
</template>
