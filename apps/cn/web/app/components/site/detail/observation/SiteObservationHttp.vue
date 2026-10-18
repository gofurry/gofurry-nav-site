<template>
  <div data-site-http data-site-http-composite class="site-observation-composite">
    <section data-site-http-response>
      <h3 class="site-observation-heading">{{ t('siteObservation.responseSummary') }}</h3>
      <dl class="site-observation-response-strip mt-4 grid min-w-0 grid-cols-2 items-end gap-4 sm:grid-cols-3">
        <div v-for="item in presentation.headline" :key="item.key" :data-site-evidence="item.key" class="min-w-0" :class="{ 'col-span-2 sm:col-span-1': item.key === 'status' }">
          <dt class="site-observation-caption">{{ item.label }}</dt><dd class="site-observation-response-value mt-1 break-words" :data-tone="item.tone">{{ item.value }}</dd>
        </div>
      </dl>
      <SiteObservationEvidence :items="presentation.evidence" class="site-observation-response-evidence mt-4" />
    </section>
    <div v-if="presentation.redirects.length" data-site-redirect-separator class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
    <section v-if="presentation.redirects.length" data-site-http-redirects class="site-observation-flow">
      <h3 class="site-observation-heading">{{ t('siteObservation.redirects') }}</h3>
      <ol class="mt-3 grid min-w-0 gap-x-8 gap-y-7 md:grid-cols-3">
        <li v-for="(url, index) in presentation.redirects" :key="index" data-site-redirect-node :data-connector="connector(index, presentation.redirects.length)"
          :style="compactPlacement(index)" class="site-detail-redirect-node site-observation-technical relative min-w-0 break-all md:col-[var(--timeline-column)] md:row-[var(--timeline-row)]">
          {{ url }}<SiteSequenceArrow :direction="connector(index, presentation.redirects.length)" />
        </li>
      </ol>
    </section>
    <div class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
    <section data-site-http-headers class="relative">
      <h3 class="site-observation-heading mb-4" :class="{ 'max-w-[50%]': presentation.headers.length }">{{ t('siteObservation.commonHeaders') }}</h3>
      <SiteObservationEvidence v-if="presentation.commonHeaders.length" :items="presentation.commonHeaders" />
      <p v-else class="site-observation-caption">{{ t('siteObservation.noEvidence') }}</p>
      <details v-if="presentation.headers.length" data-site-http-all-headers class="site-observation-disclosure site-observation-header-disclosure">
        <summary class="absolute right-0 top-0 max-w-[45%] text-right">{{ t('siteObservation.allHeaders') }} ({{ presentation.headers.length }})</summary>
        <SiteObservationEvidence :items="presentation.headers" class="site-observation-raw-headers mt-4" />
      </details>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { compactPlacement, connector } from '~/utils/serpentineSequence'
import SiteSequenceArrow from '../SiteSequenceArrow.vue'
import SiteObservationEvidence from './SiteObservationEvidence.vue'
defineProps<{ presentation: SiteObservationPresentation['http'] }>()
const { t } = useI18n()
</script>
