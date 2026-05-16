<template>
  <div class="archive-shell">
    <aside class="archive-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-top">
        <button class="icon-button sidebar-toggle" type="button" :title="t('archive.actions.toggleSidebar')" @click="sidebarCollapsed = !sidebarCollapsed">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <button class="new-chat-button" type="button" @click="startNewChat">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 5v14M5 12h14" />
          </svg>
          <span>{{ t('archive.actions.newChat') }}</span>
        </button>
      </div>

      <div class="search-wrap">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="m21 21-4.4-4.4M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
        </svg>
        <input v-model.trim="searchText" type="search" :placeholder="t('archive.searchPlaceholder')" />
      </div>

      <div class="history-list">
        <button
          v-for="(item, index) in pagedRecords"
          :key="item.id"
          class="history-item"
          :class="{ active: item.id === activeId }"
          type="button"
          @click="openRecord(item.id)"
        >
          <span class="history-index">{{ historyNumber(index) }}</span>
          <span class="history-title">{{ item.question || t('archive.untitledQuestion') }}</span>
          <span class="history-meta">{{ formatTime(item.createdAt) }}</span>
        </button>
        <div v-if="pagedRecords.length === 0" class="history-empty">
          {{ t('archive.noMatchedQuestions') }}
        </div>
      </div>

      <div class="pager">
        <button type="button" :title="t('archive.pagination.prev')" :disabled="page <= 1" @click="page -= 1">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="m6 15 6-6 6 6" />
          </svg>
          <span>{{ t('archive.pagination.prev') }}</span>
        </button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button type="button" :title="t('archive.pagination.next')" :disabled="page >= totalPages" @click="page += 1">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="m6 9 6 6 6-6" />
          </svg>
          <span>{{ t('archive.pagination.next') }}</span>
        </button>
      </div>
    </aside>

    <main class="archive-workspace">
      <header class="workspace-header">
        <div class="workspace-actions">
          <NuxtLink class="icon-button home-button" to="/nav" :title="t('archive.actions.backHome')">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M15 18l-6-6 6-6" />
            </svg>
          </NuxtLink>
          <button class="icon-button doc-button" type="button" :title="t('archive.actions.openGuide')" @click="showGuide = !showGuide">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M7 3h7l4 4v14H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z" />
              <path d="M14 3v5h5M8 13h8M8 17h6" />
            </svg>
          </button>
        </div>
        <div class="workspace-title" :title="titleText">{{ titleText }}</div>
        <div class="queue-pill" :class="{ muted: !queueStatus }">
          <span class="pulse-dot"></span>
          <span>{{ queueText }}</span>
        </div>
      </header>

      <section ref="workspaceBodyRef" class="workspace-body" @scroll="handleWorkspaceScroll">
        <div class="workspace-overlay">
          <section class="rag-flyout" :class="{ expanded: ragInfoExpanded }">
            <div class="rag-flyout-panel">
              <div class="rag-flyout-summary">{{ ragSummaryText }}</div>
              <div class="rag-flyout-list">
                <div class="rag-flyout-row">
                  <span>{{ t('archive.rag.embeddingModel') }}</span>
                  <strong :title="ragInfo?.embedding_model || '-'">{{ ragInfo?.embedding_model || '-' }}</strong>
                </div>
                <div class="rag-flyout-row">
                  <span>{{ t('archive.rag.answerModel') }}</span>
                  <strong :title="ragInfo?.answer_model || '-'">{{ ragInfo?.answer_model || '-' }}</strong>
                </div>
              </div>
            </div>
            <button
              type="button"
              class="rag-flyout-toggle"
              :aria-expanded="ragInfoExpanded"
              :title="ragInfoExpanded ? t('archive.rag.collapse') : t('archive.rag.expand')"
              :aria-label="ragInfoExpanded ? t('archive.rag.collapse') : t('archive.rag.expand')"
              @click="ragInfoExpanded = !ragInfoExpanded"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path v-if="ragInfoExpanded" d="m9 6 6 6-6 6" />
                <path v-else d="m15 6-6 6 6 6" />
              </svg>
            </button>
          </section>
        </div>
        <section class="answer-panel">
          <Transition name="conversation-fade" mode="out-in">
            <div v-if="!activeRecord" key="empty" class="empty-conversation">
              <p class="eyebrow">{{ t('archive.empty.eyebrow') }}</p>
              <h1>{{ t('archive.empty.title') }}</h1>
              <p>{{ t('archive.empty.description') }}</p>
            </div>

            <div v-else :key="activeRecord.id" class="conversation">
              <div class="response-column">
                <div class="question-block">
                  <div class="block-title">
                    <span>{{ t('archive.sections.question') }}</span>
                    <button class="copy-button" type="button" @click="copyText(activeRecord.question)">{{ t('archive.actions.copy') }}</button>
                  </div>
                  <h1>{{ activeRecord.question }}</h1>
                </div>

                <div class="answer-block" :class="{ streaming: activeRecord.status === 'streaming' }">
                  <div class="answer-heading">
                    <span>{{ t('archive.sections.answer') }}</span>
                    <div class="block-actions">
                      <span v-if="activeRecord.status === 'streaming'" class="typing-state">{{ t('archive.status.streaming') }}</span>
                      <span v-else-if="activeRecord.status === 'error'" class="error-state">{{ t('archive.status.requestFailed') }}</span>
                      <button class="copy-button" type="button" :disabled="!activeRecord.answer" @click="copyText(activeRecord.answer)">{{ t('archive.actions.copy') }}</button>
                    </div>
                  </div>
                  <p v-if="activeRecord.answer" class="answer-text">{{ activeRecord.answer }}</p>
                  <div v-else class="answer-placeholder">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                  <p v-if="activeRecord.error" class="error-message">{{ activeRecord.error }}</p>
                </div>
              </div>

              <div v-if="activeRecord.sources.length" class="sources-block">
                <div class="sources-title">
                  <span>{{ t('archive.sections.sources') }}</span>
                  <span>{{ t('archive.sources.count', { count: activeRecord.sources.length }) }}</span>
                </div>
                <details v-for="(source, index) in activeRecord.sources" :key="sourceKey(source, index)" class="source-item">
                  <summary>
                    <span class="source-rank">[{{ index + 1 }}]</span>
                    <span class="source-name">{{ source.title || t('archive.sources.document', { id: source.document_id }) }}</span>
                    <span class="source-score">{{ scoreText(source.score) }}</span>
                  </summary>
                  <div class="source-content">
                    <div class="source-meta">
                      <span>{{ source.source_type || 'unknown' }}</span>
                      <span v-if="source.document_id">{{ t('archive.sources.document', { id: source.document_id }) }}</span>
                      <span v-if="typeof source.chunk_index === 'number'">{{ t('archive.sources.chunk', { index: source.chunk_index }) }}</span>
                    </div>
                    <button class="copy-button source-copy-button" type="button" :disabled="!sourceText(source)" @click="copyText(sourceText(source))">{{ t('archive.actions.copySourceContent') }}</button>
                    <p>{{ sourceText(source) || t('archive.sources.noContent') }}</p>
                    <a v-if="source.url" :href="source.url" target="_blank" rel="noopener noreferrer">{{ t('archive.sources.open') }}</a>
                  </div>
                </details>
              </div>
            </div>
          </Transition>
        </section>
      </section>

      <button
        v-if="hasScrollableWorkspace"
        type="button"
        class="scroll-dock"
        :style="{ '--scroll-progress': `${scrollProgressLabel}%` }"
        :title="t('archive.actions.backToTop')"
        :aria-label="t('archive.actions.backToTop')"
        @click="scrollWorkspaceToTop()"
      >
        <div class="scroll-progress">
          <span>{{ scrollProgressLabel }}%</span>
        </div>
      </button>

      <form class="ask-bar" @submit.prevent="submitQuestion">
        <input
          v-model="draftQuestion"
          :disabled="isStreaming"
          type="text"
          :placeholder="t('archive.askPlaceholder')"
        />
        <button v-if="isStreaming" type="button" class="send-button stop" @click="stopStream">
          {{ t('archive.actions.stop') }}
        </button>
        <button v-else type="submit" class="send-button" :disabled="!draftQuestion.trim()">
          {{ t('archive.actions.send') }}
        </button>
      </form>

      <transition name="modal">
        <div v-if="showGuide" class="guide-modal-backdrop" @click.self="showGuide = false">
          <article class="guide-panel" role="dialog" aria-modal="true" aria-labelledby="archive-guide-title">
            <div class="guide-header">
              <div>
                <p class="eyebrow">{{ t('archive.guide.eyebrow') }}</p>
                <h2 id="archive-guide-title">{{ t('archive.guide.title') }}</h2>
              </div>
              <button class="icon-button guide-close" type="button" :title="t('archive.actions.closeGuide')" @click="showGuide = false">
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M6 6l12 12M18 6 6 18" />
                </svg>
              </button>
            </div>
            <p>{{ t('archive.guide.description') }}</p>
            <div class="guide-grid">
              <section>
                <h3>{{ t('archive.guide.usageTitle') }}</h3>
                <p>{{ t('archive.guide.usageDesc') }}</p>
              </section>
              <section>
                <h3>{{ t('archive.guide.boundaryTitle') }}</h3>
                <p>{{ t('archive.guide.boundaryDesc') }}</p>
              </section>
              <section>
                <h3>{{ t('archive.guide.noticeTitle') }}</h3>
                <p>{{ t('archive.guide.noticeDesc') }}</p>
              </section>
            </div>
          </article>
        </div>
      </transition>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'

definePageMeta({
  ssr: false
})

type RagSource = {
  document_id?: number
  chunk_id?: number
  source_type?: string
  source_id?: string
  title?: string
  url?: string
  chunk_index?: number
  token_count?: number
  score?: number
  content?: string
  snippet?: string
}

type QueueStatus = {
  max_concurrency: number
  active: number
  queued_query: number
  queued_ingest: number
  rejected: number
  oldest_wait_ms: number
}

type ChatLimits = {
  public_query_max_question_runes: number
  public_query_max_top_k: number
  public_query_rate_limit_requests: number
  public_query_rate_limit_window_seconds: number
}

type RagInfo = {
  embedding_model?: string
  answer_model?: string
  document_total?: number
  chunk_total?: number
}

type ChatRecord = {
  id: string
  question: string
  answer: string
  sources: RagSource[]
  createdAt: number
  updatedAt: number
  status: 'idle' | 'streaming' | 'done' | 'error'
  error?: string
}

const storageKey = 'gofurry.archive.chat.records.v1'
const maxRecords = 50
const pageSize = 10
const { t } = useI18n()
const langStore = useLangStore()

const records = ref<ChatRecord[]>([])
const activeId = ref<string | null>(null)
const draftQuestion = ref('')
const searchText = ref('')
const page = ref(1)
const sidebarCollapsed = ref(false)
const showGuide = ref(false)
const queueStatus = ref<QueueStatus | null>(null)
const chatLimits = ref<ChatLimits | null>(null)
const ragInfo = ref<RagInfo | null>(null)
const ragInfoExpanded = ref(false)
const streamController = ref<AbortController | null>(null)
const workspaceBodyRef = ref<HTMLElement | null>(null)
const scrollProgress = ref(0)
const hasScrollableWorkspace = ref(false)
let queueTimer: number | null = null
let workspaceMetricsFrame: number | null = null

const filteredRecords = computed(() => {
  const keyword = searchText.value.trim().toLowerCase()
  if (!keyword) {
    return records.value
  }
  return records.value.filter(item => item.question.toLowerCase().includes(keyword) || item.answer.toLowerCase().includes(keyword))
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredRecords.value.length / pageSize)))

const pagedRecords = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredRecords.value.slice(start, start + pageSize)
})

const activeRecord = computed(() => records.value.find(item => item.id === activeId.value) || null)
const isStreaming = computed(() => Boolean(activeRecord.value?.status === 'streaming'))
const scrollProgressLabel = computed(() => Math.round(scrollProgress.value))

const titleText = computed(() => {
  return activeRecord.value?.question || t('archive.actions.newChat')
})

const queueText = computed(() => {
  if (!queueStatus.value) {
    return t('archive.queue.unknown')
  }
  const active = queueStatus.value.active ?? 0
  const max = queueStatus.value.max_concurrency ?? 0
  const queued = (queueStatus.value.queued_query ?? 0) + (queueStatus.value.queued_ingest ?? 0)
  return t('archive.queue.status', { active, max, queued })
})

const ragSummaryText = computed(() => {
  if (!ragInfo.value) {
    return t('archive.rag.summaryEmpty')
  }
  return t('archive.rag.summary', {
    documents: ragMetricText(ragInfo.value.document_total),
    chunks: ragMetricText(ragInfo.value.chunk_total)
  })
})

watch(searchText, () => {
  page.value = 1
})

watch(records, persistRecords, { deep: true })

onMounted(() => {
  loadRecords()
  refreshQueue()
  window.addEventListener('keydown', handleGuideKeydown)
  window.addEventListener('resize', scheduleWorkspaceMetrics)
  queueTimer = window.setInterval(refreshQueue, 8000)
  scheduleWorkspaceMetrics()
})

onUnmounted(() => {
  stopStream()
  window.removeEventListener('keydown', handleGuideKeydown)
  window.removeEventListener('resize', scheduleWorkspaceMetrics)
  if (workspaceMetricsFrame) {
    window.cancelAnimationFrame(workspaceMetricsFrame)
  }
  if (queueTimer) {
    window.clearInterval(queueTimer)
  }
})

function loadRecords() {
  try {
    const raw = window.localStorage.getItem(storageKey)
    const parsed = raw ? JSON.parse(raw) : []
    records.value = Array.isArray(parsed)
      ? parsed.slice(0, maxRecords).map(normalizeStoredRecord)
      : []
  } catch {
    records.value = []
  }
  activeId.value = null
  scheduleWorkspaceMetrics()
}

function normalizeStoredRecord(record: ChatRecord): ChatRecord {
  if (record.status !== 'streaming') {
    return record
  }
  return {
    ...record,
    status: record.answer ? 'done' : 'idle'
  }
}

function persistRecords() {
  if (!import.meta.client) {
    return
  }
  const snapshot = records.value.slice(0, maxRecords)
  window.localStorage.setItem(storageKey, JSON.stringify(snapshot))
}

function startNewChat() {
  stopStream()
  activeId.value = null
  draftQuestion.value = ''
  showGuide.value = false
  scrollWorkspaceToTop('smooth')
  nextTick(() => {
    document.querySelector<HTMLInputElement>('.ask-bar input')?.focus()
  })
}

function openRecord(id: string) {
  stopStream()
  activeId.value = id
  draftQuestion.value = ''
  showGuide.value = false
  nextTick(() => {
    scheduleWorkspaceMetrics()
  })
}

async function submitQuestion() {
  const question = draftQuestion.value.trim()
  if (!question || isStreaming.value) {
    return
  }
  const maxRunes = chatLimits.value?.public_query_max_question_runes ?? 0
  if (maxRunes > 0 && runeLength(question) > maxRunes) {
    const record: ChatRecord = {
      id: crypto.randomUUID(),
      question,
      answer: '',
      sources: [],
      createdAt: Date.now(),
      updatedAt: Date.now(),
      status: 'error',
      error: t('archive.errors.questionTooLong', { limit: maxRunes })
    }
    records.value = [record, ...records.value.filter(item => item.id !== record.id)].slice(0, maxRecords)
    activeId.value = record.id
    return
  }

  const record: ChatRecord = {
    id: crypto.randomUUID(),
    question,
    answer: '',
    sources: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
    status: 'streaming'
  }

  records.value = [record, ...records.value.filter(item => item.id !== record.id)].slice(0, maxRecords)
  activeId.value = record.id
  draftQuestion.value = ''
  showGuide.value = false
  await streamAnswer(record)
}

async function streamAnswer(record: ChatRecord) {
  const controller = new AbortController()
  streamController.value = controller
  await refreshQueue()

  try {
    const response = await fetch('/api/rag/chat/stream', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        question: record.question,
        top_k: 6,
        include_details: false
      }),
      signal: controller.signal
    })

    if (!response.ok || !response.body) {
      throw new Error(await responseErrorMessage(response))
    }

    await readSseStream(response.body, {
      onSources: (sources) => {
        record.sources = sources
        touchRecord(record)
      },
      onDelta: (text) => {
        record.answer += text
        touchRecord(record)
      },
      onDone: (payload) => {
        if (typeof payload?.answer === 'string') {
          record.answer = payload.answer
        }
        if (Array.isArray(payload?.sources)) {
          record.sources = payload.sources
        }
        record.status = 'done'
        touchRecord(record)
      },
      onError: (message) => {
        record.status = 'error'
        record.error = message
        touchRecord(record)
      }
    })

    if (record.status === 'streaming') {
      record.status = 'done'
      touchRecord(record)
    }
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      record.status = record.answer ? 'done' : 'idle'
      touchRecord(record)
      return
    }
    record.status = 'error'
    record.error = userFacingError(error?.message || t('archive.errors.serviceUnavailable'))
    touchRecord(record)
  } finally {
    if (streamController.value === controller) {
      streamController.value = null
    }
    await refreshQueue()
  }
}

function stopStream() {
  streamController.value?.abort()
  streamController.value = null
}

function touchRecord(record: ChatRecord) {
  record.updatedAt = Date.now()
  records.value = records.value.map(item => item.id === record.id ? { ...record, sources: [...record.sources] } : item)
  scheduleWorkspaceMetrics()
}

async function refreshQueue() {
  try {
    const response: any = await $fetch('/api/rag/chat/status')
    queueStatus.value = response?.data?.queue ?? null
    chatLimits.value = response?.data?.limits ?? null
    ragInfo.value = response?.data?.rag ?? null
  } catch {
    queueStatus.value = null
    chatLimits.value = null
    ragInfo.value = null
  }
}

async function readSseStream(
  body: ReadableStream<Uint8Array>,
  handlers: {
    onSources: (sources: RagSource[]) => void
    onDelta: (text: string) => void
    onDone: (payload: any) => void
    onError: (message: string) => void
  }
) {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }
    buffer += decoder.decode(value, { stream: true })
    const frames = buffer.split(/\r?\n\r?\n/)
    buffer = frames.pop() || ''
    frames.forEach(frame => handleSseFrame(frame, handlers))
  }
  buffer += decoder.decode()
  if (buffer.trim()) {
    handleSseFrame(buffer, handlers)
  }
}

function handleSseFrame(
  frame: string,
  handlers: {
    onSources: (sources: RagSource[]) => void
    onDelta: (text: string) => void
    onDone: (payload: any) => void
    onError: (message: string) => void
  }
) {
  const lines = frame.split(/\r?\n/)
  const event = lines.find(line => line.startsWith('event:'))?.slice(6).trim() || 'message'
  const data = lines
    .filter(line => line.startsWith('data:'))
    .map(line => line.slice(5).trim())
    .join('\n')

  if (!data) {
    return
  }

  let payload: any = {}
  try {
    payload = JSON.parse(data)
  } catch {
    payload = { text: data }
  }

  if (event === 'sources' && Array.isArray(payload.sources)) {
    handlers.onSources(payload.sources)
  } else if (event === 'delta') {
    handlers.onDelta(String(payload.text || ''))
  } else if (event === 'done') {
    handlers.onDone(payload)
  } else if (event === 'error') {
    handlers.onError(String(payload.message || t('archive.errors.serviceReturnedError')))
  }
}

function formatTime(value: number) {
  return new Intl.DateTimeFormat(langStore.lang === 'zh' ? 'zh-CN' : 'en-US', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(value)
}

function historyNumber(index: number) {
  return String((page.value - 1) * pageSize + index + 1).padStart(2, '0')
}

function sourceKey(source: RagSource, index: number) {
  return `${source.document_id ?? 'public'}-${source.chunk_id ?? source.title ?? source.url ?? 'source'}-${index}`
}

function scoreText(score?: number) {
  return typeof score === 'number' ? score.toFixed(3) : '0.000'
}

function sourceText(source: RagSource) {
  return source.snippet || source.content || ''
}

function ragMetricText(value?: number) {
  return typeof value === 'number' && Number.isFinite(value)
    ? new Intl.NumberFormat(langStore.lang === 'zh' ? 'zh-CN' : 'en-US').format(value)
    : '-'
}

function runeLength(value: string) {
  return Array.from(value).length
}

async function responseErrorMessage(response: Response) {
  try {
    const payload = await response.json()
    const message = payload?.message || payload?.statusMessage
    if (message) {
      return `${response.status}: ${message}`
    }
  } catch {
    // ignore malformed error body and fall back to status mapping
  }
  return String(response.status)
}

function userFacingError(message: string) {
  if (message.includes('429') || message.toLowerCase().includes('too many public chat requests')) {
    const windowSeconds = chatLimits.value?.public_query_rate_limit_window_seconds ?? 60
    return t('archive.errors.rateLimited', { seconds: windowSeconds })
  }
  if (message.includes('400')) {
    return t('archive.errors.invalidRequest')
  }
  if (message.includes('502') || message.includes('503') || message.toLowerCase().includes('unable to reach')) {
    return t('archive.errors.serviceUnavailable')
  }
  return message
}

function handleWorkspaceScroll() {
  updateWorkspaceMetrics()
}

function scrollWorkspaceToTop(behavior: ScrollBehavior = 'smooth') {
  workspaceBodyRef.value?.scrollTo({ top: 0, behavior })
  if (behavior === 'auto') {
    updateWorkspaceMetrics()
    return
  }
  window.setTimeout(updateWorkspaceMetrics, 220)
}

function scheduleWorkspaceMetrics() {
  if (!import.meta.client || workspaceMetricsFrame) {
    return
  }
  workspaceMetricsFrame = window.requestAnimationFrame(() => {
    workspaceMetricsFrame = null
    updateWorkspaceMetrics()
  })
}

function updateWorkspaceMetrics() {
  const body = workspaceBodyRef.value
  if (!body) {
    hasScrollableWorkspace.value = false
    scrollProgress.value = 0
    return
  }
  const maxScroll = Math.max(body.scrollHeight - body.clientHeight, 0)
  hasScrollableWorkspace.value = maxScroll > 24
  if (maxScroll <= 0) {
    scrollProgress.value = 0
    return
  }
  scrollProgress.value = Math.min(100, Math.max(0, (body.scrollTop / maxScroll) * 100))
}

function handleGuideKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && showGuide.value) {
    showGuide.value = false
  }
}

async function copyText(text: string) {
  const content = text.trim()
  if (!content || !import.meta.client) {
    return
  }
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(content)
      return
    }
    copyTextFallback(content)
  } catch {
    copyTextFallback(content)
  }
}

function copyTextFallback(content: string) {
  const textarea = document.createElement('textarea')
  textarea.value = content
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

</script>

<style scoped>
.archive-shell {
  --control-bar-height: 56px;
  --motion-duration: 0.5s;

  display: flex;
  height: 100vh;
  max-height: 100vh;
  min-height: 0;
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(14, 20, 21, 0.96), rgba(25, 37, 34, 0.96)),
    repeating-linear-gradient(90deg, rgba(255,255,255,0.025) 0, rgba(255,255,255,0.025) 1px, transparent 1px, transparent 42px);
  color: #ecf4ee;
}

.archive-sidebar {
  display: flex;
  flex: 0 0 320px;
  min-width: 0;
  flex-direction: column;
  border-right: 1px solid rgba(182, 214, 190, 0.14);
  background: rgba(11, 16, 17, 0.92);
  contain: layout paint style;
}

.archive-sidebar.collapsed {
  flex-basis: 74px;
}

.sidebar-top {
  display: flex;
  gap: 10px;
  padding: 12px 16px;
}

.archive-sidebar.collapsed .sidebar-top {
  flex-direction: column;
}

.icon-button {
  display: inline-flex;
  height: 40px;
  width: 40px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.06);
  color: #f3f7f2;
  transition: background var(--motion-duration) ease, border-color var(--motion-duration) ease, transform var(--motion-duration) ease;
}

.icon-button:hover,
.new-chat-button:hover,
.send-button:hover {
  background: rgba(172, 211, 153, 0.18);
  transform: translateY(-1px);
}

.icon-button svg,
.new-chat-button svg,
.search-wrap svg,
.home-button svg {
  width: 18px;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.new-chat-button {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid rgba(185, 218, 167, 0.28);
  border-radius: 8px;
  background: rgba(113, 151, 95, 0.22);
  color: #f4faef;
  font-size: 14px;
  font-weight: 700;
  transition: background var(--motion-duration) ease, transform var(--motion-duration) ease;
}

.new-chat-button span {
  white-space: nowrap;
}

.archive-sidebar.collapsed .new-chat-button span,
.archive-sidebar.collapsed .search-wrap,
.archive-sidebar.collapsed .history-title,
.archive-sidebar.collapsed .history-meta {
  display: none;
}

.archive-sidebar.collapsed .new-chat-button {
  flex: 0 0 40px;
  width: 40px;
}

.search-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 16px 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  padding: 0 12px;
  color: rgba(236, 244, 238, 0.72);
}

.search-wrap input {
  min-width: 0;
  width: 100%;
  border: 0;
  background: transparent;
  color: #f7fbf6;
  outline: none;
  padding: 12px 0;
}

.history-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  -ms-overflow-style: none;
  scrollbar-width: none;
  padding: 6px 12px 14px;
}

.history-list::-webkit-scrollbar {
  display: none;
}

.history-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  width: 100%;
  gap: 6px 10px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: rgba(241, 248, 242, 0.78);
  margin-bottom: 8px;
  padding: 12px;
  text-align: left;
  transition: background var(--motion-duration) ease, border-color var(--motion-duration) ease;
}

.archive-sidebar.collapsed .history-item {
  grid-template-columns: 1fr;
  justify-items: center;
  margin-bottom: 10px;
  padding: 10px 0;
}

.history-item:hover,
.history-item.active {
  border-color: rgba(187, 218, 169, 0.22);
  background: rgba(255, 255, 255, 0.075);
}

.history-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  font-weight: 650;
}

.history-index {
  grid-row: span 2;
  align-self: center;
  color: rgba(174, 222, 155, 0.9);
  font-size: 12px;
  font-weight: 900;
}

.history-meta,
.history-empty {
  color: rgba(228, 237, 224, 0.48);
  font-size: 12px;
}

.history-empty {
  padding: 18px 12px;
}

.pager {
  display: flex;
  height: var(--control-bar-height);
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0 16px;
  color: rgba(234, 241, 231, 0.62);
  font-size: 12px;
}

.archive-sidebar.collapsed .pager {
  height: var(--control-bar-height);
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 4px 0;
}

.pager button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.07);
  color: inherit;
  padding: 7px 10px;
  transition: background var(--motion-duration) ease, opacity var(--motion-duration) ease;
}

.pager button:disabled {
  cursor: not-allowed;
  opacity: 0.36;
}

.pager button svg {
  display: none;
  width: 18px;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.archive-sidebar.collapsed .pager button {
  height: 22px;
  width: 40px;
  padding: 0;
}

.archive-sidebar.collapsed .pager button span,
.archive-sidebar.collapsed .pager > span {
  display: none;
}

.archive-sidebar.collapsed .pager button svg {
  display: block;
}

.archive-workspace {
  position: relative;
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  overflow: hidden;
  isolation: isolate;
  background:
    radial-gradient(circle at 24% 16%, rgba(182, 230, 157, 0.12), transparent 30%),
    radial-gradient(circle at 84% 72%, rgba(100, 167, 142, 0.13), transparent 34%),
    linear-gradient(145deg, rgba(8, 14, 15, 0.98), rgba(18, 28, 27, 0.96));
}

.archive-workspace::before,
.archive-workspace::after {
  content: "";
  position: absolute;
  pointer-events: none;
  z-index: -1;
}

.archive-workspace::before {
  inset: -24%;
  background:
    radial-gradient(circle at 18% 22%, rgba(194, 244, 165, 0.24), transparent 26%),
    radial-gradient(circle at 68% 18%, rgba(108, 206, 177, 0.18), transparent 24%),
    radial-gradient(circle at 74% 78%, rgba(238, 185, 111, 0.13), transparent 28%);
  filter: blur(34px);
  opacity: 0.76;
  transform: translate3d(-2%, -1%, 0) scale(1.02);
  animation: archiveGlowDrift 18s ease-in-out infinite alternate;
}

.archive-workspace::after {
  inset: 0;
  background:
    linear-gradient(115deg, transparent 0%, rgba(255, 255, 255, 0.045) 42%, transparent 68%),
    repeating-linear-gradient(90deg, rgba(255, 255, 255, 0.018) 0, rgba(255, 255, 255, 0.018) 1px, transparent 1px, transparent 48px),
    repeating-linear-gradient(0deg, rgba(255, 255, 255, 0.014) 0, rgba(255, 255, 255, 0.014) 1px, transparent 1px, transparent 48px);
  mask-image: radial-gradient(circle at 58% 46%, black, transparent 82%);
  opacity: 0.58;
  animation: archiveMistSweep 24s ease-in-out infinite alternate;
}

.workspace-header {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) max-content;
  align-items: center;
  gap: 14px;
  min-height: 48px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(8, 13, 14, 0.58);
  padding: 9px 18px;
}

.workspace-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workspace-actions .icon-button {
  border-color: transparent;
  background: transparent;
}

.workspace-actions .icon-button:hover {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.08);
  transform: none;
}

.workspace-title {
  justify-self: center;
  max-width: min(720px, 70vw);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #f5faf3;
  font-size: 15px;
  font-weight: 800;
}

.queue-pill {
  display: inline-flex;
  min-width: max-content;
  align-items: center;
  gap: 8px;
  color: #d9f1d2;
  font-size: 12px;
  font-weight: 800;
  white-space: nowrap;
}

.queue-pill.muted {
  color: rgba(239, 244, 237, 0.55);
}

.pulse-dot {
  height: 7px;
  width: 7px;
  border-radius: 999px;
  background: #9bd67d;
  box-shadow: 0 0 0 5px rgba(155, 214, 125, 0.12);
}

.workspace-overlay {
  position: sticky;
  top: 14px;
  z-index: 5;
  display: flex;
  justify-content: flex-end;
  height: 0;
  pointer-events: none;
}

.rag-flyout {
  display: inline-flex;
  align-items: flex-start;
  justify-content: flex-end;
  pointer-events: auto;
}

.rag-flyout-panel {
  max-width: 0;
  border: 1px solid rgba(190, 229, 172, 0.14);
  border-right: 0;
  border-radius: 16px 0 0 16px;
  background:
    linear-gradient(145deg, rgba(19, 27, 26, 0.94), rgba(11, 17, 18, 0.92)),
    radial-gradient(circle at 12% 12%, rgba(174, 222, 155, 0.1), transparent 36%);
  box-shadow: 0 12px 34px rgba(0, 0, 0, 0.24);
  margin-right: 0;
  opacity: 0;
  overflow: hidden;
  padding: 0;
  transform: translateX(10px) scale(0.98);
  transition:
    max-width 0.3s ease,
    margin-right 0.3s ease,
    opacity 0.22s ease,
    padding 0.3s ease,
    transform 0.3s ease;
}

.rag-flyout.expanded .rag-flyout-panel {
  max-width: min(228px, calc(100vw - 88px));
  margin-right: 10px;
  opacity: 1;
  padding: 12px 14px;
  transform: translateX(0) scale(1);
}

.rag-flyout-summary {
  color: rgba(233, 242, 229, 0.72);
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  margin-bottom: 12px;
}

.rag-flyout-list {
  display: grid;
  gap: 12px;
}

.rag-flyout-row {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.rag-flyout-row span {
  color: rgba(222, 234, 218, 0.58);
  font-size: 11px;
  font-weight: 700;
}

.rag-flyout-row strong {
  overflow: hidden;
  color: #f8fcf7;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rag-flyout-toggle {
  display: inline-flex;
  height: 44px;
  width: 44px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(190, 229, 172, 0.18);
  border-radius: 999px;
  background: linear-gradient(145deg, rgba(24, 33, 31, 0.96), rgba(12, 18, 19, 0.94));
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.24);
  color: #eff8ea;
  margin-top: 10px;
  transition: transform var(--motion-duration) ease, background var(--motion-duration) ease, border-color var(--motion-duration) ease;
}

.rag-flyout.expanded .rag-flyout-toggle {
  margin-top: 10px;
  border-radius: 0 999px 999px 0;
}

.rag-flyout-toggle:hover {
  background: linear-gradient(145deg, rgba(39, 54, 49, 0.98), rgba(18, 27, 26, 0.96));
  border-color: rgba(190, 229, 172, 0.28);
  transform: translateX(-1px);
}

.rag-flyout-toggle svg {
  height: 18px;
  width: 18px;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.workspace-body {
  position: relative;
  z-index: 1;
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  -ms-overflow-style: none;
  scrollbar-width: none;
  padding: 26px clamp(18px, 4vw, 54px);
}

.workspace-body::-webkit-scrollbar {
  display: none;
}

.answer-panel {
  margin: 0 auto;
  width: min(1280px, 100%);
}

.scroll-dock {
  position: absolute;
  right: clamp(14px, 2.8vw, 28px);
  bottom: calc(var(--control-bar-height) + 18px);
  z-index: 4;
  display: inline-grid;
  place-items: center;
  border: 0;
  border: 1px solid rgba(187, 226, 169, 0.18);
  border-radius: 999px;
  background: linear-gradient(145deg, rgba(13, 20, 19, 0.88), rgba(25, 36, 33, 0.8));
  box-shadow: 0 18px 42px rgba(0, 0, 0, 0.34);
  backdrop-filter: blur(18px) saturate(140%);
  cursor: pointer;
  padding: 8px;
  transition: transform var(--motion-duration) ease, background var(--motion-duration) ease, border-color var(--motion-duration) ease;
}

.scroll-dock:hover {
  background: linear-gradient(145deg, rgba(21, 33, 30, 0.92), rgba(34, 50, 45, 0.84));
  border-color: rgba(187, 226, 169, 0.28);
  transform: translateY(-1px);
}

.scroll-progress {
  position: relative;
  display: grid;
  height: 50px;
  width: 50px;
  place-items: center;
  border-radius: 999px;
  background:
    conic-gradient(from 210deg, rgba(173, 234, 149, 0.96) var(--scroll-progress), rgba(255, 255, 255, 0.08) 0),
    radial-gradient(circle at 50% 45%, rgba(227, 255, 215, 0.18), transparent 68%);
  color: #f7fff3;
  font-size: 12px;
  font-weight: 900;
}

.scroll-progress::before {
  content: "";
  position: absolute;
  inset: 4px;
  border-radius: inherit;
  background: linear-gradient(180deg, rgba(7, 12, 12, 0.94), rgba(17, 25, 23, 0.92));
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.05);
}

.scroll-progress span {
  position: relative;
  z-index: 1;
}

.guide-panel {
  border: 1px solid rgba(204, 222, 194, 0.16);
  border-radius: 14px;
  background:
    linear-gradient(145deg, rgba(20, 31, 29, 0.94), rgba(10, 15, 16, 0.9)),
    radial-gradient(circle at 12% 8%, rgba(174, 222, 155, 0.14), transparent 34%);
  box-shadow: 0 24px 90px rgba(0, 0, 0, 0.48);
  max-height: min(78vh, 720px);
  overflow-y: auto;
  padding: clamp(18px, 3vw, 28px);
  width: min(820px, calc(100vw - 32px));
}

.guide-modal-backdrop {
  position: fixed;
  z-index: 20;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(2, 6, 7, 0.54);
  backdrop-filter: blur(14px);
  padding: 18px;
}

.guide-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 16px;
}

.guide-header .eyebrow {
  margin-bottom: 8px;
}

.guide-close {
  border-color: rgba(255, 255, 255, 0.14);
}

.guide-panel h2,
.empty-conversation h1,
.question-block h1 {
  margin: 0;
  letter-spacing: 0;
}

.guide-panel p,
.empty-conversation p {
  color: rgba(235, 242, 232, 0.7);
  line-height: 1.8;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 18px;
}

.guide-grid section {
  border-left: 2px solid rgba(155, 214, 125, 0.42);
  padding-left: 14px;
}

.guide-grid h3,
.sources-title,
.answer-heading,
.block-title {
  color: #aede9b;
  font-size: 13px;
  font-weight: 800;
}

.empty-conversation {
  display: grid;
  min-height: 48vh;
  align-content: center;
  gap: 14px;
}

.eyebrow {
  margin: 0;
  color: #aede9b;
  font-size: 13px;
  font-weight: 800;
}

.empty-conversation h1 {
  max-width: 760px;
  color: #f7fbf4;
  font-size: clamp(34px, 5vw, 68px);
  font-weight: 900;
  line-height: 1.06;
}

.conversation {
  display: grid;
  gap: 18px;
}

.response-column {
  display: grid;
  align-content: start;
  gap: 18px;
}

@media (min-width: 1280px) {
  .conversation {
    grid-template-columns: minmax(0, 1fr) minmax(340px, 420px);
    align-items: start;
  }

  .response-column {
    grid-column: 1;
  }

  .sources-block {
    grid-column: 2;
  }
}

.question-block,
.answer-block,
.sources-block {
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  padding: clamp(16px, 3vw, 24px);
}

.question-block h1 {
  margin-top: 10px;
  color: #fffaf0;
  font-size: clamp(22px, 3vw, 34px);
  line-height: 1.3;
}

.answer-block.streaming {
  outline: 1px solid rgba(155, 214, 125, 0.2);
}

.answer-heading,
.sources-title,
.block-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.block-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.copy-button {
  border: 1px solid rgba(190, 222, 174, 0.22);
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.055);
  color: rgba(242, 250, 238, 0.78);
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
  padding: 7px 9px;
  transition: background var(--motion-duration) ease, color var(--motion-duration) ease, transform var(--motion-duration) ease, opacity var(--motion-duration) ease;
}

.copy-button:hover:not(:disabled) {
  background: rgba(174, 222, 155, 0.16);
  color: #f8fff4;
  transform: translateY(-1px);
}

.copy-button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.typing-state {
  color: rgba(255, 214, 146, 0.9);
}

.error-state,
.error-message {
  color: #ffb7a8;
}

.answer-text {
  margin: 14px 0 0;
  white-space: pre-wrap;
  color: rgba(248, 252, 246, 0.9);
  font-size: 16px;
  line-height: 1.9;
}

.answer-placeholder {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.answer-placeholder span {
  display: block;
  height: 13px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(255,255,255,0.05), rgba(185,218,169,0.18), rgba(255,255,255,0.05));
  animation: shimmer 1.25s ease-in-out infinite;
}

.answer-placeholder span:nth-child(2) {
  width: 82%;
}

.answer-placeholder span:nth-child(3) {
  width: 58%;
}

.source-item {
  margin-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(241, 248, 238, 0.82);
}

.source-item summary {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  cursor: pointer;
  padding: 14px 0;
}

.source-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.source-rank,
.source-score {
  color: rgba(174, 222, 155, 0.9);
  font-size: 12px;
  font-weight: 800;
}

.source-content {
  padding: 0 0 16px 28px;
}

.source-copy-button {
  margin: 0 0 10px;
}

.source-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.source-meta span {
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: rgba(235, 241, 231, 0.62);
  padding: 5px 9px;
  font-size: 12px;
}

.source-content p {
  white-space: pre-wrap;
  color: rgba(244, 250, 242, 0.78);
  line-height: 1.8;
}

.source-content a {
  color: #bde99f;
  font-size: 13px;
  font-weight: 800;
}

.ask-bar {
  position: relative;
  z-index: 1;
  display: flex;
  height: var(--control-bar-height);
  align-items: center;
  gap: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(8, 13, 14, 0.72);
  padding: 8px clamp(16px, 4vw, 42px);
}

.ask-bar input {
  height: 40px;
  min-width: 0;
  flex: 1;
  border: 1px solid rgba(204, 222, 194, 0.16);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.075);
  color: #f8fcf6;
  outline: none;
  padding: 0 14px;
}

.send-button {
  height: 40px;
  min-height: 40px;
  border: 1px solid rgba(185, 218, 167, 0.28);
  border-radius: 8px;
  background: rgba(128, 164, 93, 0.32);
  color: #f9fff4;
  font-weight: 800;
  padding: 0 20px;
  transition: background var(--motion-duration) ease, transform var(--motion-duration) ease, opacity var(--motion-duration) ease;
}

.send-button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.send-button.stop {
  border-color: rgba(255, 183, 168, 0.36);
  background: rgba(135, 74, 61, 0.38);
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity var(--motion-duration) ease;
}

.modal-enter-active .guide-panel,
.modal-leave-active .guide-panel {
  transition: opacity var(--motion-duration) ease, transform var(--motion-duration) ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .guide-panel,
.modal-leave-to .guide-panel {
  opacity: 0;
  transform: translateY(-8px);
}

.conversation-fade-enter-active,
.conversation-fade-leave-active {
  transition: opacity var(--motion-duration) ease, transform var(--motion-duration) ease;
}

.conversation-fade-enter-from,
.conversation-fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

@keyframes shimmer {
  0% {
    opacity: 0.45;
    transform: translateX(-4px);
  }
  50% {
    opacity: 1;
    transform: translateX(4px);
  }
  100% {
    opacity: 0.45;
    transform: translateX(-4px);
  }
}

@keyframes archiveGlowDrift {
  0% {
    transform: translate3d(-2%, -1%, 0) scale(1.02) rotate(0deg);
  }
  50% {
    transform: translate3d(3%, 2%, 0) scale(1.08) rotate(5deg);
  }
  100% {
    transform: translate3d(-1%, 4%, 0) scale(1.04) rotate(-4deg);
  }
}

@keyframes archiveMistSweep {
  0% {
    transform: translate3d(-4%, 0, 0);
    opacity: 0.42;
  }
  100% {
    transform: translate3d(4%, 0, 0);
    opacity: 0.66;
  }
}

@media (prefers-reduced-motion: reduce) {
  .archive-workspace::before,
  .archive-workspace::after,
  .answer-placeholder span {
    animation: none;
  }
}

@media (max-width: 820px) {
  .archive-shell {
    height: 100vh;
  }

  .archive-sidebar {
    flex-basis: 74px;
  }

  .archive-sidebar:not(.collapsed) {
    flex-basis: 280px;
  }

  .workspace-header {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .workspace-title {
    justify-self: start;
    max-width: 48vw;
  }

  .workspace-overlay {
    display: none;
  }

  .guide-grid {
    grid-template-columns: 1fr;
  }

  .queue-pill {
    max-width: 100%;
  }

  .workspace-body {
    padding: 18px 14px;
  }

  .scroll-dock {
    display: none;
  }

  .ask-bar {
    gap: 8px;
    padding-right: 10px;
    padding-left: 10px;
  }

  .send-button {
    padding: 0 14px;
  }
}
</style>
