<template>
  <div data-site-performance class="space-y-6">
    <dl class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div v-for="item in presentation.kpis" :key="item.key" :data-site-performance-kpi="item.key">
        <dt class="site-detail-note">{{ item.label }}</dt><dd class="site-detail-health__value">{{ item.value }}</dd>
      </div>
    </dl>
    <section data-site-performance-waterfall>
      <h3 class="site-overview-title">{{ t('siteObservation.waterfall') }}</h3>
      <p class="site-detail-note mt-1">{{ t('siteObservation.timingHint') }}</p>
      <ol class="mt-3 space-y-3">
        <li v-for="item in presentation.timings" :key="item.key" :data-site-timing="item.key" class="grid min-w-0 gap-x-4 gap-y-1 sm:grid-cols-[8rem_minmax(0,1fr)_6rem] sm:items-center">
          <span>{{ item.label }}</span>
          <div class="site-observation-timing-track min-w-0" aria-hidden="true"><span class="site-observation-timing-bar block" :data-total="item.key === 'total'" :style="{ width: `${item.fraction * 100}%` }" /></div>
          <span class="site-detail-note sm:text-right">{{ item.text }}</span>
        </li>
      </ol>
    </section>
    <section :data-site-performance-history-state="history.state" :data-site-performance-points="history.rows.length" :aria-busy="history.state === 'loading'">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h3 class="site-overview-title">{{ t('siteObservation.history') }}</h3>
        <div role="group" :aria-label="t('siteObservation.samples')" class="flex gap-1">
          <button v-for="sample in samples" :key="sample" type="button" :data-site-performance-sample="sample" :aria-pressed="history.sample === sample"
            class="site-observation-sample" @click="emit('sample', sample)">{{ sample }}</button>
        </div>
      </div>
      <p v-if="history.state !== 'ready'" role="status" class="site-observation-empty mt-4">{{ t('siteObservation.historyStates.' + history.state) }}</p>
      <button v-if="history.state === 'unavailable'" type="button" data-site-history-retry class="gf-button gf-button--ghost mt-2" @click="emit('retry')">{{ t('siteObservation.retry') }}</button>
      <template v-if="history.state === 'ready'">
        <p class="site-detail-note mt-2">{{ history.rows.length }} / {{ history.total }} {{ t('siteObservation.samples') }}</p>
        <SitePingHistoryChart v-if="history.rows.some(row => row.rtt !== null)" :rows="history.rows" class="mt-3" />
        <p v-else class="site-observation-empty mt-3">{{ t('siteObservation.noRtt') }}</p>
        <details class="site-observation-disclosure mt-4" data-site-history-table>
          <summary>{{ t('siteObservation.historyRecords') }}</summary>
          <div class="mt-3 max-h-96 overflow-y-auto">
            <table class="site-observation-history-table w-full border-collapse break-words text-left">
              <caption class="sr-only">{{ t('siteObservation.historyRecords') }}</caption>
              <thead><tr><th class="text-left">{{ t('siteObservation.fields.observed') }}</th><th class="text-left">{{ t('siteObservation.fields.status') }}</th><th class="text-left">RTT</th><th class="text-left">{{ t('siteObservation.fields.loss') }}</th></tr></thead>
              <tbody><tr v-for="row in history.rows" :key="row.key">
                <td>{{ row.time }}</td><td>{{ t('siteDetail.states.' + row.status) }}</td><td>{{ row.rtt === null ? '—' : `${row.rtt} ms` }}</td><td>{{ row.loss === null ? '—' : `${Math.round(row.loss * 100) / 100}%` }}</td>
              </tr></tbody>
            </table>
          </div>
        </details>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import type { SiteHistoryPresentation, SiteHistorySample } from '~/composables/useSiteObservationHistory'
import SitePingHistoryChart from './SitePingHistoryChart.vue'
defineProps<{ presentation: SiteObservationPresentation; history: SiteHistoryPresentation }>()
const emit = defineEmits<{ sample: [sample: SiteHistorySample]; retry: [] }>()
const { t } = useI18n()
const samples: SiteHistorySample[] = [20, 60, 100]
</script>
