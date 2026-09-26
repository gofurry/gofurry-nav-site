<template>
  <div data-site-http class="space-y-6">
    <SiteObservationFacts :items="presentation.facts" />
    <section v-if="presentation.redirects.length" data-site-http-redirects>
      <h3 class="site-overview-title">{{ t('siteObservation.redirects') }}</h3>
      <ol class="site-observation-chain mt-3 space-y-2">
        <li v-for="(url, index) in presentation.redirects" :key="index" class="break-words">{{ index + 1 }}. {{ url }}</li>
      </ol>
    </section>
    <section data-site-http-headers>
      <h3 class="site-overview-title mb-2">{{ t('siteObservation.commonHeaders') }}</h3>
      <SiteObservationFacts v-if="presentation.commonHeaders.length" :items="presentation.commonHeaders" />
      <p v-else class="site-detail-note">{{ t('siteObservation.noEvidence') }}</p>
      <details v-if="presentation.headers.length" data-site-http-all-headers class="site-observation-disclosure mt-4">
        <summary>{{ t('siteObservation.allHeaders') }} ({{ presentation.headers.length }})</summary>
        <SiteObservationFacts :items="presentation.headers" class="mt-3" />
      </details>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['http'] }>()
const { t } = useI18n()
</script>
