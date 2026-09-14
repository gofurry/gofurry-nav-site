<template>
  <div class="insights-page insights-workspace-page" data-certificate-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="site" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.sites.title')" :title="$t('insights.certificateIntelligence.title')" :description="$t('insights.certificateIntelligence.description')" />
      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="!overview?.as_of" class="insights-empty-state">{{ $t('insights.certificateIntelligence.empty') }}</p>
      <template v-else>
        <div class="insights-certificate-status">
          <section class="insights-workspace-summary">
            <h2>{{ $t('insights.certificateIntelligence.verificationTitle') }}</h2>
            <p class="insights-certificate-verification"><strong>{{ overview.verification.verified }}</strong><span>/ {{ overview.verification.known }} {{ $t('insights.certificateIntelligence.known') }}</span></p>
            <dl class="insights-workspace-stats insights-workspace-stats--compact">
              <div><dt>{{ $t('insights.certificateIntelligence.verified') }}</dt><dd>{{ percent(overview.verification.known ? overview.verification.verified / overview.verification.known : null) }}</dd></div>
              <div><dt>{{ $t('insights.certificateIntelligence.failed') }}</dt><dd>{{ overview.verification.failed }}</dd></div>
              <div><dt>{{ $t('insights.certificateIntelligence.coverage') }}</dt><dd>{{ percent(overview.verification.coverage) }}</dd></div>
            </dl>
          </section>
          <section class="insights-workspace-summary">
            <h2>{{ $t('insights.certificateIntelligence.expiryTitle') }}</h2>
            <dl class="insights-expiry-risk">
              <div v-for="key in expiryKeys" :key="key" :data-expiry="key"><dt>{{ $t('insights.certificateIntelligence.values.' + key) }}</dt><dd>{{ overview.expiry[key] }}</dd></div>
            </dl>
            <p class="insights-workspace-meta">{{ $t('insights.certificateIntelligence.coverage') }} {{ percent(overview.expiry.coverage) }}</p>
          </section>
        </div>
        <section class="insights-workspace-quality">
          <h2>{{ $t('insights.workspace.observationQuality') }}</h2>
          <dl class="insights-workspace-secondary"><div v-for="key in qualityKeys" :key="key"><dt>{{ $t('insights.workspace.' + key) }}</dt><dd>{{ overview.quality[key] }}</dd></div></dl>
        </section>
        <section class="insights-workspace-section" data-expiry-attention>
          <h2>{{ $t('insights.certificateIntelligence.attentionTitle') }}</h2>
          <InsightRiskList :items="overview.expiry_attention" mode="expiry" />
        </section>
        <section class="insights-workspace-section" data-verification-issues>
          <h2>{{ $t('insights.certificateIntelligence.issuesTitle') }}</h2>
          <InsightRiskList :items="overview.verification_issues" mode="verification" />
        </section>
      </template>
      <InsightWorkspaceDisclosure :title="$t('insights.certificateIntelligence.aboutTitle')">
        <p>{{ $t('insights.certificateIntelligence.about') }}</p>
        <p v-if="overview?.as_of">{{ $t('insights.certificateIntelligence.asOf', { date: overview.as_of, reference: formatWorkspaceTimestamp(overview.reference_at, locale) }) }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'

import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { useI18n } from 'vue-i18n'
import InsightRiskList from '@/components/insights/workspace/InsightRiskList.vue'
import { formatWorkspaceTimestamp } from '@/utils/insightWorkspace'
import { getNavCertificateInsightsOverview } from '@/services/nav'

const { locale, t } = useI18n()
const { data: overview, error } = await useAsyncData('site-certificate-intelligence', () => getNavCertificateInsightsOverview(20))

useSeoMeta({
  title: () => `${t('insights.certificateIntelligence.title')} | GoFurry`,
  description: () => t('insights.certificateIntelligence.description'),
  ogTitle: () => `${t('insights.certificateIntelligence.title')} | GoFurry`,
  ogDescription: () => t('insights.certificateIntelligence.description'),
})

const expiryKeys = ['expired', 'expires_within_7d', 'expires_in_8_30d', 'later'] as const
const qualityKeys = ['not_applicable', 'stale', 'not_probed', 'probe_failed', 'unknown'] as const

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

</script>
