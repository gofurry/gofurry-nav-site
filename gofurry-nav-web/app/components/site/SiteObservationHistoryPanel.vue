<template>
  <div class="info-tab-body history-panel">
    <div v-if="loading" class="panel-empty">{{ label('观测历史加载中', 'Loading history') }}</div>
    <section
      v-for="history in histories"
      :key="history.protocol"
      class="history-section"
    >
      <div class="history-head">
        <h4 class="history-title">{{ history.title }}</h4>
        <div v-if="history.totalPages > 1" class="history-pager">
          <button
            type="button"
            class="history-page-button"
            :disabled="history.page <= 1"
            @click="$emit('setPage', history.protocol, history.page - 1)"
          >
            {{ label('上一页', 'Prev') }}
          </button>
          <span class="history-page-count">{{ history.page }}/{{ history.totalPages }}</span>
          <button
            type="button"
            class="history-page-button"
            :disabled="history.page >= history.totalPages"
            @click="$emit('setPage', history.protocol, history.page + 1)"
          >
            {{ label('下一页', 'Next') }}
          </button>
        </div>
      </div>
      <div v-if="history.items.length" class="history-list">
        <div
          v-for="item in history.visibleItems"
          :key="`${history.protocol}:${item.observed_at}:${item.duration_ms}`"
          class="history-row"
        >
          <span
            :aria-label="statusText(item.status)"
            :class="['status-dot', statusDotClass(item.status)]"
            :title="statusText(item.status)"
          />
          <span class="history-summary">{{ summaryFor(history.protocol, item) }}</span>
          <span class="history-time">{{ formatTime(item.observed_at) }}</span>
        </div>
      </div>
      <div v-else class="panel-empty">{{ label('暂无历史', 'No history') }}</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { i18n } from '@/main'
import type { CollectorEnvelope } from '@/types/nav'
import type { ObservationHistoryItem, ObservationProtocol } from './detailTypes'

defineProps<{
  histories: ObservationHistoryItem[]
  loading: boolean
  summaryFor: (protocol: string, envelope: CollectorEnvelope) => string
}>()

defineEmits<{
  setPage: [protocol: ObservationProtocol, page: number]
}>()

function statusText(status: string) {
  if (status === 'success') return label('成功', 'Success')
  if (status === 'failure') return label('失败', 'Failure')
  if (status === 'skipped') return label('跳过', 'Skipped')
  return status || '-'
}

function statusDotClass(status: string) {
  if (status === 'success') return 'is-success'
  if (status === 'failure') return 'is-failure'
  if (status === 'skipped') return 'is-skipped'
  return 'is-unknown'
}

function formatTime(value: string) {
  if (!value) return '-'
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.info-tab-body {
  margin-top: 1.25rem;
}

.panel-empty {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 0.9rem 0;
  color: #64748b;
  font-size: 0.9rem;
}

:global(html.dark .panel-empty){
  border-top-color: rgba(251, 146, 60, 0.16);
  color: #94a3b8;
}

.history-section {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 1rem 0;
}

:global(html.dark .history-section){
  border-top-color: rgba(251, 146, 60, 0.16);
}

.history-section:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.history-title {
  color: #111827;
  font-size: 0.94rem;
  font-weight: 800;
}

:global(html.dark .history-title){
  color: #f8fafc;
}

.history-pager {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.78rem;
}

.history-page-button {
  border-radius: 8px;
  background: transparent;
  padding: 0.25rem 0.55rem;
  color: #475569;
  transition: background-color 500ms ease, color 500ms ease;
}

:global(html.dark .history-page-button){
  color: #cbd5e1;
}

.history-page-button:hover:not(:disabled) {
  background: rgba(255, 237, 213, 0.72);
  color: #9a4a12;
}

:global(html.dark .history-page-button:hover:not(:disabled)){
  background: rgba(251, 146, 60, 0.14);
  color: #fdba74;
}

.history-page-button:disabled {
  cursor: not-allowed;
  color: #cbd5e1;
}

:global(html.dark .history-page-button:disabled){
  color: #475569;
}

.history-page-count {
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

:global(html.dark .history-page-count){
  color: #94a3b8;
}

.history-list {
  margin-top: 0.7rem;
}

.history-row {
  display: grid;
  grid-template-columns: 1rem minmax(0, 1fr) 8.5rem;
  gap: 0.75rem;
  align-items: center;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  padding: 0.6rem 0.75rem 0.6rem 0.65rem;
  transition: background-color 500ms ease, border-color 500ms ease;
}

:global(html.dark .history-row){
  border-bottom-color: rgba(148, 163, 184, 0.12);
}

.history-row:hover {
  background: rgba(255, 237, 213, 0.52);
  border-left-color: rgba(251, 140, 47, 0.45);
}

:global(html.dark .history-row:hover){
  background: rgba(251, 146, 60, 0.12);
  border-left-color: rgba(251, 146, 60, 0.48);
}

.history-row:last-child {
  border-bottom: 0;
}

.history-summary {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #1f2937;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.55;
}

:global(html.dark .history-summary){
  color: #e2e8f0;
}

.history-time {
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.78rem;
  line-height: 1.55;
  text-align: right;
}

:global(html.dark .history-time){
  color: #94a3b8;
}

.status-dot {
  display: inline-block;
  width: 0.58rem;
  height: 0.58rem;
  flex: 0 0 auto;
  border-radius: 999px;
  vertical-align: middle;
}

.status-dot.is-success {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.14);
}

.status-dot.is-failure {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.12);
}

.status-dot.is-skipped {
  background: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.12);
}

.status-dot.is-unknown {
  background: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.12);
}

@media (max-width: 640px) {
  .history-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .history-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.35rem;
  }

  .history-time {
    text-align: left;
  }
}
</style>
