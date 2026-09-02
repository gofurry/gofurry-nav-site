<template>
  <div class="insights-page insights-intelligence-page" data-certificate-intelligence>
    <GoFurryGridBackground :fixed="false" profile="light" />
    <main class="insights-container">
      <InsightsNav />
      <SiteIntelligenceNav />
      <header class="insights-hero insights-hero--domain">
        <p class="insights-eyebrow">TLS CERTIFICATE INTELLIGENCE</p>
        <h1>{{ $t('insights.certificateIntelligence.title') }}</h1>
        <p>{{ $t('insights.certificateIntelligence.description') }}</p>
      </header>

      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="!overview?.as_of" class="insights-empty-state">{{ $t('insights.certificateIntelligence.empty') }}</p>
      <template v-else>
        <section class="intelligence-panel">
          <h2>{{ $t('insights.certificateIntelligence.verificationTitle') }}</h2>
          <div class="intelligence-stats intelligence-stats--wide">
            <article><span>{{ $t('insights.certificateIntelligence.verified') }}</span><strong>{{ overview.verification.verified }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.failed') }}</span><strong>{{ overview.verification.failed }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.known') }}</span><strong>{{ overview.verification.known }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.coverage') }}</span><strong>{{ percent(overview.verification.coverage) }}</strong></article>
          </div>
          <p class="insights-data-note">
            {{ $t('insights.certificateIntelligence.quality', overview.quality) }}
          </p>
        </section>

        <section class="intelligence-panel">
          <h2>{{ $t('insights.certificateIntelligence.expiryTitle') }}</h2>
          <div class="intelligence-stats intelligence-stats--wide">
            <article><span>{{ $t('insights.certificateIntelligence.expired') }}</span><strong>{{ overview.expiry.expired }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.within7d') }}</span><strong>{{ overview.expiry.expires_within_7d }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.in8to30d') }}</span><strong>{{ overview.expiry.expires_in_8_30d }}</strong></article>
            <article><span>{{ $t('insights.certificateIntelligence.coverage') }}</span><strong>{{ percent(overview.expiry.coverage) }}</strong></article>
          </div>
        </section>

        <CertificateInsightTable
          :title="$t('insights.certificateIntelligence.attentionTitle')"
          :items="overview.expiry_attention"
          mode="expiry"
        />
        <CertificateInsightTable
          :title="$t('insights.certificateIntelligence.issuesTitle')"
          :items="overview.verification_issues"
          mode="verification"
        />
      </template>

      <section class="insights-data-info">
        <h2>{{ $t('insights.certificateIntelligence.aboutTitle') }}</h2>
        <p>{{ $t('insights.certificateIntelligence.about') }}</p>
        <p v-if="overview?.as_of">
          {{ $t('insights.certificateIntelligence.asOf', { date: overview.as_of, reference: formatTimestamp(overview.reference_at) }) }}
        </p>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import CertificateInsightTable from '@/components/insights/CertificateInsightTable.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import InsightsNav from '@/components/insights/InsightsNav.vue'
import SiteIntelligenceNav from '@/components/insights/SiteIntelligenceNav.vue'
import { getNavCertificateInsightsOverview } from '@/services/nav'

const { locale, t } = useI18n()
const { data: overview, error } = await useAsyncData('site-certificate-intelligence', () => getNavCertificateInsightsOverview(20))

useSeoMeta({
  title: () => `${t('insights.certificateIntelligence.title')} | GoFurry`,
  description: () => t('insights.certificateIntelligence.description'),
  ogTitle: () => `${t('insights.certificateIntelligence.title')} | GoFurry`,
  ogDescription: () => t('insights.certificateIntelligence.description'),
})

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

function formatTimestamp(value: string | null) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', { dateStyle: 'medium', timeZone: 'UTC' }).format(new Date(value))
}
</script>
