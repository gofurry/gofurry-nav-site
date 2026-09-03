<template>
  <div
    class="insights-page insights-compare-page"
    data-site-compare
    :data-compare-count="selectedIDs.length"
    :data-compare-status="compare?.status || (ready ? 'loading' : 'builder')"
  >
    <main class="insights-container">
      <EcosystemNavigation context="site" />
      <h1 class="sr-only">{{ $t('insights.siteCompare.title') }}</h1>

      <section class="compare-builder">
        <h2>{{ $t('insights.compare.builderTitle') }}</h2>
        <p>{{ builderHint }}</p>
        <form class="compare-builder__form" @submit.prevent="applySelection">
          <label>
            <span>{{ $t('insights.compare.idsLabel') }}</span>
            <input v-model="input" inputmode="numeric" autocomplete="off" :placeholder="$t('insights.compare.idsPlaceholder')" />
          </label>
          <button type="submit">{{ $t('insights.compare.apply') }}</button>
        </form>
        <p v-if="inputError || invalidURL" class="compare-builder__error">{{ $t('insights.compare.invalid') }}</p>
      </section>

      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="compare?.status === 'insufficient_data'" class="insights-empty-state">{{ $t('insights.compare.insufficientData') }}</p>

      <section v-else-if="compare?.status === 'ready'" class="compare-result" data-compare-result>
        <p class="insights-data-note">{{ $t('insights.compare.commonSnapshot', { date: compare.as_of || '—' }) }}</p>
        <div class="compare-table-wrap">
          <table class="compare-table">
            <thead>
              <tr>
                <th>{{ $t('insights.compare.fact') }}</th>
                <th v-for="item in compare.sites" :key="item.site.id" :data-compare-entity-id="item.site.id">
                  <NuxtLink :to="localePath(`/site/${item.site.id}`)">{{ item.site.name || `#${item.site.id}` }}</NuxtLink>
                  <small>#{{ item.site.id }}</small>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="key in capabilityKeys" :key="key">
                <th>{{ $t(`insights.metrics.${key}.title`) }}</th>
                <td v-for="item in compare.sites" :key="item.site.id" :data-capability-state="siteCapability(item, key)">
                  {{ $t(`insights.entity.states.${siteCapability(item, key)}`) }}
                </td>
              </tr>
              <tr>
                <th>{{ $t('insights.siteCompare.primaryTarget') }}</th>
                <td v-for="item in compare.sites" :key="item.site.id">{{ item.certificate?.target || '—' }}</td>
              </tr>
              <tr>
                <th>{{ $t('insights.siteCompare.certificateVerification') }}</th>
                <td v-for="item in compare.sites" :key="item.site.id">{{ certificateVerification(item) }}</td>
              </tr>
              <tr>
                <th>{{ $t('insights.siteCompare.certificateExpiry') }}</th>
                <td v-for="item in compare.sites" :key="item.site.id">{{ certificateExpiry(item) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="insights-data-info">
        <h2>{{ $t('insights.compare.aboutTitle') }}</h2>
        <p>{{ $t('insights.siteCompare.about') }}</p>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSiteCompare } from '@/services/nav'
import type { NavInsightMetricKey, SiteCompareItem, SiteInsightCapabilityState } from '@/types/insights'
import { insightCompareReady, parseInsightCompareIDs } from '@/utils/insightCompare'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { t } = useI18n()
const capabilityKeys = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'] as const satisfies readonly NavInsightMetricKey[]
const parsedIDs = computed(() => parseInsightCompareIDs(route.query.ids))
const selectedIDs = computed(() => parsedIDs.value ?? [])
const invalidURL = computed(() => parsedIDs.value === null)
const ready = computed(() => insightCompareReady(selectedIDs.value))
const requestKey = computed(() => selectedIDs.value.join(','))
const input = ref(typeof route.query.ids === 'string' ? route.query.ids : '')
const inputError = ref(false)

watch(() => route.query.ids, value => { input.value = typeof value === 'string' ? value : '' })

const { data: compare, error } = await useAsyncData('site-compare', async () => {
  if (!ready.value) return null
  return getSiteCompare(selectedIDs.value)
}, { watch: [requestKey] })

const builderHint = computed(() => selectedIDs.value.length === 0
  ? t('insights.compare.emptyHint')
  : selectedIDs.value.length === 1
    ? t('insights.compare.oneHint')
    : t('insights.compare.readyHint', { count: selectedIDs.value.length }))

function applySelection() {
  const ids = parseInsightCompareIDs(input.value)
  inputError.value = ids === null
  if (ids === null) return
  void router.push({ path: route.path, query: ids.length ? { ids: ids.join(',') } : {} })
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

useSeoMeta({
  title: () => `${t('insights.siteCompare.title')} | GoFurry`,
  description: () => t('insights.siteCompare.description'),
  robots: 'noindex, follow',
})
</script>
