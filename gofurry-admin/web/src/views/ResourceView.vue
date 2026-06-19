<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getJSON, listJSON, sendJSON } from '../api'
import BulkReplacePanel from '../components/BulkReplacePanel.vue'
import FieldEditor from '../components/FieldEditor.vue'
import { findResource } from '../resources'
import type { ResourceConfig } from '../types'

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
const form = reactive<Record<string, unknown>>({})
const message = ref('')
const resolvingSteamAsset = ref(false)

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
}

async function startEdit(id: string) {
  if (!config.value) return
  message.value = ''
  const data = await getJSON<Record<string, unknown>>(`${config.value.detailEndpoint}/${id}`)
  editingId.value = id
  resetForm()
  Object.assign(form, data)
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
  } finally {
    deletingId.value = null
  }
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
type SteamAssetResponse = {
  appid: number
  kind: string
  url: string
  digest?: string
  filename?: string
  source?: string
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

function syncGameSteamLinks() {
  if (!applyGameSteamDefaults(form)) {
    message.value = '请先填写 Steam AppID'
    return
  }
  message.value = '已补全 SteamDB / Gamalytic 链接'
}

async function fillGameHeaderFromSteam() {
  const appid = currentGameAppid()
  if (!appid) {
    message.value = '请先填写 Steam AppID'
    return
  }

  resolvingSteamAsset.value = true
  try {
    const asset = await getJSON<SteamAssetResponse>(`/api/v1/game/games/steam-asset?appid=${appid}&kind=header`)
    form.header = asset.url
    message.value = `已填入 Steam 官方封面：${asset.kind}`
  } catch (error) {
    message.value = error instanceof Error ? error.message : '获取 Steam 官方封面失败'
  } finally {
    resolvingSteamAsset.value = false
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
  resetForm()
  await loadList()
})

onMounted(async () => {
  resetForm()
  await loadList()
})
</script>

<template>
  <div v-if="config" class="space-y-6">
    <div class="border border-[var(--line)] bg-[var(--panel)]/70 p-4">
      <div class="text-xs uppercase tracking-[0.3em] text-[var(--accent)]">{{ config.section === 'nav' ? '导航库' : '游戏库' }}</div>
      <h1 class="mt-2 text-3xl font-semibold">{{ config.title }}</h1>
    </div>

    <BulkReplacePanel v-if="config.bulkReplace" :config="config" @saved="loadList" />

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_560px]">
      <section class="space-y-4 border border-[var(--line)] bg-[var(--panel)]/70 p-4">
        <div class="flex flex-col gap-3 md:flex-row">
          <input v-model="keyword" class="w-full border border-[var(--line)] bg-black/20 px-3 py-2 text-sm outline-none focus:border-[var(--accent)]" placeholder="搜索关键字" />
          <button class="border border-[var(--line)] px-4 py-2 text-sm" @click="loadList">查询</button>
          <button class="border border-[var(--accent)] bg-[var(--accent)]/10 px-4 py-2 text-sm" @click="startCreate">新建</button>
        </div>

        <div class="overflow-auto border border-[var(--line)]">
          <table class="min-w-full border-collapse text-sm">
            <thead class="bg-black/20 text-left text-[var(--text-muted)]">
              <tr>
                <th v-for="column in config.columns" :key="column.key" class="border-b border-[var(--line)] px-3 py-3">{{ column.label }}</th>
                <th class="border-b border-[var(--line)] px-3 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in items" :key="String(item.id)" class="odd:bg-white/[0.02]">
                <td v-for="column in config.columns" :key="column.key" class="border-b border-[var(--line)] px-3 py-3 align-top">{{ formatCell(item[column.key]) }}</td>
                <td class="border-b border-[var(--line)] px-3 py-3">
                  <div class="flex gap-2">
                    <button class="border border-[var(--line)] px-3 py-1" @click="startEdit(String(item.id))">编辑</button>
                    <button class="border border-[var(--danger)] px-3 py-1 text-[var(--danger)]" @click="removeItem(String(item.id))">{{ deletingId === String(item.id) ? '删除中…' : '删除' }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex items-center justify-between text-sm text-[var(--text-muted)]">
          <div>共 {{ total }} 条</div>
          <div class="flex items-center gap-2">
            <button class="border border-[var(--line)] px-3 py-1" :disabled="pageNum <= 1" @click="pageNum -= 1; loadList()">上一页</button>
            <span>第 {{ pageNum }} 页</span>
            <button class="border border-[var(--line)] px-3 py-1" :disabled="items.length < pageSize" @click="pageNum += 1; loadList()">下一页</button>
          </div>
        </div>
      </section>

      <section class="space-y-4 border border-[var(--line)] bg-[var(--panel)]/70 p-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm font-semibold">{{ editingId ? `编辑 #${editingId}` : '新建记录' }}</div>
            <div class="text-xs text-[var(--text-muted)]">右侧直接编辑并提交。</div>
          </div>
          <div class="text-xs text-[var(--text-muted)]">{{ loading ? '列表刷新中…' : '' }}</div>
        </div>

        <template
          v-for="field in config.fields"
          :key="field.key"
        >
          <FieldEditor
            :field="field"
            :model-value="getField(field.key)"
            @update:model-value="setField(field.key, $event)"
          />

          <div
            v-if="isGameEditor && field.key === 'header'"
            class="flex flex-wrap items-center gap-2 border border-[var(--line)] bg-black/20 p-3"
          >
            <button
              class="border border-[var(--line)] px-3 py-1.5 text-xs disabled:opacity-50"
              type="button"
              :disabled="resolvingSteamAsset"
              @click="fillGameHeaderFromSteam"
            >
              {{ resolvingSteamAsset ? '获取中…' : '填入官方封面' }}
            </button>
            <span class="text-xs text-[var(--text-muted)]">通过 Steam StoreBrowse 解析带 hash 的官方资源地址。</span>
          </div>

          <div
            v-if="isGameEditor && field.key === 'links'"
            class="flex flex-wrap items-center gap-2 border border-[var(--line)] bg-black/20 p-3"
          >
            <button class="border border-[var(--line)] px-3 py-1.5 text-xs" type="button" @click="syncGameSteamLinks">
              补全 Steam 链接
            </button>
            <span class="text-xs text-[var(--text-muted)]">自动补 SteamDB 与 Gamalytic 默认链接。</span>
          </div>
        </template>

        <div class="flex items-center gap-3">
          <button class="border border-[var(--accent)] bg-[var(--accent)]/10 px-4 py-2 text-sm" @click="submit">{{ saving ? '保存中…' : editingId ? '保存修改' : '创建记录' }}</button>
          <button class="border border-[var(--line)] px-4 py-2 text-sm" @click="resetForm">重置表单</button>
        </div>
        <div v-if="message" class="text-sm text-[var(--text-muted)]">{{ message }}</div>
      </section>
    </div>
  </div>
</template>
