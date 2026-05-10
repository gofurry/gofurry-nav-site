<template>
  <div class="grid gap-6 xl:grid-cols-[minmax(0,420px)_minmax(0,1fr)]">
    <section class="space-y-6">
      <article class="panel overflow-hidden">
        <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-teal-300/60 to-transparent" />
        <div class="relative flex items-start justify-between gap-4">
          <div>
            <div class="flex items-center gap-3">
              <div class="grid h-11 w-11 place-items-center border border-teal-300/30 bg-teal-300/10 text-teal-200">
                <Sparkles :size="20" />
              </div>
              <div>
                <p class="text-xs uppercase tracking-[0.26em] text-teal-200/60">AI QA Console</p>
                <h3 class="mt-1 text-xl font-semibold text-white">腾讯云推理调试台</h3>
              </div>
            </div>
            <p class="mt-4 max-w-md text-sm leading-6 text-slate-400">
              这里直接走后端代理腾讯云 OpenAI-compatible 接口，适合在真实环境里调试检索、prompt 和流式输出。
            </p>
          </div>
          <span class="status-pill" :class="providerTone">{{ providerLabel }}</span>
        </div>

        <dl class="mt-6 grid gap-px overflow-hidden border border-white/10 bg-white/10 text-sm md:grid-cols-2">
          <div class="bg-black/20 p-4">
            <dt class="text-slate-500">Base URL</dt>
            <dd class="mt-2 break-all text-slate-100">{{ provider?.base_url || '-' }}</dd>
          </div>
          <div class="bg-black/20 p-4">
            <dt class="text-slate-500">Model</dt>
            <dd class="mt-2 text-slate-100">{{ provider?.model || '-' }}</dd>
          </div>
          <div class="bg-black/20 p-4">
            <dt class="text-slate-500">配置状态</dt>
            <dd class="mt-2 text-slate-100">{{ provider?.configured ? 'configured' : 'not configured' }}</dd>
          </div>
          <div class="bg-black/20 p-4">
            <dt class="text-slate-500">健康状态</dt>
            <dd class="mt-2 text-slate-100">{{ provider?.healthy ? 'healthy' : 'degraded' }}</dd>
          </div>
        </dl>
        <p v-if="healthError" class="mt-4 text-sm leading-6 text-rose-200">{{ healthError }}</p>
      </article>

      <form class="panel space-y-5" @submit.prevent="runQuery">
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-xs uppercase tracking-[0.24em] text-slate-500">Query</p>
            <h4 class="mt-1 text-lg font-semibold text-white">发起一次单轮问答</h4>
          </div>
          <span class="status-pill" :class="queryStatusTone">{{ queryStatusLabel }}</span>
        </div>

        <label class="block">
          <span class="label">问题</span>
          <textarea v-model="question" class="control min-h-36 resize-none py-3" placeholder="例如：GoFurry 现在能检索哪些资料？" />
        </label>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block">
            <span class="label">Top K</span>
            <input v-model="topKText" class="control" inputmode="numeric" placeholder="6" @input="sanitizeTopK" />
          </label>
          <label class="block">
            <span class="label">文档 ID</span>
            <input v-model="filters.documentIds" class="control" placeholder="1,2,3" />
          </label>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block">
            <span class="label">来源类型</span>
            <input v-model="filters.sourceType" class="control" placeholder="website, manual" />
          </label>
          <label class="block">
            <span class="label">分类</span>
            <input v-model="filters.category" class="control" placeholder="intro, faq" />
          </label>
        </div>

        <label class="block">
          <span class="label">语言</span>
          <input v-model="filters.language" class="control" placeholder="zh-CN, en-US" />
        </label>

        <div class="flex flex-wrap gap-3">
          <button class="primary-button" :disabled="busy" type="submit">
            <Send :size="16" />
            {{ busy ? '生成中' : '开始问答' }}
          </button>
          <button class="ghost-button" :disabled="!busy" type="button" @click="stopQuery">
            <Square :size="16" />
            停止
          </button>
          <button class="ghost-button" :disabled="busy" type="button" @click="resetPanel">
            <Trash2 :size="16" />
            清空
          </button>
        </div>

        <div class="rounded border border-white/10 bg-white/[0.03] p-4 text-sm leading-6 text-slate-400">
          <p class="font-medium text-slate-200">调试说明</p>
          <ul class="mt-2 space-y-1">
            <li>• 先检索，再调用腾讯云模型；没有 sources 时会直接返回拒答提示。</li>
            <li>• 流式输出会先显示状态，再逐步拼接回答内容。</li>
            <li>• 右侧会保留本次请求上下文、sources 和 token 统计。</li>
          </ul>
        </div>
      </form>

      <article class="panel">
        <div class="flex items-center gap-2 text-slate-300">
          <Clock3 :size="18" class="text-teal-200" />
          <span class="text-sm">运行轨迹</span>
        </div>
        <div v-if="timeline.length" class="mt-4 space-y-3">
          <div v-for="item in timeline" :key="item.at + item.stage" class="border border-white/10 bg-black/20 p-3">
            <div class="flex items-center justify-between gap-3">
              <strong class="text-sm text-white">{{ item.stage }}</strong>
              <span class="text-xs text-slate-500">{{ formatTime(item.at) }}</span>
            </div>
            <p class="mt-2 text-sm leading-6 text-slate-300">{{ item.message }}</p>
          </div>
        </div>
        <p v-else class="mt-4 text-sm text-slate-500">还没有开始新的请求。</p>
      </article>
    </section>

    <section class="space-y-6">
      <article class="panel">
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-2 text-slate-300">
            <MessageSquareText :size="18" class="text-teal-200" />
            <span class="text-sm">流式回答</span>
          </div>
          <span class="status-pill" :class="answerTone">{{ queryStatusLabel }}</span>
        </div>

        <div class="mt-4 min-h-[320px] border border-white/10 bg-black/20 p-5">
          <p v-if="answer" class="whitespace-pre-wrap text-sm leading-7 text-slate-100">{{ answer }}</p>
          <div v-else class="grid min-h-[280px] place-items-center text-center text-sm text-slate-500">
            <div>
              <BotMessageSquare :size="28" class="mx-auto mb-3 text-teal-200/80" />
              <p>等待输入问题并开始问答。</p>
            </div>
          </div>
        </div>

        <div class="mt-4 flex flex-wrap gap-3 text-xs text-slate-500">
          <span class="mini-chip">耗时 {{ formatDuration(elapsedMs) }}</span>
          <span class="mini-chip">模型 {{ response?.usage.answer_model || provider?.model || '-' }}</span>
          <span class="mini-chip">Top K {{ response?.usage.top_k ?? topKValue }}</span>
          <span class="mini-chip">sources {{ sources.length }}</span>
        </div>

        <div v-if="response" class="mt-4 grid gap-px overflow-hidden border border-white/10 bg-white/10 md:grid-cols-4">
          <div class="bg-black/20 p-4">
            <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Prompt</p>
            <p class="mt-3 text-2xl font-semibold text-white">{{ response.usage.prompt_tokens || 0 }}</p>
          </div>
          <div class="bg-black/20 p-4">
            <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Completion</p>
            <p class="mt-3 text-2xl font-semibold text-white">{{ response.usage.completion_tokens || 0 }}</p>
          </div>
          <div class="bg-black/20 p-4">
            <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Total</p>
            <p class="mt-3 text-2xl font-semibold text-white">{{ response.usage.total_tokens || 0 }}</p>
          </div>
          <div class="bg-black/20 p-4">
            <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Reasoning</p>
            <p class="mt-3 text-2xl font-semibold text-white">{{ response.usage.reasoning_tokens || 0 }}</p>
          </div>
        </div>
      </article>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <article class="panel">
          <div class="flex items-center gap-2 text-slate-300">
            <Database :size="18" class="text-teal-200" />
            <span class="text-sm">Sources</span>
          </div>
          <div v-if="sources.length" class="mt-4 space-y-3">
            <section v-for="(source, index) in sources" :key="source.chunk_id" class="border border-white/10 bg-black/20 p-4 transition hover:border-teal-300/30">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="mini-chip">#{{ index + 1 }}</span>
                    <strong class="text-sm text-white">{{ source.title || documentLabel(source.document_id) }}</strong>
                  </div>
                  <p class="mt-2 text-xs leading-5 text-slate-500">{{ sourceLine(source) }}</p>
                </div>
                <span class="text-xs text-teal-200/80">{{ source.score.toFixed(4) }}</span>
              </div>
              <a v-if="source.url" class="mt-3 block truncate text-xs text-teal-200/80 hover:text-teal-100" :href="source.url" rel="noreferrer" target="_blank">
                {{ source.url }}
              </a>
              <p class="mt-3 text-sm leading-6 text-slate-300">{{ source.content }}</p>
            </section>
          </div>
          <div v-else class="mt-4 grid min-h-[180px] place-items-center text-sm text-slate-500">
            还没有收到 sources。
          </div>
        </article>

        <aside class="space-y-6">
          <article class="panel">
            <div class="flex items-center gap-2 text-slate-300">
              <ShieldCheck :size="18" class="text-teal-200" />
              <span class="text-sm">请求上下文</span>
            </div>
            <pre class="mt-4 overflow-auto border border-white/10 bg-black/20 p-4 text-xs leading-6 text-slate-300">{{ requestPreview }}</pre>
          </article>

          <article class="panel">
            <div class="flex items-center gap-2 text-slate-300">
              <Activity :size="18" class="text-teal-200" />
              <span class="text-sm">错误与状态</span>
            </div>
            <p v-if="error" class="mt-4 rounded border border-rose-300/20 bg-rose-300/10 p-4 text-sm leading-6 text-rose-200">
              {{ error }}
            </p>
            <p v-else class="mt-4 text-sm leading-6 text-slate-500">
              请求结束后会在这里显示最终状态，适合快速判断是否是检索、模型还是网络层出的问题。
            </p>
          </article>
        </aside>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Activity, BotMessageSquare, Clock3, Database, MessageSquareText, Send, ShieldCheck, Sparkles, Square, Trash2 } from 'lucide-vue-next'
import { health, queryRagStream } from '../api'
import type { HealthInfo, QueryResponse, QuerySource } from '../types'

type QueryFilterFields = {
  sourceType: string
  category: string
  language: string
  documentIds: string
}

type TimelineItem = {
  stage: string
  message: string
  at: string
}

const question = ref('GoFurry 现在可以回答哪些问题？')
const topKText = ref('6')
const filters = ref<QueryFilterFields>({
  sourceType: '',
  category: '',
  language: '',
  documentIds: '',
})
const busy = ref(false)
const answer = ref('')
const error = ref('')
const elapsedMs = ref(0)
const response = ref<QueryResponse | null>(null)
const sources = ref<QuerySource[]>([])
const timeline = ref<TimelineItem[]>([])
const healthInfo = ref<HealthInfo | null>(null)
const healthError = ref('')
let startedAt = 0
let abortController: AbortController | null = null

const topKValue = computed(() => parsePositiveInt(topKText.value, 6))
const provider = computed(() => healthInfo.value?.tencent || null)
const providerTone = computed(() => {
  if (provider.value?.healthy) {
    return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
  }
  if (provider.value?.configured) {
    return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  }
  return 'border-slate-500/30 bg-slate-500/10 text-slate-200'
})
const providerLabel = computed(() => {
  if (provider.value?.healthy) return 'healthy'
  if (provider.value?.configured) return 'configured'
  return 'not configured'
})
const queryStatusLabel = computed(() => {
  if (busy.value) return 'streaming'
  if (error.value) return 'error'
  if (response.value) return 'done'
  return 'idle'
})
const queryStatusTone = computed(() => {
  if (busy.value) {
    return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  }
  if (error.value) {
    return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  }
  if (response.value) {
    return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
  }
  return 'border-slate-500/30 bg-slate-500/10 text-slate-200'
})
const answerTone = computed(() => queryStatusTone.value)
const requestPreview = computed(() =>
  JSON.stringify(
    {
      question: question.value.trim(),
      top_k: topKValue.value,
      filters: buildQueryFilters(),
    },
    null,
    2,
  ),
)

onMounted(() => {
  void refreshHealth()
})

async function refreshHealth() {
  healthError.value = ''
  try {
    healthInfo.value = await health()
  } catch (err) {
    healthError.value = (err as Error).message || '健康检查失败'
  }
}

async function runQuery() {
  const trimmedQuestion = question.value.trim()
  if (!trimmedQuestion) {
    error.value = '问题不能为空。'
    return
  }
  busy.value = true
  error.value = ''
  answer.value = ''
  response.value = null
  sources.value = []
  timeline.value = []
  elapsedMs.value = 0
  startedAt = Date.now()
  abortController?.abort()
  abortController = new AbortController()

  try {
    const result = await queryRagStream(
      trimmedQuestion,
      topKValue.value,
      buildQueryFilters(),
      {
        onStatus: (payload) => {
          timeline.value = [
            ...timeline.value,
            {
              stage: payload.stage,
              message: payload.message,
              at: new Date().toISOString(),
            },
          ]
          elapsedMs.value = Date.now() - startedAt
        },
        onSources: (items) => {
          sources.value = items
        },
        onDelta: (text) => {
          answer.value += text
          elapsedMs.value = Date.now() - startedAt
        },
        onDone: (payload) => {
          response.value = payload
          sources.value = payload.sources || sources.value
          answer.value = payload.answer || answer.value
          elapsedMs.value = Date.now() - startedAt
        },
        onError: (message) => {
          error.value = message
        },
      },
      abortController.signal,
    )
    response.value = result
    sources.value = result.sources || sources.value
    answer.value = result.answer || answer.value
    elapsedMs.value = Date.now() - startedAt
  } catch (err) {
    const message = (err as Error).message || '问答请求失败'
    if (message.toLowerCase().includes('aborted')) {
      error.value = '请求已停止。'
    } else {
      error.value = message
    }
  } finally {
    busy.value = false
    abortController = null
    void refreshHealth()
  }
}

function stopQuery() {
  abortController?.abort()
}

function resetPanel() {
  question.value = 'GoFurry 现在可以回答哪些问题？'
  topKText.value = '6'
  filters.value = {
    sourceType: '',
    category: '',
    language: '',
    documentIds: '',
  }
  busy.value = false
  answer.value = ''
  error.value = ''
  response.value = null
  sources.value = []
  timeline.value = []
  elapsedMs.value = 0
  abortController?.abort()
  abortController = null
}

function buildQueryFilters() {
  return {
    source_type: parseCSV(filters.value.sourceType),
    document_ids: parseIDs(filters.value.documentIds),
    category: parseCSV(filters.value.category),
    language: parseCSV(filters.value.language),
  }
}

function parseCSV(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function parseIDs(value: string) {
  return value
    .split(',')
    .map((item) => parsePositiveInt(item, 0))
    .filter((item) => item > 0)
}

function parsePositiveInt(value: string, fallback: number) {
  const parsed = Number.parseInt(value.replace(/\D/g, ''), 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

function sanitizeTopK(event: Event) {
  topKText.value = (event.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 3)
}

function sourceLine(source: QuerySource) {
  const pieces = [
    `doc #${source.document_id}`,
    `chunk #${source.chunk_index}`,
    `${source.token_count || 0} chars`,
    source.source_type,
    source.source_id,
  ].filter(Boolean)
  return pieces.join(' / ')
}

function documentLabel(documentId: number) {
  return `Document #${documentId}`
}

function formatDuration(value: number) {
  if (value < 1000) return `${value} ms`
  return `${(value / 1000).toFixed(1)} s`
}

function formatTime(value: string) {
  return new Date(value).toLocaleTimeString()
}
</script>

<style scoped>
.panel {
  position: relative;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.035);
  padding: 1.5rem;
}

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

input.control {
  height: 2.75rem;
}

textarea.control {
  min-height: 9rem;
}

.label {
  display: block;
  margin-bottom: 0.45rem;
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: rgba(148, 163, 184, 0.82);
}

.primary-button,
.ghost-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    color 180ms ease,
    opacity 180ms ease,
    transform 180ms ease;
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

.primary-button:disabled,
.ghost-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.status-pill,
.mini-chip {
  display: inline-flex;
  border-style: solid;
  border-width: 1px;
  padding: 0.22rem 0.5rem;
  font-size: 0.72rem;
}
</style>
