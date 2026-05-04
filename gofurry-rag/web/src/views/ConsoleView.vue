<template>
  <main class="min-h-screen overflow-hidden bg-[#05080d] text-slate-100">
    <section
      v-if="!authenticated"
      class="gate-enter relative grid min-h-screen place-items-center px-6"
    >
      <div class="absolute inset-0 opacity-70">
        <div class="absolute left-[18%] top-[14%] h-48 w-48 rounded-full bg-teal-400/10 blur-3xl" />
        <div class="absolute bottom-[18%] right-[20%] h-64 w-64 rounded-full bg-cyan-300/10 blur-3xl" />
      </div>

      <form
        class="relative w-full max-w-sm border border-white/10 bg-white/[0.045] p-8 shadow-2xl shadow-teal-950/30 backdrop-blur-xl"
        @submit.prevent="performLogin"
      >
        <div class="mb-10 flex items-center gap-3">
          <div class="grid h-11 w-11 place-items-center border border-teal-300/30 bg-teal-300/10 text-teal-200">
            <ShieldCheck :size="22" />
          </div>
          <div>
            <p class="text-sm uppercase tracking-[0.28em] text-teal-200/70">gofurry-rag</p>
            <h1 class="mt-1 text-xl font-semibold text-white">控制台入口</h1>
          </div>
        </div>

        <label class="block text-sm text-slate-300" for="passcode">唯一口令</label>
        <input
          id="passcode"
          v-model="password"
          class="mt-3 h-12 w-full border border-white/10 bg-black/30 px-4 text-white outline-none transition focus:border-teal-300/70"
          type="password"
          autocomplete="current-password"
          autofocus
        />
        <button
          class="mt-6 flex h-12 w-full items-center justify-center gap-2 bg-teal-300 px-4 text-sm font-semibold text-slate-950 transition hover:bg-teal-200 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="busy"
          type="submit"
        >
          <LogIn :size="18" />
          进入
        </button>
        <p v-if="notice" class="mt-4 text-sm text-rose-300">{{ notice }}</p>
      </form>
    </section>

    <section v-else class="grid min-h-screen grid-cols-[260px_minmax(0,1fr)]">
      <aside class="border-r border-white/10 bg-black/20 px-5 py-6 backdrop-blur-xl">
        <div class="mb-10 flex items-center gap-3">
          <div class="grid h-10 w-10 place-items-center border border-teal-300/30 bg-teal-300/10 text-teal-200">
            <Database :size="21" />
          </div>
          <div>
            <strong class="block text-sm tracking-[0.2em] text-white">GOFURRY RAG</strong>
            <span class="text-xs text-slate-500">Knowledge Observatory</span>
          </div>
        </div>

        <nav class="space-y-2">
          <button
            v-for="item in menuItems"
            :key="item.key"
            class="flex h-11 w-full items-center gap-3 border px-3 text-left text-sm transition"
            :class="activeMenu === item.key ? 'border-teal-300/35 bg-teal-300/10 text-teal-100' : 'border-transparent text-slate-400 hover:border-white/10 hover:bg-white/[0.045] hover:text-slate-100'"
            @click="activeMenu = item.key"
          >
            <component :is="item.icon" :size="18" />
            {{ item.label }}
          </button>
        </nav>

        <div class="absolute bottom-6 left-5 right-auto w-[220px]">
          <button
            class="flex h-10 w-full items-center justify-center gap-2 border border-white/10 text-sm text-slate-300 transition hover:border-rose-300/30 hover:bg-rose-300/10 hover:text-rose-100"
            @click="performLogout"
          >
            <LogOut :size="17" />
            退出
          </button>
        </div>
      </aside>

      <section class="thin-scrollbar max-h-screen overflow-y-auto px-8 py-7">
        <header class="mb-7 flex items-end justify-between gap-6">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-teal-200/60">{{ currentKicker }}</p>
            <h2 class="mt-2 text-3xl font-semibold tracking-tight text-white">{{ currentTitle }}</h2>
          </div>
          <div class="flex items-center gap-3 text-sm text-slate-400">
            <span class="h-2 w-2 rounded-full bg-teal-300 shadow-[0_0_22px_rgba(45,212,191,0.9)]" />
            {{ healthState.status || 'unknown' }}
          </div>
        </header>

        <transition name="fade-slide" mode="out-in">
          <section v-if="activeMenu === 'overview'" key="overview" class="space-y-7">
            <div class="grid gap-px overflow-hidden border border-white/10 bg-white/10 md:grid-cols-4">
              <MetricCell label="文档" :value="overviewData?.document_total ?? 0" />
              <MetricCell label="Chunks" :value="overviewData?.chunk_total ?? 0" />
              <MetricCell label="已向量化" :value="overviewData?.embedded_chunk_total ?? 0" />
              <MetricCell label="Ready" :value="overviewData?.ready_documents ?? 0" />
            </div>
            <div class="grid gap-6 lg:grid-cols-[1fr_360px]">
              <div class="border border-white/10 bg-white/[0.035] p-6">
                <div class="mb-5 flex items-center gap-2 text-slate-300">
                  <Activity :size="18" class="text-teal-200" />
                  状态分布
                </div>
                <div class="space-y-4">
                  <StatusBar label="pending" :value="overviewData?.pending_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-slate-400" />
                  <StatusBar label="processing" :value="overviewData?.processing_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-amber-300" />
                  <StatusBar label="ready" :value="overviewData?.ready_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-teal-300" />
                  <StatusBar label="failed" :value="overviewData?.failed_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-rose-300" />
                </div>
              </div>
              <div class="border border-white/10 bg-white/[0.035] p-6">
                <div class="mb-6 flex items-center gap-2 text-slate-300">
                  <Clock3 :size="18" class="text-teal-200" />
                  最近更新
                </div>
                <p class="text-2xl text-white">{{ formatDate(overviewData?.last_document_update_at) }}</p>
                <button class="mt-8 flex items-center gap-2 text-sm text-teal-200 transition hover:text-teal-100" @click="loadOverview">
                  <RefreshCw :size="16" />
                  刷新态势
                </button>
              </div>
            </div>
          </section>

          <section v-else-if="activeMenu === 'documents'" key="documents" class="space-y-6">
            <div class="flex gap-2 border-b border-white/10">
              <button
                v-for="tab in documentTabs"
                :key="tab.key"
                class="px-4 py-3 text-sm transition"
                :class="documentTab === tab.key ? 'border-b border-teal-300 text-teal-100' : 'text-slate-500 hover:text-slate-200'"
                @click="documentTab = tab.key"
              >
                {{ tab.label }}
              </button>
            </div>

            <div v-if="documentTab === 'ingest'" class="grid gap-6 xl:grid-cols-[minmax(0,560px)_1fr]">
              <form class="space-y-5 border border-white/10 bg-white/[0.035] p-6" @submit.prevent="submitText">
                <Field label="标题">
                  <input v-model="form.title" class="control" placeholder="GoFurry 网站介绍" />
                </Field>
                <div class="grid gap-4 md:grid-cols-2">
                  <Field label="来源类型">
                    <input v-model="form.source_type" class="control" placeholder="manual" />
                  </Field>
                  <Field label="来源 ID">
                    <input v-model="form.source_id" class="control" placeholder="about-page" />
                  </Field>
                </div>
                <Field label="URL">
                  <input v-model="form.url" class="control" placeholder="https://example.com/about" />
                </Field>
                <Field label="正文">
                  <textarea v-model="form.content" class="control min-h-56 resize-none py-3" />
                </Field>
                <button class="primary-button" :disabled="busy" type="submit">
                  <Send :size="17" />
                  提交入库
                </button>
              </form>
              <div class="border border-white/10 bg-black/10 p-6 text-slate-400">
                <FileText :size="24" class="mb-5 text-teal-200" />
                <p class="max-w-xl leading-7">
                  文本提交后进入 pending，后台 worker 会切分、生成 embedding，并写入 pgvector。文档列表会自动刷新状态。
                </p>
              </div>
            </div>

            <div v-else class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
              <div class="border border-white/10 bg-white/[0.03]">
                <div class="flex flex-wrap items-center gap-3 border-b border-white/10 p-4">
                  <select v-model="filters.status" class="control h-10 w-40" @change="loadDocuments">
                    <option value="">全部状态</option>
                    <option value="pending">pending</option>
                    <option value="processing">processing</option>
                    <option value="ready">ready</option>
                    <option value="failed">failed</option>
                  </select>
                  <input v-model="filters.keyword" class="control h-10 w-64" placeholder="标题关键词" @keyup.enter="loadDocuments" />
                  <button class="ghost-button" @click="loadDocuments">
                    <RefreshCw :size="16" />
                    刷新
                  </button>
                  <span class="ml-auto text-xs text-slate-500">每 3 秒自动刷新</span>
                </div>
                <div class="thin-scrollbar max-h-[560px] overflow-auto">
                  <table class="w-full border-collapse text-sm">
                    <thead class="sticky top-0 bg-[#080d14] text-left text-xs uppercase tracking-[0.16em] text-slate-500">
                      <tr>
                        <th class="px-4 py-3">ID</th>
                        <th class="px-4 py-3">标题</th>
                        <th class="px-4 py-3">状态</th>
                        <th class="px-4 py-3">Chunks</th>
                        <th class="px-4 py-3">更新</th>
                        <th class="px-4 py-3"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="doc in documents.items" :key="doc.id" class="border-t border-white/10 transition hover:bg-white/[0.04]">
                        <td class="px-4 py-4 text-slate-500">#{{ doc.id }}</td>
                        <td class="px-4 py-4">
                          <p class="font-medium text-slate-100">{{ doc.title || 'Untitled' }}</p>
                          <p class="mt-1 text-xs text-slate-500">{{ doc.source_type }} {{ doc.source_id }}</p>
                        </td>
                        <td class="px-4 py-4">
                          <span class="status-pill" :class="statusClass(doc.status)">{{ doc.status }}</span>
                          <p v-if="doc.error_message" class="mt-2 text-xs text-rose-300">{{ doc.error_message }}</p>
                        </td>
                        <td class="px-4 py-4 text-slate-300">{{ doc.chunk_count }}</td>
                        <td class="px-4 py-4 text-slate-500">{{ formatDate(doc.updated_at) }}</td>
                        <td class="px-4 py-4">
                          <div class="flex justify-end gap-2">
                            <button class="icon-button" title="查看 Chunks" @click="selectDocument(doc)">
                              <Eye :size="16" />
                            </button>
                            <button class="icon-button text-rose-200 hover:border-rose-300/40 hover:bg-rose-300/10" title="删除" @click="removeDocument(doc.id)">
                              <Trash2 :size="16" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <aside class="border border-white/10 bg-white/[0.035] p-5">
                <div class="mb-4 flex items-center gap-2 text-slate-300">
                  <Layers :size="18" class="text-teal-200" />
                  Chunks {{ selectedDocument ? `#${selectedDocument.id}` : '' }}
                </div>
                <div v-if="!selectedDocument" class="py-16 text-center text-sm text-slate-500">选择文档查看 chunks</div>
                <div v-else class="thin-scrollbar max-h-[580px] space-y-3 overflow-auto pr-1">
                  <article v-for="chunk in chunks.items" :key="chunk.id" class="border border-white/10 bg-black/20 p-4 transition hover:border-teal-300/25">
                    <div class="mb-3 flex items-center justify-between text-xs text-slate-500">
                      <span>#{{ chunk.chunk_index }} · {{ chunk.token_count }} chars</span>
                      <span :class="chunk.has_embedding ? 'text-teal-200' : 'text-amber-200'">{{ chunk.embedding_dim }}d</span>
                    </div>
                    <p class="text-sm leading-6 text-slate-300">{{ chunk.content }}</p>
                  </article>
                </div>
              </aside>
            </div>
          </section>

          <section v-else key="search" class="grid gap-6 xl:grid-cols-[520px_1fr]">
            <form class="border border-white/10 bg-white/[0.035] p-6" @submit.prevent="runQuery">
              <Field label="问题">
                <textarea v-model="question" class="control min-h-36 resize-none py-3" />
              </Field>
              <Field label="Top K">
                <input v-model.number="topK" class="control" type="number" min="1" max="20" />
              </Field>
              <button class="primary-button mt-5" :disabled="busy" type="submit">
                <Search :size="17" />
                检索
              </button>
            </form>

            <div class="border border-white/10 bg-white/[0.03] p-6">
              <div class="mb-5 flex items-center gap-2 text-slate-300">
                <BookOpen :size="18" class="text-teal-200" />
                Sources
              </div>
              <div v-if="!queryResult" class="py-20 text-center text-sm text-slate-500">等待检索</div>
              <div v-else class="space-y-4">
                <p class="text-slate-300">{{ queryResult.answer }}</p>
                <article v-for="source in queryResult.sources" :key="source.chunk_id" class="border-l border-teal-300/40 bg-black/20 p-4">
                  <div class="mb-2 flex items-center justify-between gap-4">
                    <strong class="text-sm text-white">{{ source.title || `Document #${source.document_id}` }}</strong>
                    <span class="text-xs text-teal-200">{{ source.score.toFixed(4) }}</span>
                  </div>
                  <p class="text-sm leading-6 text-slate-400">{{ source.content }}</p>
                </article>
              </div>
            </div>
          </section>
        </transition>

        <p v-if="notice" class="fixed bottom-5 right-6 border border-teal-300/20 bg-black/80 px-4 py-3 text-sm text-teal-100 shadow-xl shadow-black/30">
          {{ notice }}
        </p>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import {
  Activity,
  BookOpen,
  Clock3,
  Database,
  Eye,
  FileText,
  Gauge,
  Layers,
  LogIn,
  LogOut,
  RefreshCw,
  Search,
  Send,
  ShieldCheck,
  Trash2,
} from 'lucide-vue-next'
import {
  authState,
  createTextDocument,
  deleteDocument,
  health,
  listChunks,
  listDocuments,
  login,
  logout,
  overview,
  queryRag,
} from '../api'
import type { ChunkItem, DocumentItem, Overview, PageResult, QueryResponse } from '../types'

type MenuKey = 'overview' | 'documents' | 'search'
type DocumentTab = 'ingest' | 'list'

const MetricCell = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: Number, required: true },
  },
  setup(props) {
    return () =>
      h('div', { class: 'bg-[#080d14] p-5' }, [
        h('p', { class: 'text-xs uppercase tracking-[0.18em] text-slate-500' }, props.label),
        h('p', { class: 'mt-3 text-3xl font-semibold text-white' }, String(props.value)),
      ])
  },
})

const StatusBar = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: Number, required: true },
    total: { type: Number, required: true },
    tone: { type: String, required: true },
  },
  setup(props) {
    return () => {
      const pct = props.total > 0 ? Math.round((props.value / props.total) * 100) : 0
      return h('div', [
        h('div', { class: 'mb-2 flex justify-between text-sm' }, [
          h('span', { class: 'text-slate-400' }, props.label),
          h('span', { class: 'text-slate-200' }, `${props.value} · ${pct}%`),
        ]),
        h('div', { class: 'h-1.5 bg-white/10' }, [
          h('div', { class: `h-full ${props.tone}`, style: { width: `${pct}%` } }),
        ]),
      ])
    }
  },
})

const Field = defineComponent({
  props: {
    label: { type: String, required: true },
  },
  setup(props, { slots }) {
    return () =>
      h('label', { class: 'block' }, [
        h('span', { class: 'mb-2 block text-sm text-slate-400' }, props.label),
        slots.default?.(),
      ])
  },
})

const menuItems = [
  { key: 'overview' as MenuKey, label: '整体态势', icon: Gauge },
  { key: 'documents' as MenuKey, label: '文档管理', icon: FileText },
  { key: 'search' as MenuKey, label: '文档检索', icon: Search },
]
const documentTabs = [
  { key: 'ingest' as DocumentTab, label: '文本入库' },
  { key: 'list' as DocumentTab, label: '文档' },
]

const authenticated = ref(false)
const password = ref('')
const busy = ref(false)
const notice = ref('')
const activeMenu = ref<MenuKey>('overview')
const documentTab = ref<DocumentTab>('ingest')
const healthState = reactive<Record<string, unknown>>({ status: 'unknown' })
const overviewData = ref<Overview | null>(null)
const documents = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const chunks = reactive<PageResult<ChunkItem>>({ items: [], total: 0 })
const selectedDocument = ref<DocumentItem | null>(null)
const filters = reactive({ status: '', keyword: '' })
const queryResult = ref<QueryResponse | null>(null)
const question = ref('GoFurry 会返回什么？')
const topK = ref(6)
const form = reactive({
  title: '',
  source_type: 'manual',
  source_id: '',
  url: '',
  content: '',
})
let documentPoll: number | undefined

const currentTitle = computed(() => {
  if (activeMenu.value === 'documents') return '文档管理'
  if (activeMenu.value === 'search') return '文档检索'
  return '整体态势'
})

const currentKicker = computed(() => {
  if (activeMenu.value === 'documents') return 'INGEST / DOCUMENTS'
  if (activeMenu.value === 'search') return 'RETRIEVAL'
  return 'OBSERVABILITY'
})

async function performLogin() {
  busy.value = true
  notice.value = ''
  try {
    await login(password.value)
    authenticated.value = true
    password.value = ''
    await loadInitialData()
  } catch (error) {
    notice.value = (error as Error).message
  } finally {
    busy.value = false
  }
}

async function performLogout() {
  await logout()
  authenticated.value = false
  stopDocumentPolling()
  activeMenu.value = 'overview'
  documents.items = []
  chunks.items = []
  selectedDocument.value = null
}

async function loadInitialData() {
  await Promise.all([loadHealth(), loadOverview()])
}

async function loadHealth() {
  Object.assign(healthState, await health())
}

async function loadOverview() {
  overviewData.value = await overview()
}

async function loadDocuments() {
  const result = await listDocuments({ page: 1, page_size: 80, status: filters.status, keyword: filters.keyword })
  documents.items = result.items
  documents.total = result.total
  if (selectedDocument.value) {
    const fresh = result.items.find((item) => item.id === selectedDocument.value?.id)
    if (fresh) selectedDocument.value = fresh
  }
}

async function submitText() {
  busy.value = true
  notice.value = ''
  try {
    await createTextDocument({ ...form, metadata: {} })
    form.content = ''
    notice.value = '已创建 pending 文档'
    documentTab.value = 'list'
    await Promise.all([loadDocuments(), loadOverview()])
  } catch (error) {
    notice.value = (error as Error).message
  } finally {
    busy.value = false
  }
}

async function selectDocument(doc: DocumentItem) {
  selectedDocument.value = doc
  const result = await listChunks(doc.id)
  chunks.items = result.items
  chunks.total = result.total
}

async function removeDocument(id: number) {
  if (!window.confirm(`删除文档 #${id}？`)) return
  await deleteDocument(id)
  if (selectedDocument.value?.id === id) {
    selectedDocument.value = null
    chunks.items = []
  }
  await Promise.all([loadDocuments(), loadOverview()])
}

async function runQuery() {
  busy.value = true
  notice.value = ''
  try {
    queryResult.value = await queryRag(question.value, topK.value)
  } catch (error) {
    notice.value = (error as Error).message
  } finally {
    busy.value = false
  }
}

function startDocumentPolling() {
  stopDocumentPolling()
  void loadDocuments()
  documentPoll = window.setInterval(() => {
    void loadDocuments()
  }, 3000)
}

function stopDocumentPolling() {
  if (documentPoll) {
    window.clearInterval(documentPoll)
    documentPoll = undefined
  }
}

function statusClass(status: string) {
  if (status === 'ready') return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
  if (status === 'processing') return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  if (status === 'failed') return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  return 'border-slate-400/20 bg-slate-400/10 text-slate-300'
}

function formatDate(value?: string) {
  if (!value) return '尚无记录'
  return new Date(value).toLocaleString()
}

watch([activeMenu, documentTab, authenticated], ([menu, tab, loggedIn]) => {
  if (loggedIn && menu === 'documents' && tab === 'list') {
    startDocumentPolling()
  } else {
    stopDocumentPolling()
  }
})

onMounted(async () => {
  try {
    const state = await authState()
    authenticated.value = state.authenticated
    if (state.authenticated) {
      await loadInitialData()
    }
  } catch {
    authenticated.value = false
  }
})

onUnmounted(stopDocumentPolling)
</script>

<style scoped>
.control {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.24);
  padding: 0 0.9rem;
  color: white;
  outline: none;
  transition:
    border-color 180ms ease,
    background-color 180ms ease;
}

.control:focus {
  border-color: rgba(94, 234, 212, 0.72);
  background: rgba(0, 0, 0, 0.38);
}

input.control,
select.control {
  height: 2.75rem;
}

.primary-button,
.ghost-button,
.icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    color 180ms ease,
    opacity 180ms ease;
}

.primary-button {
  height: 2.75rem;
  background: #5eead4;
  padding: 0 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: #061015;
}

.primary-button:hover {
  background: #99f6e4;
}

.primary-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.ghost-button {
  height: 2.5rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0 0.8rem;
  color: #cbd5e1;
}

.ghost-button:hover {
  border-color: rgba(94, 234, 212, 0.38);
  background: rgba(94, 234, 212, 0.08);
  color: #ccfbf1;
}

.icon-button {
  height: 2rem;
  width: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
}

.icon-button:hover {
  border-color: rgba(94, 234, 212, 0.38);
  background: rgba(94, 234, 212, 0.08);
  color: #ccfbf1;
}

.status-pill {
  display: inline-flex;
  border-width: 1px;
  padding: 0.22rem 0.5rem;
  font-size: 0.72rem;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
