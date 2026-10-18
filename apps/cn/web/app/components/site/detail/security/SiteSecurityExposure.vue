<template>
  <div data-site-security-exposure class="space-y-8">
    <section data-site-port-check :data-state="ports.state" class="site-detail-surface">
      <div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="site-overview-title">{{ t('siteSecurity.portCheck') }}</h3><p>{{ ports.label }}</p></div>
      <p class="site-detail-note mt-1">{{ t('siteSecurity.portHint') }}</p>
      <p v-if="ports.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="ports.truncated" data-site-port-truncated class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <dl class="mt-4 flex flex-wrap gap-x-5 gap-y-2"><div v-for="item in ports.summary" :key="item.key" :data-site-evidence="item.key" class="inline-flex items-baseline gap-2"><dt class="site-detail-note">{{ item.label }}</dt><dd class="site-security-port-count">{{ item.value }}</dd></div></dl>
      <SiteObservationFacts v-if="ports.errors.length" :items="ports.errors" class="mt-2" />
      <div class="mt-4 space-y-1.5">
      <div v-for="item in ports.results" :key="item.key" data-site-port-result :data-state="item.status" class="site-security-row">
        <dl class="grid min-w-0 grid-cols-[3rem_minmax(0,1fr)_minmax(0,1fr)_auto] gap-2"><div v-for="field in item.facts" :key="field.key" class="min-w-0"><dt class="sr-only">{{ field.label }}</dt><dd :data-tone="field.key === 'status' ? item.tone : undefined" class="break-all">{{ field.value }}</dd></div></dl>
        <SiteObservationFacts v-if="item.errors.length" :items="item.errors" class="mt-2" />
      </div>
      </div>
      <details class="site-security-disclosure mt-3" data-site-port-metadata><summary>{{ t('siteSecurity.probeMetadata') }}</summary><SiteObservationFacts :items="ports.metadata" class="mt-2" /><SiteObservationFacts :items="ports.facts" /></details>
    </section>
    <section data-site-waf-canary :data-state="waf.state" class="site-detail-surface">
      <h3 class="site-overview-title flex items-center gap-2">{{ t('siteSecurity.wafCanary') }}<SiteDetailHelpTooltip :label="t('siteSecurity.wafCanary')" :text="t('siteSecurity.wafHint')" /></h3>
      <p class="site-security-primary-value mt-3" :data-tone="waf.tone">{{ waf.summaryText }}</p><p class="site-detail-note mt-1">{{ waf.label }}</p>
      <p v-if="waf.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="waf.truncated" data-site-waf-truncated class="site-security-attention mt-3">{{ t('siteSecurity.truncated') }}</p>
      <dl class="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4"><div v-for="item in waf.summary" :key="item.key" :data-site-evidence="item.key"><dt class="site-detail-note">{{ item.label }}</dt><dd>{{ item.value }}</dd></div></dl>
      <SiteObservationFacts v-if="waf.errors.length" :items="waf.errors" class="mt-2" />
      <details class="site-security-disclosure mt-3" data-site-waf-cases>
        <summary>{{ t('siteSecurity.cases') }}</summary>
        <div class="mt-3 space-y-1.5">
        <div v-for="item in waf.cases" :key="item.key" data-site-waf-case class="site-security-row">
          <dl class="grid min-w-0 gap-3 md:grid-cols-3"><div v-for="field in item.facts" :key="field.key" class="min-w-0"><dt class="site-detail-note">{{ field.label }}</dt><dd class="break-all">{{ field.value }}</dd></div></dl>
          <SiteObservationFacts v-if="item.errors.length" :items="item.errors" class="mt-2" />
        </div>
        </div>
        <p v-if="!waf.cases.length" class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
        <SiteObservationFacts :items="waf.metadata" class="mt-2" /><SiteObservationFacts :items="waf.facts" />
      </details>
    </section>
  </div>
</template>
<script setup lang="ts">
import SiteDetailHelpTooltip from '../SiteDetailHelpTooltip.vue'
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ ports: SiteSecurityPresentation['portCheck']; waf: SiteSecurityPresentation['wafCanary'] }>()
const { t } = useI18n()
</script>
