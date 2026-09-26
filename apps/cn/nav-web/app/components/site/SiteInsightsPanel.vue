<template>
  <section class="site-insights-panel" data-site-insights :data-site-insights-state="unavailable || !insights ? 'unavailable' : insights.capabilities.length || insights.recent_changes.length ? 'ready' : 'empty'">
    <div class="entity-insights-heading">
      <div>
        <p class="entity-insights-eyebrow">{{ $t('insights.entity.siteEyebrow') }}</p>
        <h2>{{ $t('insights.entity.siteTitle') }}</h2>
        <p>{{ $t('insights.entity.siteDescription') }}</p>
      </div>
      <NuxtLink :to="localePath('/insights/sites')" class="entity-insights-back-link">
        {{ $t('insights.entity.backToSites') }}
      </NuxtLink>
    </div>

    <p v-if="unavailable || !insights" class="entity-insights-empty" data-site-insights-unavailable>
      {{ $t('insights.emptyStates.unavailable') }}
    </p>
    <template v-else>
      <div class="site-insights-capabilities">
        <article
          v-for="capability in capabilities"
          :key="capability.key"
          class="site-insights-capability"
          :data-capability-key="capability.key"
          :data-capability-state="capability.state ?? 'missing'"
        >
          <div class="site-insights-capability__heading">
            <h3>{{ $t(capability.labelKey) }}</h3>
            <span :class="`site-insights-state site-insights-state--${stateTone(capability.state)}`">
              {{ capability.state ? $t(`insights.entity.states.${capability.state}`) : $t('insights.emptyStates.unavailable') }}
            </span>
          </div>
          <dl>
            <div>
              <dt>{{ $t('insights.entity.ecosystemValue') }}</dt>
              <dd>{{ formatPercent(capability.ecosystem.value) }}</dd>
            </div>
            <div>
              <dt>{{ $t('insights.entity.ecosystemCoverage') }}</dt>
              <dd>{{ formatPercent(capability.ecosystem.coverage) }}</dd>
            </div>
          </dl>
          <p>{{ $t('insights.entity.sameDayAsOf', { date: capability.as_of || '—' }) }}</p>
        </article>
      </div>

      <InsightsEntityTimeline :items="insights.recent_changes" />
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import InsightsEntityTimeline from '@/components/insights/InsightsEntityTimeline.vue'
import type { SiteInsightCapabilityKey, SiteInsightCapabilityState, SiteInsights } from '@/types/insights'
import { siteCapabilityRegistry } from '~/utils/siteCapabilityRegistry'

interface CapabilityView {
  key: SiteInsightCapabilityKey
  labelKey: string
  as_of: string | null
  state: SiteInsightCapabilityState | null
  ecosystem: {
    value: number | null
    coverage: number | null
  }
}

const props = defineProps<{
  insights: SiteInsights | null
  unavailable?: boolean
}>()

const localePath = useLocalePath()
const previewCapabilities = siteCapabilityRegistry.filter(capability => capability.preview)
const capabilities = computed<CapabilityView[]>(() => {
  const byKey = new Map((props.insights?.capabilities ?? []).map(capability => [capability.key, capability]))
  return previewCapabilities.map(({ key, labelKey }) => ({
    key, labelKey,
    as_of: null,
    state: null,
    ecosystem: { value: null, coverage: null },
    ...byKey.get(key),
  }))
})

function stateTone(state: SiteInsightCapabilityState | null) {
  if (state === 'supported') return 'positive'
  if (state === 'unsupported') return 'negative'
  return 'ambiguous'
}

function formatPercent(value: number | null) {
  return value === null || !Number.isFinite(value) ? '—' : `${(value * 100).toFixed(1)}%`
}
</script>
