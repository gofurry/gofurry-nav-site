<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  ChevronLeft,
  ChevronRight,
  CloudDownload,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Trash2,
  X,
} from 'lucide-vue-next'
import { getJSON, listJSON, sendJSON } from '../api'
import BulkReplacePanel from '../components/BulkReplacePanel.vue'
import FieldEditor from '../components/FieldEditor.vue'
import { findResource } from '../resources'
import type { ResourceConfig, ResourceField } from '../types'

const route = useRoute()
const config = computed<ResourceConfig | undefined>(() => findResource(String(route.params.section), String(route.params.resource)))
const isGameEditor = computed(() => config.value?.section === 'game' && config.value?.key === 'games')

const keyword = ref('')
const pageNum = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref<Record<string, unknown>[]>([])
const loading = ref(false)
const saving = ref(false)
const deletingId = ref<string | null>(null)
const editingId = ref<string | null>(null)
const editorOpen = ref(false)
const form = reactive<Record<string, unknown>>({})
const message = ref('')
const prefillingSteam = ref(false)

function deepClone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value))
}

function resetForm() {
  Object.keys(form).forEach((key) => delete form[key])
  if (!config.value) return
  Object.assign(form, deepClone(config.value.defaults))
}

function getField(path: string) {
  return getFromTarget(form, path)
}

function setField(path: string, value: unknown) {
  setOnTarget(form, path, value)
}

async function loadList() {
  if (!config.value) return
  loading.value = true
  try {
    const result = await listJSON<Record<string, unknown>>(config.value.listEndpoint, pageNum.value, pageSize.value, keyword.value)
    total.value = result.total
    items.value = result.list
  } finally {
    loading.value = false
  }
}

async function startCreate() {
  editingId.value = null
  message.value = ''
  resetForm()
  editorOpen.value = true
}

async function startEdit(id: string) {
  if (!config.value) return
  message.value = ''
  const data = await getJSON<Record<string, unknown>>(`${config.value.detailEndpoint}/${id}`)
  editingId.value = id
  resetForm()
  Object.assign(form, data)
  editorOpen.value = true
}

function closeEditor() {
  editorOpen.value = false
  editingId.value = null
  message.value = ''
}

async function submit() {
  if (!config.value) return
  saving.value = true
  message.value = ''
  try {
    const payload = normalizePayload()
    if (editingId.value) {
      await sendJSON(`${config.value.detailEndpoint}/${editingId.value}`, 'PUT', payload)
      message.value = '已保存修改'
    } else {
      await sendJSON(config.value.detailEndpoint, 'POST', payload)
      message.value = '已创建新记录'
    }
    await loadList()
    if (!editingId.value) {
      resetForm()
    }
  } catch (error) {
    message.value = error instanceof Error ? error.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function removeItem(id: string) {
  if (!config.value) return
  deletingId.value = id
  try {
    await sendJSON(`${config.value.detailEndpoint}/${id}`, 'DELETE')
    await loadList()
    if (editingId.value === id) {
      closeEditor()
    }
  } finally {
    deletingId.value = null
  }
}

function isWideField(field: ResourceField) {
  if (isGameEditor.value && (field.key === 'appid' || field.key === 'weight')) return true
  return ['textarea', 'string-array', 'kv-array', 'remote-multi'].includes(field.type)
}

function formatCell(value: unknown) {
  if (Array.isArray(value)) {
    return value
      .map((item) => (typeof item === 'object' && item ? JSON.stringify(item) : String(item)))
      .join(' | ')
  }
  if (typeof value === 'object' && value) {
    return JSON.stringify(value)
  }
  return String(value ?? '')
}

function normalizePayload() {
  const payload = deepClone(form)
  for (const field of config.value?.fields ?? []) {
    const value = getFromTarget(payload as Record<string, unknown>, field.key)
    if (field.type === 'number' || field.type === 'remote-select') {
      setOnTarget(payload as Record<string, unknown>, field.key, value === '' || value == null ? 0 : Number(value))
    }
  }
  if (isGameEditor.value) {
    applyGameSteamDefaults(payload as Record<string, unknown>)
  }
  return payload
}

type KVItem = { key: string; value: string }
type SteamPrefillResponse = {
  appid: number
  name: string
  name_en: string
  info: string
  info_en: string
  groups: KVItem[]
  release_date: string
  developers: string[]
  publishers: string[]
  header: string
  links: KVItem[]
}

function currentGameAppid(target: Record<string, unknown> = form) {
  const value = target.appid
  const appid = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(appid) && appid > 0 ? Math.trunc(appid) : 0
}

function defaultGameLinks(appid: number): KVItem[] {
  return [
    { key: 'steamdb', value: `https://steamdb.info/app/${appid}/` },
    { key: 'gamalytic', value: `https://gamalytic.com/game/${appid}` },
  ]
}

function asKVArray(value: unknown): KVItem[] {
  return Array.isArray(value)
    ? value.map((item) => {
        const row = item && typeof item === 'object' ? item as Record<string, unknown> : {}
        return {
          key: String(row.key ?? '').trim(),
          value: String(row.value ?? '').trim(),
        }
      })
    : []
}

function mergeKVArray(current: unknown, incoming: KVItem[]) {
  const result = asKVArray(current)
  for (const item of incoming) {
    const key = item.key.trim()
    const value = item.value.trim()
    if (!key || !value) continue

    const existing = result.find((row) => row.key.toLowerCase() === key.toLowerCase())
    if (existing) {
      if (!existing.value) existing.value = value
      continue
    }
    result.push({ key, value })
  }
  return result
}

function applyGameSteamDefaults(target: Record<string, unknown>) {
  const appid = currentGameAppid(target)
  if (!appid) return false

  const links = asKVArray(target.links)
  for (const item of defaultGameLinks(appid)) {
    const existing = links.find((link) => link.key.toLowerCase() === item.key)
    if (existing) {
      if (!existing.value) existing.value = item.value
      continue
    }
    links.push(item)
  }
  target.links = links
  return true
}

async function prefillGameFromSteam() {
  const appid = currentGameAppid()
  if (!appid) {
    message.value = '请先填写 Steam AppID'
    return
  }

  prefillingSteam.value = true
  message.value = ''
  try {
    const metadata = await getJSON<SteamPrefillResponse>(`/api/v1/game/games/steam-prefill?appid=${appid}`)
    const filled: string[] = []

    if (metadata.name) {
      form.name = metadata.name
      filled.push('中文名')
    }
    if (metadata.name_en) {
      form.name_en = metadata.name_en
      filled.push('英文名')
    }
    if (metadata.info) {
      form.info = metadata.info
      filled.push('中文简介')
    }
    if (metadata.info_en) {
      form.info_en = metadata.info_en
      filled.push('英文简介')
    }
    if (metadata.groups?.length) {
      form.groups = mergeKVArray(form.groups, metadata.groups)
      filled.push('社群')
    }
    if (metadata.release_date) {
      form.release_date = metadata.release_date
      filled.push('发行日')
    }
    if (metadata.developers?.length) {
      form.developers = [...metadata.developers]
      filled.push('开发者')
    }
    if (metadata.publishers?.length) {
      form.publishers = [...metadata.publishers]
      filled.push('发行商')
    }
    if (metadata.header) {
      form.header = metadata.header
      filled.push('封面')
    }
    form.links = mergeKVArray(form.links, metadata.links ?? [])
    applyGameSteamDefaults(form)
    filled.push('Steam 链接')

    message.value = `已从 Steam 填入：${Array.from(new Set(filled)).join('、')}`
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Steam 信息获取失败'
  } finally {
    prefillingSteam.value = false
  }
}

function getFromTarget(target: Record<string, unknown>, path: string) {
  return path.split('.').reduce<unknown>((acc, key) => (acc as Record<string, unknown>)?.[key], target)
}

function setOnTarget(target: Record<string, unknown>, path: string, value: unknown) {
  const keys = path.split('.')
  let cursor = target
  keys.forEach((key, index) => {
    if (index === keys.length - 1) {
      cursor[key] = value
      return
    }
    if (!cursor[key] || typeof cursor[key] !== 'object') {
      cursor[key] = {}
    }
    cursor = cursor[key] as Record<string, unknown>
  })
}

watch(() => route.fullPath, async () => {
  pageNum.value = 1
  keyword.value = ''
  editorOpen.value = false
  editingId.value = null
  resetForm()
  await loadList()
})

onMounted(async () => {
  resetForm()
  await loadList()
})
</script>

<template>
  <div v-if="config" class="resource-page">
    <header class="resource-command-bar">
      <div class="resource-command-bar__title">
        <h1>{{ config.title }}</h1>
        <span>{{ total }} 条</span>
        <span v-if="loading" class="text-[var(--accent)]">刷新中</span>
      </div>

      <div class="resource-command-bar__actions">
        <label class="resource-search">
          <Search :size="17" aria-hidden="true" />
          <input
            v-model="keyword"
            placeholder="搜索关键字"
            @keyup.enter="loadList"
          />
        </label>
        <button class="ui-button px-4 py-2 text-sm" type="button" @click="loadList">
          查询
        </button>
        <button class="ui-button ui-button--primary flex items-center gap-2 px-4 py-2 text-sm" type="button" @click="startCreate">
          <Plus :size="17" />
          新建
        </button>
      </div>
    </header>

    <BulkReplacePanel v-if="config.bulkReplace" :config="config" @saved="loadList" />

    <section class="resource-table-section">
      <div class="data-surface resource-table-scroll">
        <table class="resource-table">
          <thead>
            <tr>
              <th v-for="column in config.columns" :key="column.key">{{ column.label }}</th>
              <th class="resource-table__actions-heading">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in items"
              :key="String(item.id)"
              :class="{ 'resource-table__row--active': editingId === String(item.id) && editorOpen }"
            >
              <td v-for="column in config.columns" :key="column.key">{{ formatCell(item[column.key]) }}</td>
              <td class="resource-table__actions">
                <button
                  class="table-action"
                  type="button"
                  title="编辑"
                  aria-label="编辑"
                  @click="startEdit(String(item.id))"
                >
                  <Pencil :size="16" />
                </button>
                <button
                  class="table-action table-action--danger"
                  type="button"
                  title="删除"
                  aria-label="删除"
                  :disabled="deletingId === String(item.id)"
                  @click="removeItem(String(item.id))"
                >
                  <Trash2 :size="16" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="!loading && items.length === 0" class="resource-empty">
          没有匹配的数据
        </div>
      </div>

      <footer class="resource-pagination">
        <span>第 {{ pageNum }} 页</span>
        <div class="flex items-center gap-1">
          <button
            class="table-action"
            type="button"
            title="上一页"
            aria-label="上一页"
            :disabled="pageNum <= 1"
            @click="pageNum -= 1; loadList()"
          >
            <ChevronLeft :size="18" />
          </button>
          <button
            class="table-action"
            type="button"
            title="下一页"
            aria-label="下一页"
            :disabled="items.length < pageSize"
            @click="pageNum += 1; loadList()"
          >
            <ChevronRight :size="18" />
          </button>
        </div>
      </footer>
    </section>

    <Teleport to="body">
      <aside v-if="editorOpen" class="resource-editor" :aria-label="editingId ? `编辑 #${editingId}` : '新建记录'">
        <header class="resource-editor__header">
          <div>
            <div class="resource-editor__eyebrow">{{ config.title }}</div>
            <h2>{{ editingId ? `编辑 #${editingId}` : '新建记录' }}</h2>
          </div>
          <button class="icon-button" type="button" title="关闭编辑器" aria-label="关闭编辑器" @click="closeEditor">
            <X :size="20" />
          </button>
        </header>

        <div class="resource-editor__body">
          <template v-for="field in config.fields" :key="field.key">
            <div :class="{ 'resource-editor__field--wide': isWideField(field) }">
              <FieldEditor
                :field="field"
                :model-value="getField(field.key)"
                @update:model-value="setField(field.key, $event)"
              />
            </div>
          </template>
        </div>

        <footer class="resource-editor__footer">
          <div class="min-w-0 flex-1 text-xs text-[var(--text-muted)]">{{ message }}</div>
          <button
            v-if="isGameEditor"
            class="ui-button flex items-center gap-2 px-3 py-2 text-sm"
            type="button"
            :disabled="prefillingSteam"
            @click="prefillGameFromSteam"
          >
            <CloudDownload :size="16" />
            {{ prefillingSteam ? '获取中' : 'Steam 填入' }}
          </button>
          <button class="ui-button flex items-center gap-2 px-3 py-2 text-sm" type="button" @click="resetForm">
            <RotateCcw :size="16" />
            重置
          </button>
          <button class="ui-button ui-button--primary px-4 py-2 text-sm" type="button" :disabled="saving" @click="submit">
            {{ saving ? '保存中' : editingId ? '保存修改' : '创建记录' }}
          </button>
        </footer>
      </aside>
    </Teleport>
  </div>
</template>
