<template>
  <div class="light-probe-panel">
    <div class="light-probe-header">
      <h3 class="info-tabs-title">{{ label('低频轻探测', 'Light probes') }}</h3>
    </div>
    <div v-if="entries.length" class="light-probe-grid">
      <button
        v-for="probe in entries"
        :key="probe.protocol"
        type="button"
        class="light-probe-card"
        @click="selectedProbe = probe"
      >
        <div class="light-probe-card-head">
          <span class="light-probe-card-title">{{ protocolName(probe.protocol) }}</span>
          <span
            :aria-label="statusText(probe.status)"
            :class="['status-dot', statusDotClass(probe.status)]"
            :title="statusText(probe.status)"
          />
        </div>
        <div v-if="probe.items.length" class="light-probe-facts">
          <div
            v-for="item in probe.items"
            :key="item.label"
            class="light-probe-fact"
          >
            <span class="light-probe-label">{{ item.label }}</span>
            <span class="light-probe-value">{{ displayValue(item.value) }}</span>
          </div>
        </div>
        <div v-else class="light-probe-empty">{{ label('暂无数据', 'No data') }}</div>
        <div class="light-probe-detail-hint">
          {{ label('点击查看详情', 'Click for details') }}
        </div>
      </button>
    </div>
    <div v-else class="panel-empty">{{ label('暂无数据', 'No data') }}</div>

    <Teleport to="body">
      <Transition name="probe-modal">
        <div
          v-if="selectedProbe"
          class="probe-modal-backdrop"
          @click.self="selectedProbe = null"
        >
          <article
            class="probe-modal-dialog"
            role="dialog"
            aria-modal="true"
          >
            <header class="probe-modal-header">
              <div>
                <p class="probe-modal-eyebrow">{{ label('低频轻探测详情', 'Light probe detail') }}</p>
                <h3 class="probe-modal-title">{{ protocolName(selectedProbe.protocol) }}</h3>
              </div>
              <div class="probe-modal-actions">
                <button
                  type="button"
                  class="probe-modal-close"
                  @click="selectedProbe = null"
                >
                  {{ label('关闭', 'Close') }}
                </button>
              </div>
            </header>

            <div class="probe-modal-body">
              <div class="probe-modal-summary-grid">
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('观测时间', 'Observed') }}</p>
                  <p class="probe-modal-summary-value">{{ formatTime(selectedProbe.observedAt) }}</p>
                </div>
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('耗时', 'Duration') }}</p>
                  <p class="probe-modal-summary-value">{{ formatDuration(selectedProbe.durationMs) }}</p>
                </div>
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('结果', 'Result') }}</p>
                  <p class="probe-modal-summary-value">{{ selectedProbe.errorCode || statusText(selectedProbe.status) }}</p>
                </div>
              </div>

              <div v-if="selectedProbe.errorMessage" class="probe-modal-error">
                {{ selectedProbe.errorMessage }}
              </div>

              <div class="probe-modal-sections">
                <section
                  v-for="section in selectedProbeDetailSections"
                  :key="section.title"
                  class="probe-modal-section"
                >
                  <h4 class="probe-modal-section-title">{{ section.title }}</h4>
                  <div v-if="section.items.length" class="modal-info-list">
                    <div
                      v-for="item in section.items"
                      :key="item.label"
                      class="modal-info-row"
                    >
                      <span class="modal-info-label">{{ item.label }}</span>
                      <span class="modal-info-value">{{ displayValue(item.value) }}</span>
                    </div>
                  </div>
                  <div v-else class="modal-empty">{{ label('暂无数据', 'No data') }}</div>
                </section>
              </div>
            </div>
          </article>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/main'
import type { DetailSection, LightProbeEntry } from './detailTypes'

const props = defineProps<{
  entries: LightProbeEntry[]
  detailSectionsFor: (probe: LightProbeEntry) => DetailSection[]
}>()

const selectedProbe = ref<LightProbeEntry | null>(null)
const selectedProbeDetailSections = computed(() => selectedProbe.value ? props.detailSectionsFor(selectedProbe.value) : [])

function displayValue(value: string | string[]) {
  return Array.isArray(value) ? value.join(', ') : value
}

function protocolName(protocol: string) {
  const map: Record<string, string> = {
    rdap: 'RDAP',
    robots: 'robots.txt',
    security_txt: 'security.txt',
    page_assets: 'Page assets',
    port_check: 'Port check',
    waf_canary: 'WAF canary',
  }
  return map[protocol] || protocol
}

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

function formatDuration(value: number) {
  return Number.isFinite(value) && value >= 0 ? `${Math.round(value)}ms` : '-'
}

function formatTime(value: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.light-probe-panel {
  min-width: 0;
}

.light-probe-header {
  display: flex;
  align-items: center;
  min-height: 2.45rem;
}

.info-tabs-title {
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.35;
}

:global(html.dark .info-tabs-title){
  color: #f8fafc;
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

.light-probe-grid {
  margin-top: 1.15rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.light-probe-card {
  display: block;
  width: 100%;
  min-width: 0;
  break-inside: avoid;
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.42);
  padding: 0.74rem 0.8rem;
  text-align: left;
  transition: background-color 500ms ease, box-shadow 500ms ease;
}

:global(html.dark .light-probe-card){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.10);
}

.light-probe-card:hover,
.light-probe-card:focus-visible {
  background: rgba(255, 247, 235, 0.72);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.22), 0 0 0 4px rgba(251, 140, 47, 0.055);
  outline: none;
}

:global(html.dark .light-probe-card:hover),
:global(html.dark .light-probe-card:focus-visible){
  background: rgba(30, 41, 59, 0.78);
  box-shadow: inset 0 0 0 1px rgba(251, 146, 60, 0.22), 0 0 0 4px rgba(251, 146, 60, 0.06);
}

.light-probe-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.light-probe-card-title {
  min-width: 0;
  color: #111827;
  font-size: 0.92rem;
  font-weight: 800;
}

:global(html.dark .light-probe-card-title){
  color: #f8fafc;
}

.light-probe-facts {
  margin-top: 0.62rem;
  display: grid;
  gap: 0.3rem;
}

.light-probe-fact {
  display: grid;
  grid-template-columns: minmax(4.4rem, 0.42fr) minmax(0, 1fr);
  gap: 0.5rem;
  align-items: start;
  min-width: 0;
  font-size: 0.84rem;
  line-height: 1.45;
}

.light-probe-label {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #64748b;
  font-weight: 800;
}

:global(html.dark .light-probe-label){
  color: #94a3b8;
}

.light-probe-value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

:global(html.dark .light-probe-value){
  color: #e2e8f0;
}

.light-probe-empty {
  margin-top: 0.7rem;
  color: #64748b;
  font-size: 0.84rem;
}

:global(html.dark .light-probe-empty){
  color: #94a3b8;
}

.light-probe-detail-hint {
  margin-top: 0.65rem;
  color: #ea580c;
  font-size: 0.74rem;
  font-weight: 800;
  opacity: 0;
  transition: opacity 500ms ease;
}

.light-probe-card:hover .light-probe-detail-hint,
.light-probe-card:focus-visible .light-probe-detail-hint {
  opacity: 1;
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

.probe-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.46);
  padding: 1.5rem 1rem;
  backdrop-filter: blur(6px);
}

:global(html.dark .probe-modal-backdrop){
  background: rgba(2, 6, 23, 0.66);
}

.probe-modal-dialog {
  display: flex;
  width: min(100%, 56rem);
  max-height: 88vh;
  flex-direction: column;
  overflow: hidden;
  border-radius: 8px;
  background:
    radial-gradient(circle at 8% 0%, rgba(251, 140, 47, 0.08), transparent 30%),
    linear-gradient(120deg, rgba(255, 247, 235, 0.88), rgba(255, 250, 242, 0.94)),
    rgba(255, 247, 235, 0.90);
  color: #111827;
  box-shadow:
    inset 0 0 0 1px rgba(251, 140, 47, 0.16),
    0 24px 70px rgba(15, 23, 42, 0.22);
}

:global(html.dark .probe-modal-dialog){
  background:
    radial-gradient(circle at 8% 0%, rgba(251, 146, 60, 0.12), transparent 30%),
    linear-gradient(120deg, rgba(15, 23, 42, 0.94), rgba(30, 41, 59, 0.92)),
    rgba(15, 23, 42, 0.94);
  color: #e2e8f0;
  box-shadow:
    inset 0 0 0 1px rgba(251, 146, 60, 0.16),
    0 24px 70px rgba(0, 0, 0, 0.44);
}

.probe-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(251, 140, 47, 0.14);
  padding: 1rem 1.25rem 0.9rem;
}

:global(html.dark .probe-modal-header){
  border-bottom-color: rgba(251, 146, 60, 0.16);
}

.probe-modal-eyebrow {
  color: #ea580c;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.35;
}

.probe-modal-title {
  margin-top: 0.18rem;
  color: #111827;
  font-size: 1.22rem;
  font-weight: 850;
  line-height: 1.25;
}

:global(html.dark .probe-modal-title){
  color: #f8fafc;
}

.probe-modal-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.85rem;
}

.probe-modal-close {
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.78);
  padding: 0.42rem 0.72rem;
  color: #475569;
  font-size: 0.86rem;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  transition: background-color 500ms ease, color 500ms ease, box-shadow 500ms ease;
}

:global(html.dark .probe-modal-close){
  background: rgba(15, 23, 42, 0.68);
  color: #cbd5e1;
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}

.probe-modal-close:hover,
.probe-modal-close:focus-visible {
  background: #fdba74;
  color: #111827;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18), 0 0 0 4px rgba(251, 140, 47, 0.06);
  outline: none;
}

:global(html.dark .probe-modal-close:hover),
:global(html.dark .probe-modal-close:focus-visible){
  background: rgba(251, 146, 60, 0.24);
  color: #fff7ed;
  box-shadow: inset 0 0 0 1px rgba(251, 146, 60, 0.20), 0 0 0 4px rgba(251, 146, 60, 0.07);
}

.probe-modal-body {
  overflow-y: auto;
  padding: 1rem 1.25rem 1.25rem;
}

.probe-modal-summary-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
}

.probe-modal-summary-item {
  min-width: 0;
  padding: 0.72rem 0;
}

.probe-modal-summary-label {
  color: #64748b;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.45;
}

:global(html.dark .probe-modal-summary-label){
  color: #94a3b8;
}

.probe-modal-summary-value {
  margin-top: 0.18rem;
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.45;
}

:global(html.dark .probe-modal-summary-value){
  color: #e2e8f0;
}

.probe-modal-error {
  margin-top: 1rem;
  border-left: 2px solid rgba(239, 68, 68, 0.34);
  background: rgba(254, 242, 242, 0.62);
  padding: 0.72rem 0.85rem;
  color: #991b1b;
  font-size: 0.88rem;
  line-height: 1.55;
}

:global(html.dark .probe-modal-error){
  background: rgba(127, 29, 29, 0.28);
  color: #fecaca;
}

.probe-modal-sections {
  margin-top: 1rem;
  display: grid;
  gap: 0.95rem;
}

.probe-modal-section {
  min-width: 0;
  border-top: 1px solid rgba(251, 140, 47, 0.14);
  padding-top: 0.95rem;
}

:global(html.dark .probe-modal-section){
  border-top-color: rgba(251, 146, 60, 0.16);
}

.probe-modal-section:first-child {
  border-top: 0;
  padding-top: 0;
}

.probe-modal-section-title {
  margin-bottom: 0.55rem;
  color: #111827;
  font-size: 0.94rem;
  font-weight: 850;
  line-height: 1.35;
}

:global(html.dark .probe-modal-section-title){
  color: #f8fafc;
}

.modal-info-list {
  display: grid;
}

.modal-info-row {
  display: grid;
  grid-template-columns: minmax(7.5rem, 12rem) minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
  min-width: 0;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  padding: 0.48rem 0.65rem 0.48rem 0.55rem;
  transition: background-color 500ms ease, border-color 500ms ease;
}

:global(html.dark .modal-info-row){
  border-bottom-color: rgba(148, 163, 184, 0.12);
}

.modal-info-row:hover {
  background: rgba(255, 237, 213, 0.48);
  border-left-color: rgba(251, 140, 47, 0.42);
}

:global(html.dark .modal-info-row:hover){
  background: rgba(251, 146, 60, 0.12);
  border-left-color: rgba(251, 146, 60, 0.48);
}

.modal-info-row:last-child {
  border-bottom: 0;
}

.modal-info-label {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #64748b;
  font-size: 0.88rem;
  font-weight: 800;
  line-height: 1.5;
}

:global(html.dark .modal-info-label){
  color: #94a3b8;
}

.modal-info-value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.5;
}

:global(html.dark .modal-info-value){
  color: #e2e8f0;
}

.modal-empty {
  color: #64748b;
  font-size: 0.88rem;
}

:global(html.dark .modal-empty){
  color: #94a3b8;
}

.probe-modal-enter-active,
.probe-modal-leave-active {
  transition: opacity 160ms ease;
}

.probe-modal-enter-active article,
.probe-modal-leave-active article {
  transition: transform 160ms ease;
}

.probe-modal-enter-from,
.probe-modal-leave-to {
  opacity: 0;
}

.probe-modal-enter-from article,
.probe-modal-leave-to article {
  transform: translateY(8px) scale(0.98);
}

@media (max-width: 640px) {
  .probe-modal-backdrop {
    align-items: stretch;
    padding: 0.75rem;
  }

  .probe-modal-dialog {
    max-height: calc(100vh - 1.5rem);
  }

  .probe-modal-header {
    align-items: flex-start;
    padding: 0.95rem 1rem 0.85rem;
  }

  .probe-modal-body {
    padding: 0.9rem 1rem 1rem;
  }

  .modal-info-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.12rem;
  }
}

@media (min-width: 768px) {
  .probe-modal-summary-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .probe-modal-summary-item {
    padding: 0.72rem 1rem;
  }

  .probe-modal-summary-item:first-child {
    padding-left: 0;
  }

  .light-probe-grid {
    display: block;
    column-count: 2;
    column-gap: 0.75rem;
  }

  .light-probe-card {
    margin-bottom: 0.75rem;
  }
}
</style>
