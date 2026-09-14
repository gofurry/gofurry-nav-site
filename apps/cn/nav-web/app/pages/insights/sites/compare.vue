<template>
  <div class="insights-page insights-workspace-page insights-compare-page" data-site-compare :data-compare-count="selectedIDs.length"
    :data-compare-status="invalidURL ? 'invalid' : error ? 'error' : compare?.status || (ready ? 'loading' : 'builder')">
    <main class="insights-container">
      <EcosystemNavigation context="site" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.sites.title')" :title="$t('insights.siteCompare.title')" :description="$t('insights.siteCompare.description')" />
      <div v-if="invalidURL" class="insight-compare-invalid" role="alert">
        <p>{{ $t('insights.compare.invalid') }}</p><button type="button" @click="updateSelection([])">{{ $t('insights.comparePicker.reset') }}</button>
      </div>
      <InsightComparePicker domain="site" :selected-ids="selectedIDs" :entities="entities" @change="updateSelection" />
      <section class="insight-compare-result" :aria-busy="pending" aria-live="polite">
        <p v-if="invalidURL" class="insights-workspace-note">{{ $t('insights.comparePicker.restart') }}</p>
        <p v-else-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
        <p v-else-if="pending && ready" class="insights-empty-state">{{ $t('insights.comparePicker.loadingComparison') }}</p>
        <p v-else-if="compare?.status === 'insufficient_data'" class="insights-empty-state">{{ $t('insights.compare.insufficientData') }}</p>
        <template v-else-if="compare?.status === 'ready'">
          <h2>{{ $t('insights.comparePicker.resultsTitle') }}</h2>
          <div class="insight-compare-horizons">
            <span>{{ $t('insights.compare.commonSnapshot', { date: compare.as_of || '—' }) }}</span>
          </div>
          <InsightCompareMatrix domain="site" :entities="entities" :groups="groups" :label="$t('insights.siteCompare.title')" data-compare-result />
        </template>
      </section>
      <InsightWorkspaceDisclosure :title="$t('insights.compare.aboutTitle')">
        <p>{{ $t('insights.siteCompare.about') }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'
import InsightComparePicker from '@/components/insights/compare/InsightComparePicker.vue'
import InsightCompareMatrix from '@/components/insights/compare/InsightCompareMatrix.vue'
import type { CompareMatrixGroup } from '@/types/insightCompare'
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSiteCompare } from '@/services/nav'
import type { NavInsightMetricKey, SiteCompareItem, SiteInsightCapabilityState } from '@/types/insights'
import { insightCompareReady, parseInsightCompareIDs } from '@/utils/insightCompare'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const capabilityKeys = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'] as const satisfies readonly NavInsightMetricKey[]
const parsedIDs = computed(() => parseInsightCompareIDs(route.query.ids))
const selectedIDs = computed(() => parsedIDs.value ?? [])
const invalidURL = computed(() => parsedIDs.value === null)
const ready = computed(() => insightCompareReady(selectedIDs.value))
const requestKey = computed(() => selectedIDs.value.join(','))
const { data: snapshot, error, pending } = await useAsyncData('site-compare', async () => {
  const selection = requestKey.value
  const result = ready.value ? await getSiteCompare(selectedIDs.value) : null
  return { selection, result }
}, { watch: [requestKey] })
const compare = computed(() => snapshot.value?.selection === requestKey.value ? snapshot.value.result : null)
const orderedItems = computed(() => selectedIDs.value.flatMap(id => {
  const item = compare.value?.sites.find(item => item.site.id === id)
  return item ? [item] : []
}))
const entities = computed(() => orderedItems.value.map(item => item.site))

function updateSelection(ids: number[]) {
  if (parseInsightCompareIDs(ids.join(',')) === null) return
  void router.push({ path: route.path, query: { ...route.query, ids: ids.length ? ids.join(',') : undefined } })
}

function siteCapability(item: SiteCompareItem, key: NavInsightMetricKey): SiteInsightCapabilityState {
  return item.capabilities.find(capability => capability.key === key)?.state ?? 'unavailable'
}

function certificateVerification(item: SiteCompareItem) {
  if (!item.certificate) return t('insights.siteCompare.noValidCertificate')
  if (item.certificate.verified === true) return t('insights.certificateIntelligence.verified')
  if (item.certificate.verified === false) {
    const issue = item.certificate.verification_issue
    return issue ? `${t('insights.certificateIntelligence.failed')} · ${t(`insights.certificateIntelligence.values.${issue}`)}` : t('insights.certificateIntelligence.failed')
  }
  return t('insights.entity.states.unknown')
}

function certificateExpiry(item: SiteCompareItem) {
  const status = item.certificate?.expiry_status
  return status ? t(`insights.certificateIntelligence.values.${status}`) : '—'
}

const groups = computed<CompareMatrixGroup[]>(() => [
  { key: 'capabilities', label: t('insights.comparePicker.groups.capabilities'), rows: capabilityKeys.map(key => ({
    key, label: t('insights.metrics.' + key + '.name'), cells: orderedItems.value.map(item => ({
      text: t('insights.entity.states.' + siteCapability(item, key)), state: siteCapability(item, key),
    })),
  })) },
  { key: 'certificate', label: t('insights.comparePicker.groups.certificate'), rows: [
    { key: 'target', label: t('insights.siteCompare.primaryTarget'), cells: orderedItems.value.map(item => ({ text: item.certificate?.target || '—' })) },
    { key: 'verification', label: t('insights.siteCompare.certificateVerification'), cells: orderedItems.value.map(item => ({ text: certificateVerification(item) })) },
    { key: 'expiry', label: t('insights.siteCompare.certificateExpiry'), cells: orderedItems.value.map(item => ({ text: certificateExpiry(item) })) },
  ] },
])

useSeoMeta({
  title: () => `${t('insights.siteCompare.title')} | GoFurry`,
  description: () => t('insights.siteCompare.description'),
  robots: 'noindex, follow',
})
</script>
