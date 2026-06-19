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
              这里直接走后端代理腾讯云 OpenAI-compatible 接口，适合在真实环境里调试检索、prompt、引用和流式输出。
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
        <div v-if="ollamaQueue" class="mt-4 border border-white/10 bg-black/20 p-4">
          <div class="flex items-center justify-between gap-3">
            <span class="text-sm text-slate-300">Ollama 队列</span>
            <span class="mini-chip border-teal-300/30 bg-teal-300/10 text-teal-100">
              {{ ollamaQueue.active }}/{{ ollamaQueue.max_concurrency }}
            </span>
          </div>
          <div class="mt-3 grid gap-3 text-sm md:grid-cols-3">
            <div>
              <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Query</p>
              <p class="mt-2 text-slate-100">{{ ollamaQueue.queued_query }}/{{ ollamaQueue.query_queue_size }}</p>
            </div>
            <div>
              <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Ingest</p>
              <p class="mt-2 text-slate-100">{{ ollamaQueue.queued_ingest }}/{{ ollamaQueue.ingest_queue_size }}</p>
            </div>
            <div>
              <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Oldest wait</p>
              <p class="mt-2 text-slate-100">{{ formatDuration(ollamaQueue.oldest_wait_ms) }}</p>
            </div>
          </div>
          <div class="mt-3 flex flex-wrap gap-2 text-xs text-slate-500">
            <span class="mini-chip">rejected {{ ollamaQueue.rejected }}</span>
            <span class="mini-chip">timeout {{ ollamaQueue.wait_timeout_seconds }} s</span>
          </div>
        </div>
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
          <textarea v-model="question" class="control min-h-36 resize-none py-3" placeholder="例如：gofurry 现在可以回答哪些问题？" />
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

        <label class="flex cursor-pointer items-start gap-3 border border-white/10 bg-black/20 px-4 py-3 text-sm text-slate-300 transition hover:border-teal-300/30">
          <input
            v-model="includeDetails"
            class="mt-1 h-4 w-4 rounded border-white/20 bg-black/40 text-teal-300 focus:ring-teal-300"
            type="checkbox"
          />
          <span>
            <strong class="block text-slate-100">包含引用详情</strong>
            <span class="mt-1 block text-xs leading-5 text-slate-500">
              勾选后会额外返回每个引用的完整文章、元数据和血缘，方便后续迁移到业务前端。
            </span>
          </span>
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
            <li>• 右侧会保留本次请求上下文、sources、引用详情和 token 统计。</li>
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
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-center gap-2 text-slate-300">
            <MessageSquareText :size="18" class="text-teal-200" />
            <span class="text-sm">流式回答</span>
          </div>
          <span class="status-pill" :class="answerTone">{{ queryStatusLabel }}</span>
        </div>

        <div class="mt-4 flex flex-wrap gap-3">
          <button class="ghost-button h-9" type="button" :disabled="!answerTextForCopy" @click="copyFinalAnswer">
            <Copy :size="15" />
            复制最终回答
          </button>
          <button class="ghost-button h-9" type="button" :disabled="!citationTextForCopy" @click="copyCitationPack">
            <Copy :size="15" />
            复制引用
          </button>
          <span v-if="copyFeedback" class="mini-chip border-teal-300/30 bg-teal-300/10 text-teal-100">{{ copyFeedback }}</span>
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
          <span class="mini-chip">引用 {{ citations.length }}</span>
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

      <article class="panel">
        <div class="flex items-center gap-2 text-slate-300">
          <Database :size="18" class="text-teal-200" />
          <span class="text-sm">Sources 调试</span>
        </div>
        <div v-if="sources.length" class="mt-4 space-y-3">
          <section
            v-for="(source, index) in sources"
            :key="source.chunk_id"
            class="border border-white/10 bg-black/20 p-4 transition hover:border-teal-300/30"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="mini-chip">#{{ index + 1 }}</span>
                  <strong class="truncate text-sm text-white">{{ source.title || documentLabel(source.document_id) }}</strong>
                  <span class="mini-chip" :class="sourcePromptTone(source)">{{ sourcePromptLabel(source) }}</span>
                </div>
                <p class="mt-2 text-xs leading-5 text-slate-500">{{ sourceLine(source) }}</p>
              </div>
              <span class="shrink-0 text-xs text-teal-200/80">{{ source.score.toFixed(4) }}</span>
            </div>
            <a
              v-if="source.url"
              class="mt-3 block truncate text-xs text-teal-200/80 hover:text-teal-100"
              :href="source.url"
              rel="noreferrer"
              target="_blank"
            >
              {{ source.url }}
            </a>
            <p class="mt-3 text-sm leading-6 text-slate-300">{{ source.content }}</p>
          </section>
        </div>
        <div v-else class="mt-4 grid min-h-[180px] place-items-center text-sm text-slate-500">还没有收到 sources。</div>
      </article>

      <article v-if="citations.length" class="panel">
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-2 text-slate-300">
            <BookOpen :size="18" class="text-teal-200" />
            <span class="text-sm">引用详情（默认折叠）</span>
          </div>
          <span class="status-pill border-teal-300/30 bg-teal-300/10 text-teal-100">{{ citations.length }} items</span>
        </div>

        <div class="mt-4 space-y-3">
          <details
            v-for="citation in citations"
            :key="citation.rank + '-' + citation.lineage.chunk_id"
            class="citation-details border border-white/10 bg-black/20 p-4 transition open:border-teal-300/30 open:bg-black/30"
          >
            <summary class="citation-summary">
              <div class="flex min-w-0 flex-1 items-start gap-3">
                <span class="mini-chip shrink-0">#{{ citation.rank }}</span>
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <strong class="truncate text-sm text-white">
                      {{ citation.document.title || citation.source.title || documentLabel(citation.lineage.document_id) }}
                    </strong>
                    <span class="mini-chip" :class="citation.used_in_prompt ? 'border-teal-300/30 bg-teal-300/10 text-teal-100' : 'border-amber-300/30 bg-amber-300/10 text-amber-100'">
                      {{ citation.used_in_prompt ? '已进入上下文' : '仅检索到' }}
                    </span>
                  </div>
                  <p class="mt-2 text-xs leading-5 text-slate-500">{{ citationLineageLabel(citation) }}</p>
                </div>
              </div>
              <span class="shrink-0 text-xs text-teal-200/80">{{ citation.source.score.toFixed(4) }}</span>
            </summary>

            <div class="mt-4 space-y-4">
              <div class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                <span class="mini-chip">文档 #{{ citation.document.id }}</span>
                <span class="mini-chip">Chunk #{{ citation.chunk.id }}</span>
                <span class="mini-chip">ChunkIndex {{ citation.chunk.chunk_index }}</span>
                <span class="mini-chip">{{ citation.chunk.token_count || citation.source.token_count }} tokens</span>
              </div>

              <div class="grid gap-3 md:grid-cols-2">
                <div class="rounded border border-white/10 bg-white/[0.03] p-3">
                  <p class="text-xs uppercase tracking-[0.18em] text-slate-500">血缘</p>
                  <div class="mt-3 grid gap-2 text-sm text-slate-200">
                    <p>来源：{{ citation.source.source_type || '-' }}{{ citation.source.source_id ? ` / ${citation.source.source_id}` : '' }}</p>
                    <p>文档状态：{{ citation.document.status || '-' }}</p>
                    <p>抓取/索引：{{ formatTime(citation.document.last_indexed_at || citation.document.processed_at || citation.document.created_at) }}</p>
                    <p>更新时间：{{ formatTime(citation.document.updated_at) }}</p>
                  </div>
                </div>
                <div class="rounded border border-white/10 bg-white/[0.03] p-3">
                  <p class="text-xs uppercase tracking-[0.18em] text-slate-500">元数据</p>
                  <pre class="debug-pre mt-3 text-xs leading-6 text-slate-300">{{ stringifyMetadata(citation.document.metadata) }}</pre>
                </div>
              </div>

              <div class="rounded border border-white/10 bg-white/[0.03] p-3">
                <p class="text-xs uppercase tracking-[0.18em] text-slate-500">完整文章</p>
                <pre class="debug-pre mt-3 text-sm leading-7 text-slate-200">{{ citation.document.content || '无内容' }}</pre>
              </div>

              <div class="rounded border border-white/10 bg-white/[0.03] p-3">
                <p class="text-xs uppercase tracking-[0.18em] text-slate-500">Chunk 原文</p>
                <pre class="debug-pre mt-3 text-sm leading-7 text-slate-200">{{ citation.chunk.content || '无内容' }}</pre>
              </div>
            </div>
          </details>
        </div>
      </article>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
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
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Activity, BookOpen, BotMessageSquare, Clock3, Copy, Database, MessageSquareText, Send, ShieldCheck, Sparkles, Square, Trash2 } from 'lucide-vue-next'
import { health, queryRagStream } from '../api'
import type { HealthInfo, QueryCitation, QueryResponse, QuerySource } from '../types'

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

const question = ref('gofurry 现在可以回答哪些问题？')
const topKText = ref('6')
const includeDetails = ref(false)
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
const citations = ref<QueryCitation[]>([])
const timeline = ref<TimelineItem[]>([])
const healthInfo = ref<HealthInfo | null>(null)
const healthError = ref('')
const copyFeedback = ref('')
let startedAt = 0
let abortController: AbortController | null = null
let copyFeedbackTimer: number | undefined

const topKValue = computed(() => parsePositiveInt(topKText.value, 6))
const provider = computed(() => healthInfo.value?.tencent || null)
const ollamaQueue = computed(() => healthInfo.value?.ollama?.queue || null)
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
const answerTextForCopy = computed(() => response.value?.answer?.trim() || answer.value.trim())
const citationTextForCopy = computed(() => {
  if (citations.value.length) {
    return buildCitationClipboardText(citations.value)
  }
  return buildSourceClipboardText(sources.value)
})
const requestPreview = computed(() =>
  JSON.stringify(
    {
      question: question.value.trim(),
      top_k: topKValue.value,
      include_details: includeDetails.value,
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
  citations.value = []
  timeline.value = []
  elapsedMs.value = 0
  copyFeedback.value = ''
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
          citations.value = payload.citations || []
          answer.value = payload.answer || answer.value
          elapsedMs.value = Date.now() - startedAt
        },
        onError: (message) => {
          error.value = message
        },
      },
      abortController.signal,
      includeDetails.value,
    )
    response.value = result
    sources.value = result.sources || sources.value
    citations.value = result.citations || citations.value
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
  question.value = 'gofurry 现在可以回答哪些问题？'
  topKText.value = '6'
  includeDetails.value = false
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
  citations.value = []
  timeline.value = []
  elapsedMs.value = 0
  copyFeedback.value = ''
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
    `${source.token_count || 0} tokens`,
    source.source_type,
    source.source_id,
  ].filter(Boolean)
  return pieces.join(' / ')
}

function sourcePromptLabel(source: QuerySource) {
  const citation = citations.value.find((item) => item.lineage.chunk_id === source.chunk_id)
  if (!citation) return '仅检索'
  return citation.used_in_prompt ? '已进上下文' : '仅检索'
}

function sourcePromptTone(source: QuerySource) {
  const citation = citations.value.find((item) => item.lineage.chunk_id === source.chunk_id)
  if (!citation) {
    return 'border-slate-500/30 bg-slate-500/10 text-slate-200'
  }
  return citation.used_in_prompt
    ? 'border-teal-300/30 bg-teal-300/10 text-teal-100'
    : 'border-amber-300/30 bg-amber-300/10 text-amber-100'
}

function citationLineageLabel(citation: QueryCitation) {
  const pieces = [
    `doc #${citation.lineage.document_id}`,
    `chunk #${citation.lineage.chunk_id}`,
    `chunk index ${citation.lineage.chunk_index}`,
    citation.lineage.source_type,
    citation.lineage.source_id,
  ].filter(Boolean)
  return pieces.join(' / ')
}

function stringifyMetadata(value: unknown) {
  if (value == null) {
    return '{}'
  }
  if (typeof value === 'string') {
    return value.trim() || '{}'
  }
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function buildSourceClipboardText(items: QuerySource[]) {
  if (!items.length) {
    return ''
  }
  const lines: string[] = ['引用摘要：']
  items.forEach((source, index) => {
    lines.push(`- [${index + 1}] ${source.title || documentLabel(source.document_id)}`)
    lines.push(`  - ${sourceLine(source)}`)
    lines.push(`  - score: ${source.score.toFixed(4)}`)
    if (source.url) {
      lines.push(`  - url: ${source.url}`)
    }
    if (source.content) {
      lines.push(`  - content: ${source.content}`)
    }
  })
  return lines.join('\n').trim()
}

function buildCitationClipboardText(items: QueryCitation[]) {
  if (!items.length) {
    return buildSourceClipboardText(sources.value)
  }
  const lines: string[] = ['引用详情：']
  items.forEach((citation) => {
    lines.push(`- [${citation.rank}] ${citation.document.title || citation.source.title || documentLabel(citation.lineage.document_id)}`)
    lines.push(`  - 状态: ${citation.used_in_prompt ? '已进入上下文' : '仅检索到'}`)
    lines.push(`  - 血缘: ${citationLineageLabel(citation)}`)
    lines.push(`  - 元数据: ${stringifyMetadata(citation.document.metadata)}`)
    lines.push(`  - 完整文章: ${citation.document.content || '无内容'}`)
    lines.push(`  - Chunk 原文: ${citation.chunk.content || '无内容'}`)
  })
  return lines.join('\n').trim()
}

async function copyFinalAnswer() {
  const text = answerTextForCopy.value.trim()
  if (!text) {
    setCopyFeedback('没有可复制的回答')
    return
  }
  try {
    await copyText(text)
    setCopyFeedback('已复制最终回答')
  } catch {
    setCopyFeedback('复制失败')
  }
}

async function copyCitationPack() {
  const text = citationTextForCopy.value.trim()
  if (!text) {
    setCopyFeedback('没有可复制的引用')
    return
  }
  try {
    await copyText(text)
    setCopyFeedback('已复制引用')
  } catch {
    setCopyFeedback('复制失败')
  }
}

async function copyText(text: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

function setCopyFeedback(message: string) {
  copyFeedback.value = message
  if (copyFeedbackTimer !== undefined) {
    window.clearTimeout(copyFeedbackTimer)
  }
  copyFeedbackTimer = window.setTimeout(() => {
    copyFeedback.value = ''
  }, 1800)
}

function documentLabel(documentId: number) {
  return `Document #${documentId}`
}

function formatDuration(value: number) {
  if (value < 1000) return `${value} ms`
  return `${(value / 1000).toFixed(1)} s`
}

function formatTime(value: string) {
  if (!value) return '-'
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

.citation-details {
  border-radius: 0;
}

.citation-summary {
  display: flex;
  cursor: pointer;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  list-style: none;
}

.citation-summary::-webkit-details-marker {
  display: none;
}

.debug-pre {
  max-height: 19rem;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
