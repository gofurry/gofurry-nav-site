<template>
  <div class="archive-page archive-shell">
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
        <input v-model.trim="searchText" type="text" :placeholder="t('archive.searchPlaceholder')" />
        <button
          v-if="searchText"
          class="search-clear-button"
          type="button"
          :title="t('archive.actions.clearSearch')"
          :aria-label="t('archive.actions.clearSearch')"
          @click="searchText = ''"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M7 7l10 10M17 7 7 17" />
          </svg>
        </button>
      </div>

      <div class="history-list">
        <div
          v-for="(item, index) in pagedSessions"
          :key="item.id"
          class="history-row"
          :class="{ active: item.id === activeSessionId }"
        >
          <button class="history-item" type="button" @click="openSession(item.id)">
            <span class="history-index">{{ historyNumber(index) }}</span>
            <span class="history-title">{{ sessionTitle(item) }}</span>
            <span class="history-meta">{{ sessionMeta(item) }}</span>
          </button>
          <button
            class="history-delete-button"
            type="button"
            :title="t('archive.actions.deleteSession')"
            :aria-label="t('archive.actions.deleteSession')"
            @click="requestDeleteSession(item.id)"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M8 7h8M10 7V5h4v2M9 10v7M15 10v7M6 7l1 13h10l1-13" />
            </svg>
          </button>
        </div>
        <div v-if="pagedSessions.length === 0" class="history-empty">
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
          <NuxtLink class="icon-button home-button" to="/" :title="t('archive.actions.backHome')">
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
            <div v-if="!activeSession" key="empty" class="empty-conversation">
              <p class="eyebrow">{{ t('archive.empty.eyebrow') }}</p>
              <h1>{{ t('archive.empty.title') }}</h1>
              <p>{{ t('archive.empty.description') }}</p>
            </div>

            <div v-else :key="activeSession.id" class="conversation-list">
              <article v-for="message in activeSession.messages" :key="message.id" class="conversation">
                <div class="response-column">
                  <div class="question-block">
                    <div class="block-title">
                      <span>{{ t('archive.sections.question') }}</span>
                      <button class="copy-button" type="button" @click="copyText(message.question)">{{ t('archive.actions.copy') }}</button>
                    </div>
                    <h1>{{ message.question }}</h1>
                  </div>

                  <div class="answer-block" :class="{ streaming: message.status === 'streaming' }">
                    <div class="answer-heading">
                      <span>{{ t('archive.sections.answer') }}</span>
                      <div class="block-actions">
                        <span v-if="message.status === 'streaming'" class="typing-state">{{ t('archive.status.streaming') }}</span>
                        <span v-else-if="message.status === 'error'" class="error-state">{{ t('archive.status.requestFailed') }}</span>
                        <button class="copy-button" type="button" :disabled="!message.answer || !message.citations.length" @click="copyAnswerWithCitations(message)">{{ t('archive.actions.copyWithSources') }}</button>
                        <button class="copy-button" type="button" :disabled="!message.answer" @click="copyText(message.answer)">{{ t('archive.actions.copy') }}</button>
                      </div>
                    </div>
                    <MdPreview
                      v-if="message.answer"
                      class="answer-text answer-markdown"
                      :editor-id="`archive-answer-${message.id}`"
                      :model-value="message.answer"
                      preview-theme="vuepress"
                      code-theme="github"
                    />
                    <div
                      v-else
                      class="answer-placeholder"
                      :class="{ 'answer-placeholder--active': message.status === 'streaming' }"
                    >
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                    <p v-if="message.error" class="error-message">{{ message.error }}</p>
                  </div>
                </div>

                <div v-if="message.citations.length" class="citations-block">
                  <div class="citations-title">
                    <span>{{ t('archive.sections.citations') }}</span>
                    <span>{{ t('archive.citations.count', { count: message.citations.length }) }}</span>
                  </div>
                  <details
                    v-for="(citation, index) in message.citations"
                    :key="citationKey(citation, index)"
                    class="citation-item"
                    @toggle="handleCitationToggle($event, citation, index)"
                  >
                    <summary>
                      <span class="citation-rank">[{{ index + 1 }}]</span>
                      <span class="citation-head">
                        <span class="citation-name">{{ citation.title || t('archive.citations.untitled', { index: index + 1 }) }}</span>
                      </span>
                      <span class="citation-score">{{ scoreText(citation.score) }}</span>
                    </summary>
                    <div class="citation-content">
                      <div class="citation-meta">
                        <span v-if="citation.source_type" class="citation-type">{{ citation.source_type }}</span>
                        <span v-if="typeof citation.chunk_index === 'number'">{{ t('archive.citations.chunk', { index: citation.chunk_index }) }}</span>
                      </div>
                      <button class="copy-button citation-copy-button" type="button" :disabled="!citationText(citation)" @click="copyText(citationText(citation))">{{ t('archive.actions.copyCitationSnippet') }}</button>
                      <MdPreview
                        v-if="citationText(citation) && isCitationOpen(citation, index)"
                        class="citation-markdown"
                        :editor-id="`archive-citation-${message.id}-${index}`"
                        :model-value="citationText(citation)"
                        preview-theme="vuepress"
                        code-theme="github"
                      />
                      <p v-else-if="!citationText(citation)" class="citation-empty">{{ t('archive.citations.noContent') }}</p>
                      <a v-if="citation.url" :href="citation.url" target="_blank" rel="noopener noreferrer">{{ t('archive.citations.open') }}</a>
                    </div>
                  </details>
                </div>
              </article>
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

      <transition name="modal">
        <div v-if="deleteSessionTarget" class="guide-modal-backdrop" @click.self="closeDeleteSessionDialog">
          <article class="delete-dialog" role="dialog" aria-modal="true" aria-labelledby="delete-session-title">
            <div class="delete-dialog-icon">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M8 7h8M10 7V5h4v2M9 10v7M15 10v7M6 7l1 13h10l1-13" />
              </svg>
            </div>
            <div class="delete-dialog-copy">
              <p class="eyebrow">{{ t('archive.deleteSession.eyebrow') }}</p>
              <h2 id="delete-session-title">{{ t('archive.deleteSession.title') }}</h2>
              <p>{{ t('archive.deleteSession.description', { title: sessionTitle(deleteSessionTarget) }) }}</p>
            </div>
            <div class="delete-dialog-actions">
              <button type="button" class="delete-dialog-cancel" @click="closeDeleteSessionDialog">
                {{ t('archive.actions.cancel') }}
              </button>
              <button type="button" class="delete-dialog-confirm" @click="confirmDeleteSession">
                {{ t('archive.actions.confirmDelete') }}
              </button>
            </div>
          </article>
        </div>
      </transition>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { useThemeStore } from '@/stores/theme'

const MdPreview = defineAsyncComponent(async () => {
  await import('md-editor-v3/lib/preview.css')
  const module = await import('md-editor-v3')
  return module.MdPreview
})

definePageMeta({
  ssr: false
})

useHead({
  meta: [
    { name: 'robots', content: 'noindex, follow' }
  ]
})

type ArchiveCitation = {
  source_type?: string
  title?: string
  url?: string
  chunk_index?: number
  score?: number
  snippet?: string
}

type LegacyArchiveSource = ArchiveCitation & {
  document_id?: number
  chunk_id?: number
  source_id?: string
  token_count?: number
  content?: string
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
  public_query_context_max_turns: number
  public_query_context_max_runes: number
}

type RagInfo = {
  embedding_model?: string
  answer_model?: string
  document_total?: number
  chunk_total?: number
}

type ChatMessage = {
  id: string
  question: string
  answer: string
  citations: ArchiveCitation[]
  createdAt: number
  updatedAt: number
  status: 'idle' | 'streaming' | 'done' | 'error'
  error?: string
}

type ChatSession = {
  id: string
  title: string
  messages: ChatMessage[]
  createdAt: number
  updatedAt: number
}

type StoredChatSession = Partial<ChatSession> & {
  messages?: Partial<ChatMessage>[]
}

type StoredChatRecord = Partial<ChatMessage> & {
  sources?: LegacyArchiveSource[]
}

const storageKey = 'gofurry.archive.chat.sessions.v1'
const legacyStorageKey = 'gofurry.archive.chat.records.v1'
const maxSessions = 50
const maxTurnsPerSession = 20
const defaultContextTurns = 3
const pageSize = 10
const streamDeltaFlushMs = 200
const { t } = useI18n()
const langStore = useLangStore()
const themeStore = useThemeStore()
const route = useRoute()

const sessions = ref<ChatSession[]>([])
const activeSessionId = ref<string | null>(null)
const deleteSessionId = ref<string | null>(null)
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
const openCitationKeys = ref<Record<string, boolean>>({})
const workspaceBodyRef = ref<HTMLElement | null>(null)
const scrollProgress = ref(0)
const hasScrollableWorkspace = ref(false)
let queueTimer: number | null = null
let workspaceMetricsFrame: number | null = null
let lastAppliedRoutePrompt = ''

const filteredSessions = computed(() => {
  const keyword = searchText.value.trim().toLowerCase()
  if (!keyword) {
    return sessions.value
  }
  return sessions.value.filter(session => {
    const title = sessionTitle(session).toLowerCase()
    return title.includes(keyword) || session.messages.some(message =>
      message.question.toLowerCase().includes(keyword) || message.answer.toLowerCase().includes(keyword)
    )
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredSessions.value.length / pageSize)))

const pagedSessions = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredSessions.value.slice(start, start + pageSize)
})

const activeSession = computed(() => sessions.value.find(item => item.id === activeSessionId.value) || null)
const deleteSessionTarget = computed(() => sessions.value.find(item => item.id === deleteSessionId.value) || null)
const isStreaming = computed(() => Boolean(activeSession.value?.messages.some(message => message.status === 'streaming')))
const scrollProgressLabel = computed(() => Math.round(scrollProgress.value))
const contextTurnsLimit = computed(() => chatLimits.value?.public_query_context_max_turns ?? defaultContextTurns)

const titleText = computed(() => {
  return activeSession.value ? sessionTitle(activeSession.value) : t('archive.actions.newChat')
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
    chunks: ragMetricText(ragInfo.value.chunk_total),
    memory: ragMetricText(contextTurnsLimit.value)
  })
})

watch(searchText, () => {
  page.value = 1
})

watch(sessions, persistSessions, { deep: true })

watch(
  () => route.query.q,
  () => {
    applyRoutePrompt()
  }
)

onMounted(() => {
  themeStore.initTheme()
  loadSessions()
  applyRoutePrompt()
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

function loadSessions() {
  try {
    const raw = window.localStorage.getItem(storageKey)
    const parsed = raw ? JSON.parse(raw) : []
    sessions.value = Array.isArray(parsed)
      ? parsed.slice(0, maxSessions).map(normalizeStoredSession).filter(session => session.messages.length > 0)
      : []
    window.localStorage.removeItem(legacyStorageKey)
  } catch {
    sessions.value = []
  }
  activeSessionId.value = null
  scheduleWorkspaceMetrics()
}

function normalizeStoredSession(session: StoredChatSession): ChatSession {
  const createdAt = typeof session?.createdAt === 'number' ? session.createdAt : Date.now()
  const messages = Array.isArray(session?.messages)
    ? session.messages.slice(0, maxTurnsPerSession).map(normalizeStoredMessage).filter(message => message.question || message.answer)
    : []
  const lastMessage = messages.length ? messages[messages.length - 1] : null
  const updatedAt = typeof session?.updatedAt === 'number'
    ? session.updatedAt
    : lastMessage?.updatedAt ?? createdAt
  return {
    id: typeof session?.id === 'string' && session.id ? session.id : crypto.randomUUID(),
    title: typeof session?.title === 'string' ? session.title : '',
    messages,
    createdAt,
    updatedAt
  }
}

function normalizeStoredMessage(record: StoredChatRecord): ChatMessage {
  const createdAt = typeof record?.createdAt === 'number' ? record.createdAt : Date.now()
  const status = record?.status === 'streaming'
    ? record.answer ? 'done' : 'idle'
    : record?.status === 'done' || record?.status === 'error'
      ? record.status
      : 'idle'
  return {
    id: typeof record?.id === 'string' && record.id ? record.id : crypto.randomUUID(),
    question: typeof record?.question === 'string' ? record.question : '',
    answer: typeof record?.answer === 'string' ? record.answer : '',
    citations: normalizeArchiveCitations(record?.citations ?? record.sources),
    createdAt,
    updatedAt: typeof record?.updatedAt === 'number' ? record.updatedAt : createdAt,
    status,
    error: typeof record?.error === 'string' ? record.error : undefined
  }
}

function normalizeArchiveCitations(input: unknown): ArchiveCitation[] {
  if (!Array.isArray(input)) {
    return []
  }

  return input.map((item) => {
    const citation = item && typeof item === 'object' ? item as LegacyArchiveSource : {}
    return {
      source_type: typeof citation.source_type === 'string' ? citation.source_type : '',
      title: typeof citation.title === 'string' ? citation.title : '',
      url: typeof citation.url === 'string' ? citation.url : '',
      chunk_index: typeof citation.chunk_index === 'number' ? citation.chunk_index : undefined,
      score: typeof citation.score === 'number' ? citation.score : undefined,
      snippet: typeof citation.snippet === 'string'
        ? citation.snippet
        : typeof citation.content === 'string'
          ? citation.content
          : ''
    }
  })
}

function persistSessions() {
  if (!import.meta.client) {
    return
  }
  const snapshot = sessions.value.slice(0, maxSessions).map(session => ({
    ...session,
    messages: session.messages.slice(-maxTurnsPerSession)
  }))
  window.localStorage.setItem(storageKey, JSON.stringify(snapshot))
}

function startNewChat() {
  stopStream()
  activeSessionId.value = null
  draftQuestion.value = ''
  showGuide.value = false
  scrollWorkspaceToTop('smooth')
  nextTick(() => {
    document.querySelector<HTMLInputElement>('.ask-bar input')?.focus()
  })
}

function routePromptText() {
  const raw = route.query.q
  if (Array.isArray(raw)) {
    return raw[0]?.trim() || ''
  }
  return typeof raw === 'string' ? raw.trim() : ''
}

function applyRoutePrompt() {
  if (!import.meta.client) {
    return
  }
  const prompt = routePromptText()
  if (!prompt || prompt === lastAppliedRoutePrompt || isStreaming.value) {
    return
  }
  lastAppliedRoutePrompt = prompt
  activeSessionId.value = null
  draftQuestion.value = prompt
  showGuide.value = false
  nextTick(() => {
    document.querySelector<HTMLInputElement>('.ask-bar input')?.focus()
  })
}

function openSession(id: string) {
  stopStream()
  activeSessionId.value = id
  draftQuestion.value = ''
  showGuide.value = false
  nextTick(() => {
    scheduleWorkspaceMetrics()
  })
}

function requestDeleteSession(id: string) {
  deleteSessionId.value = id
}

function closeDeleteSessionDialog() {
  deleteSessionId.value = null
}

function confirmDeleteSession() {
  const id = deleteSessionId.value
  if (!id) {
    return
  }
  if (id === activeSessionId.value) {
    stopStream()
    activeSessionId.value = null
    draftQuestion.value = ''
  }
  sessions.value = sessions.value.filter(session => session.id !== id)
  deleteSessionId.value = null
  page.value = Math.min(page.value, Math.max(1, Math.ceil(filteredSessions.value.length / pageSize)))
  scheduleWorkspaceMetrics()
}

async function submitQuestion() {
  const question = draftQuestion.value.trim()
  if (!question || isStreaming.value) {
    return
  }
  const session = ensureActiveSession(question)
  const maxRunes = chatLimits.value?.public_query_max_question_runes ?? 0
  if (maxRunes > 0 && runeLength(question) > maxRunes) {
    const message: ChatMessage = {
      id: crypto.randomUUID(),
      question,
      answer: '',
      citations: [],
      createdAt: Date.now(),
      updatedAt: Date.now(),
      status: 'error',
      error: t('archive.errors.questionTooLong', { limit: maxRunes })
    }
    appendMessage(session, message)
    return
  }

  const message: ChatMessage = {
    id: crypto.randomUUID(),
    question,
    answer: '',
    citations: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
    status: 'streaming'
  }

  appendMessage(session, message)
  draftQuestion.value = ''
  showGuide.value = false
  await streamAnswer(session.id, message)
}

async function streamAnswer(sessionId: string, message: ChatMessage) {
  const controller = new AbortController()
  streamController.value = controller
  const context = buildRequestContext(sessionId, message.id)
  let pendingDelta = ''
  let deltaFlushTimer: number | null = null

  function clearDeltaFlushTimer() {
    if (deltaFlushTimer === null) {
      return
    }
    window.clearTimeout(deltaFlushTimer)
    deltaFlushTimer = null
  }

  function flushPendingDelta() {
    clearDeltaFlushTimer()
    if (!pendingDelta) {
      return
    }
    message.answer += pendingDelta
    pendingDelta = ''
    touchMessage(sessionId, message)
  }

  function scheduleDeltaFlush() {
    if (deltaFlushTimer !== null) {
      return
    }
    deltaFlushTimer = window.setTimeout(() => {
      deltaFlushTimer = null
      flushPendingDelta()
    }, streamDeltaFlushMs)
  }

  await refreshQueue()

  try {
    const response = await fetch('/api/rag/chat/stream', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        question: message.question,
        context,
        top_k: 6,
        include_details: false
      }),
      signal: controller.signal
    })

    if (!response.ok || !response.body) {
      throw new Error(await responseErrorMessage(response))
    }

    await readSseStream(response.body, {
      onCitations: (citations) => {
        message.citations = citations
        touchMessage(sessionId, message)
      },
      onDelta: (text) => {
        pendingDelta += text
        scheduleDeltaFlush()
      },
      onDone: (payload) => {
        if (typeof payload?.answer === 'string') {
          clearDeltaFlushTimer()
          pendingDelta = ''
          message.answer = payload.answer
        } else {
          flushPendingDelta()
        }
        message.citations = normalizeArchiveCitations(payload?.citations ?? payload?.sources)
        message.status = 'done'
        touchMessage(sessionId, message)
      },
      onError: (errorMessage) => {
        flushPendingDelta()
        message.status = 'error'
        message.error = errorMessage
        touchMessage(sessionId, message)
      }
    })

    if (message.status === 'streaming') {
      flushPendingDelta()
      message.status = 'done'
      touchMessage(sessionId, message)
    }
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      flushPendingDelta()
      message.status = message.answer ? 'done' : 'idle'
      touchMessage(sessionId, message)
      return
    }
    flushPendingDelta()
    message.status = 'error'
    message.error = userFacingError(error?.message || t('archive.errors.serviceUnavailable'))
    touchMessage(sessionId, message)
  } finally {
    clearDeltaFlushTimer()
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

function ensureActiveSession(question: string) {
  const active = activeSession.value
  if (active) {
    return active
  }
  const now = Date.now()
  const session: ChatSession = {
    id: crypto.randomUUID(),
    title: question,
    messages: [],
    createdAt: now,
    updatedAt: now
  }
  sessions.value = [session, ...sessions.value].slice(0, maxSessions)
  activeSessionId.value = session.id
  return session
}

function appendMessage(session: ChatSession, message: ChatMessage) {
  const nextSession = {
    ...session,
    title: session.title || message.question,
    messages: [...session.messages, message].slice(-maxTurnsPerSession),
    updatedAt: Date.now()
  }
  sessions.value = [nextSession, ...sessions.value.filter(item => item.id !== session.id)].slice(0, maxSessions)
  activeSessionId.value = nextSession.id
  scheduleWorkspaceMetrics()
}

function touchMessage(sessionId: string, message: ChatMessage) {
  message.updatedAt = Date.now()
  sessions.value = sessions.value.map((session) => {
    if (session.id !== sessionId) {
      return session
    }
    return {
      ...session,
      messages: session.messages.map(item => item.id === message.id ? { ...message, citations: [...message.citations] } : item),
      updatedAt: message.updatedAt
    }
  }).sort((left, right) => right.updatedAt - left.updatedAt).slice(0, maxSessions)
  scheduleWorkspaceMetrics()
}

function sessionTitle(session: ChatSession) {
  return session.title || session.messages[0]?.question || t('archive.untitledQuestion')
}

function sessionMeta(session: ChatSession) {
  return t('archive.session.meta', {
    count: session.messages.length,
    time: formatTime(session.updatedAt)
  })
}

type QueryContextTurn = {
  question: string
  answer: string
  citations: QueryContextCitation[]
}

type QueryContextCitation = {
  title?: string
  url?: string
  source_type?: string
  snippet?: string
  score?: number
  chunk_index?: number
}

function buildRequestContext(sessionId: string, currentMessageId: string): QueryContextTurn[] {
  const maxTurns = Math.max(0, contextTurnsLimit.value)
  const maxRunes = chatLimits.value?.public_query_context_max_runes ?? 8000
  if (maxTurns <= 0 || maxRunes <= 0) {
    return []
  }

  const session = sessions.value.find(item => item.id === sessionId)
  const turns = (session?.messages ?? [])
    .filter(item => item.id !== currentMessageId && item.status === 'done' && item.question.trim() && item.answer.trim())
    .slice(-maxTurns)
    .map(item => ({
      question: item.question.trim(),
      answer: item.answer.trim(),
      citations: item.citations.map(citation => ({
        title: citation.title,
        url: citation.url,
        source_type: citation.source_type,
        snippet: citation.snippet,
        score: citation.score,
        chunk_index: citation.chunk_index
      }))
    }))

  return trimContextTurns(turns, maxRunes)
}

function trimContextTurns(turns: QueryContextTurn[], maxRunes: number): QueryContextTurn[] {
  let remaining = maxRunes
  const result: QueryContextTurn[] = []

  for (const turn of [...turns].reverse()) {
    if (remaining <= 0) {
      break
    }
    const question = clipRunes(turn.question, remaining)
    remaining -= runeLength(question)
    const answer = clipRunes(turn.answer, remaining)
    remaining -= runeLength(answer)
    const citations = trimContextCitations(turn.citations, remaining)
    remaining -= citations.reduce((total, citation) => {
      return total
        + runeLength(citation.title || '')
        + runeLength(citation.url || '')
        + runeLength(citation.source_type || '')
        + runeLength(citation.snippet || '')
    }, 0)
    if (question || answer || citations.length) {
      result.unshift({ question, answer, citations })
    }
  }

  return result
}

function trimContextCitations(citations: QueryContextCitation[], maxRunes: number): QueryContextCitation[] {
  let remaining = maxRunes
  const result: QueryContextCitation[] = []
  for (const citation of citations) {
    if (remaining <= 0) {
      break
    }
    const title = clipRunes(citation.title || '', remaining)
    remaining -= runeLength(title)
    const url = clipRunes(citation.url || '', remaining)
    remaining -= runeLength(url)
    const sourceType = clipRunes(citation.source_type || '', remaining)
    remaining -= runeLength(sourceType)
    const snippet = clipRunes(citation.snippet || '', remaining)
    remaining -= runeLength(snippet)
    if (title || url || sourceType || snippet) {
      result.push({
        title,
        url,
        source_type: sourceType,
        snippet,
        score: citation.score,
        chunk_index: citation.chunk_index
      })
    }
  }
  return result
}

function clipRunes(value: string, limit: number) {
  if (limit <= 0) {
    return ''
  }
  const runes = Array.from(value)
  return runes.length <= limit ? value : runes.slice(0, limit).join('')
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
    onCitations: (citations: ArchiveCitation[]) => void
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
    onCitations: (citations: ArchiveCitation[]) => void
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

  if (event === 'citations' && Array.isArray(payload.citations)) {
    handlers.onCitations(normalizeArchiveCitations(payload.citations))
  } else if (event === 'sources' && Array.isArray(payload.sources)) {
    handlers.onCitations(normalizeArchiveCitations(payload.sources))
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

function citationKey(citation: ArchiveCitation, index: number) {
  return `${citation.title ?? citation.url ?? citation.source_type ?? 'citation'}-${citation.chunk_index ?? index}-${index}`
}

function handleCitationToggle(event: Event, citation: ArchiveCitation, index: number) {
  const element = event.currentTarget as HTMLDetailsElement | null
  const key = citationKey(citation, index)
  openCitationKeys.value = {
    ...openCitationKeys.value,
    [key]: Boolean(element?.open)
  }
}

function isCitationOpen(citation: ArchiveCitation, index: number) {
  return Boolean(openCitationKeys.value[citationKey(citation, index)])
}

function scoreText(score?: number) {
  return typeof score === 'number' ? score.toFixed(3) : '0.000'
}

function citationText(citation: ArchiveCitation) {
  return citation.snippet || ''
}

function copyAnswerWithCitations(record: ChatMessage) {
  copyText(formatAnswerWithCitations(record))
}

function formatAnswerWithCitations(record: ChatMessage) {
  const answer = record.answer.trim()
  if (!answer) {
    return ''
  }
  if (!record.citations.length) {
    return answer
  }

  const citations = record.citations.map((citation, index) => {
    const lines = [`[${index + 1}] ${citation.title || t('archive.citations.untitled', { index: index + 1 })}`]
    const snippet = citationText(citation)
    if (snippet) {
      lines.push(snippet)
    }
    if (citation.url) {
      lines.push(citation.url)
    }
    return lines.join('\n')
  }).join('\n\n')

  return `${answer}\n\n${t('archive.copy.sourcesHeading')}\n${citations}`
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
