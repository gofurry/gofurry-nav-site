<template>
  <div data-site-dns class="space-y-5">
    <section data-site-dns-summary class="site-observation-composite">
      <h3 class="site-observation-heading">{{ t('siteObservation.dnsSummary') }}</h3>
      <dl class="site-observation-count-strip mt-4 grid grid-cols-5 gap-2"><div v-for="item in presentation.facts.slice(0, 5)" :key="item.key" :data-site-evidence="item.key"><dt class="site-observation-caption">{{ item.label }}</dt><dd class="site-observation-measure mt-1">{{ item.value }}</dd></div></dl>
      <SiteObservationEvidence :items="presentation.facts.slice(5)" class="site-observation-response-evidence mt-4" />
    </section>
    <section v-if="presentation.chains.length" data-site-dns-chain class="site-observation-flow">
      <h3 class="site-observation-heading">{{ t('siteObservation.resolution') }}</h3>
      <ol v-for="(chain, index) in presentation.chains" :key="index" class="site-observation-technical mt-3 flex flex-wrap gap-2">
        <li v-for="(value, step) in chain" :key="step" class="inline-flex min-w-0 items-center gap-2"><PhArrowRight v-if="step" class="site-detail-icon shrink-0" aria-hidden="true" /><span class="min-w-0 break-all">{{ value }}</span></li>
      </ol>
    </section>
    <section data-site-dns-ledger class="site-observation-evidence">
      <h3 class="site-observation-heading">{{ t('siteObservation.recordLedger') }}</h3>
      <section v-for="group in presentation.groups" :key="group.type" :data-site-dns-group="group.type" class="mt-4 space-y-1.5">
        <h4 class="site-observation-ledger-heading mb-2 flex items-center gap-3"><span>{{ group.type }}</span><span class="min-w-0 flex-1" aria-hidden="true" /><span class="site-observation-caption">{{ group.records.length }}</span></h4>
        <div v-for="(record, index) in group.records" :key="index" class="site-observation-ledger-row">
          <div class="flex flex-wrap justify-between gap-2"><span class="site-observation-technical min-w-0 break-all">{{ record.value }}</span><span class="site-observation-caption shrink-0">TTL {{ record.ttl }}</span></div>
          <details v-if="record.details.length" class="site-observation-disclosure mt-2">
            <summary>{{ t('siteObservation.recordDetails') }}</summary>
            <SiteObservationEvidence :items="record.details" class="mt-2" />
          </details>
        </div>
      </section>
      <p v-if="!presentation.groups.length" class="site-observation-caption mt-3">{{ t('siteObservation.noEvidence') }}</p>
      <div v-if="presentation.signals.length" class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
      <section v-if="presentation.signals.length" data-site-dns-risks class="site-observation-signal-rail">
        <h3 class="site-observation-heading">{{ t('siteObservation.dnsObservationSignals') }}</h3>
        <ul class="mt-3 flex flex-wrap gap-x-5 gap-y-3"><li v-for="signal in presentation.signals" :key="signal.code" :data-site-dns-signal="signal.code" class="flex min-w-0 items-start gap-2">
          <PhWarning v-if="signal.tone === 'warning'" class="site-detail-icon mt-1 shrink-0" data-tone="warning" aria-hidden="true" />
          <PhInfo v-else class="site-detail-icon mt-1 shrink-0" aria-hidden="true" />
          <p :data-tone="signal.tone" class="min-w-0 break-words">{{ signal.label }}</p>
        </li></ul>
      </section>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhArrowRight, PhInfo, PhWarning } from '@phosphor-icons/vue'
import SiteObservationEvidence from './SiteObservationEvidence.vue'
defineProps<{ presentation: SiteObservationPresentation['dns'] }>()
const { t } = useI18n()
</script>
