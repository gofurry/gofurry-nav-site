<template>
  <div class="space-y-6">
    <section data-site-security-headers class="site-detail-surface">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h3 class="site-overview-title">{{ t('siteSecurity.headers') }}</h3>
        <NuxtLink data-site-security-raw-headers :to="rawHeadersTo" class="site-detail-text-link inline-flex items-center gap-1">{{ t('siteSecurity.rawHeaders') }}<PhArrowRight class="site-detail-icon" aria-hidden="true" /></NuxtLink>
      </div>
      <p class="site-detail-note mt-1">{{ t('siteSecurity.headerHint') }}</p>
      <p class="site-detail-note mt-1">{{ t('siteSecurity.lastObserved') }} · {{ headers.observed }}<span v-if="headers.stale"> · {{ t('siteSecurity.stale') }}</span></p>
      <p v-if="headers.truncated" class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <div class="mt-4 space-y-1.5">
      <div v-for="item in headers.rows" :key="item.key" :data-site-security-header="item.key" :data-state="item.state" class="site-security-row min-w-0">
        <div class="flex min-w-0 flex-wrap items-baseline justify-between gap-2"><h4 class="site-detail-label min-w-0 break-words">{{ item.name }}</h4><p :data-tone="item.tone">{{ item.label }}</p></div>
        <div class="mt-2 min-w-0"><p class="whitespace-pre-wrap break-all">{{ item.value }}</p>
          <details v-if="item.details.length" class="site-security-disclosure mt-2"><summary>{{ t('siteSecurity.collectorSummary') }}</summary><SiteObservationFacts :items="item.details" class="mt-2" /></details>
        </div>
      </div>
      </div>
    </section>
    <section data-site-security-txt :data-state="securityTxt.state" class="site-detail-surface site-detail-surface--secondary">
      <div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="site-overview-title">security.txt</h3><p :data-tone="securityTxt.tone">{{ securityTxt.label }}</p></div>
      <p v-if="securityTxt.stale" class="site-detail-note mt-1">{{ t('siteSecurity.stale') }}</p>
      <p v-if="securityTxt.truncated" class="site-detail-note mt-1">{{ t('siteSecurity.truncated') }}</p>
      <ul v-if="securityTxt.validationLabels.length" data-site-security-txt-validation class="site-security-attention mt-3 space-y-1.5"><li v-for="error in securityTxt.validationLabels" :key="error" class="flex items-start gap-2 break-all"><PhWarning class="site-detail-icon mt-1 shrink-0" data-tone="warning" aria-hidden="true" />{{ error }}</li></ul>
      <SiteObservationFacts :items="securityTxt.facts" class="mt-3" />
      <details data-site-security-txt-details class="site-security-disclosure mt-3">
        <summary>{{ t('siteSecurity.collectorSummary') }}</summary>
        <SiteObservationFacts v-if="securityTxt.errors.length" :items="securityTxt.errors" class="mt-2" />
        <SiteObservationFacts :items="securityTxt.metaFacts" class="mt-2" />
      </details>
    </section>
  </div>
</template>
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import { PhArrowRight, PhWarning } from '@phosphor-icons/vue'
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ headers: SiteSecurityPresentation['headers']; securityTxt: SiteSecurityPresentation['securityTxt']; rawHeadersTo: RouteLocationRaw }>()
const { t } = useI18n()
</script>
