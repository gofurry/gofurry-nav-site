<template>
  <section class="detail-observation-overview">
    <div class="overview-strip">
      <div
        v-for="item in stripItems"
        :key="item.label"
        class="overview-stat"
        :class="item.tone"
      >
        <div class="text-[11px] text-slate-500">{{ item.label }}</div>
        <div class="truncate text-sm font-bold text-slate-900">{{ item.value }}</div>
      </div>
    </div>

    <div class="overview-detail-grid">
      <div class="protocol-rail-panel">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-black text-slate-950">{{ label('当前采集', 'Current Checks') }}</h3>
          </div>
          <div class="rounded-full bg-orange-100 px-3 py-1 text-xs text-orange-700">
            {{ protocolAvailabilityText }}
          </div>
        </div>

        <div class="protocol-rail">
          <article
            v-for="entry in protocolEntries"
            :key="entry.protocol"
            class="protocol-node"
            :class="entry.tone"
          >
            <div class="min-w-0">
              <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
                <div class="flex items-center gap-2">
                  <span class="protocol-dot" />
                  <strong class="text-sm text-slate-950">{{ entry.label }}</strong>
                </div>
                <span class="font-mono text-[11px] text-slate-500">{{ entry.observedAt }}</span>
              </div>
              <div class="protocol-metrics mt-2 grid gap-1 text-xs text-slate-600 sm:grid-cols-2">
                <span>{{ label('耗时', 'Time') }}: <b>{{ entry.duration }}</b></span>
                <span>{{ label('过期阈值', 'Stale') }}: <b>{{ entry.staleAfter }}</b></span>
              </div>
            </div>
          </article>
        </div>

        <div class="signal-note mt-4 px-4 py-3 text-xs text-slate-600">
          <div class="mb-1 font-semibold text-slate-700">{{ label('观测信号', 'Signals') }}</div>
          <div v-if="riskMessages.length" class="space-y-1">
            <div v-for="message in riskMessages" :key="message">{{ message }}</div>
          </div>
          <div v-else>{{ label('暂无需要关注的信号', 'No notable signals') }}</div>
        </div>
      </div>

      <div class="security-matrix-panel">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-black text-slate-950">{{ label('安全响应头', 'Security Headers') }}</h3>
          </div>
          <div class="rounded-full bg-orange-100 px-3 py-1 text-xs text-orange-700">
            {{ securityHeaderRatio }}
          </div>
        </div>

        <div class="security-matrix">
          <div
            v-for="item in securityHeaderItems"
            :key="item.label"
            class="security-header-cell"
            :class="{ 'is-ok': item.ok }"
          >
            <span class="status-dot" />
            <span class="min-w-0 truncate">{{ item.label }}</span>
            <span class="ml-auto text-xs">{{ item.ok ? label('是', 'Yes') : label('否', 'No') }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { i18n } from '@/main'
import type { ObservationStripItem, ProtocolTrackEntry, SecurityHeaderItem } from './detailTypes'

defineProps<{
  protocolAvailabilityText: string
  protocolEntries: ProtocolTrackEntry[]
  riskMessages: string[]
  securityHeaderItems: SecurityHeaderItem[]
  securityHeaderRatio: string
  stripItems: ObservationStripItem[]
}>()

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.detail-observation-overview {
  margin-top: 1.5rem;
  border-radius: 8px;
  background:
    radial-gradient(circle at 10% 0%, rgba(251, 140, 47, 0.10), transparent 36%),
    rgba(255, 247, 235, 0.70);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.overview-strip {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.overview-stat {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-radius: 0.55rem;
  background: rgba(255, 232, 196, 0.72);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  padding: 0.72rem 0.82rem;
}

.overview-stat.is-ok {
  background: rgba(220, 252, 231, 0.64);
  box-shadow: inset 0 0 0 1px rgba(16, 185, 129, 0.10);
}

.overview-stat.is-warn {
  background: rgba(253, 224, 71, 0.34);
  box-shadow: inset 0 0 0 1px rgba(245, 158, 11, 0.12);
}

.overview-detail-grid {
  margin-top: 0.9rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.9rem;
}

.protocol-rail-panel,
.security-matrix-panel {
  border-radius: 0.75rem;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.protocol-rail {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.65rem;
}

.protocol-node {
  position: relative;
  border-radius: 0.65rem;
  background: rgba(255, 230, 191, 0.70);
  padding: 0.85rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.08);
}

.protocol-dot {
  height: 0.75rem;
  width: 0.75rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: #f59e0b;
}

.protocol-node.is-ok .protocol-dot,
.security-header-cell.is-ok .status-dot {
  background: #10b981;
}

.protocol-node.is-bad .protocol-dot {
  background: #f43f5e;
}

.security-matrix {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.6rem;
}

.security-header-cell {
  display: flex;
  min-height: 2.6rem;
  min-width: 0;
  align-items: center;
  gap: 0.55rem;
  border-radius: 0.55rem;
  background: rgba(255, 230, 191, 0.68);
  padding: 0.65rem 0.8rem;
  color: #475569;
  font-size: 0.78rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.07);
}

.signal-note {
  border-radius: 0.55rem;
  background: rgba(255, 230, 191, 0.66);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.07);
}

.status-dot {
  height: 0.5rem;
  width: 0.5rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: #cbd5e1;
}

@media (min-width: 900px) {
  .overview-strip {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (min-width: 1180px) {
  .overview-detail-grid {
    grid-template-columns: minmax(0, 1.55fr) minmax(22rem, 0.9fr);
  }
}

@media (min-width: 640px) {
  .protocol-metrics span:last-child {
    justify-self: end;
    text-align: right;
  }

  .security-matrix {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 760px) {
  .protocol-rail {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
