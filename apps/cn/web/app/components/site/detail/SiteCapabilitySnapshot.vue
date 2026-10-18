<template>
  <section data-site-capability-snapshot data-site-overview-capability-composite :data-site-capabilities-state="state" class="min-w-0" aria-labelledby="site-capability-title">
    <div class="site-overview-section-head flex min-w-0 items-center gap-3">
      <h3 id="site-capability-title" class="site-overview-heading shrink-0">{{ t('siteOverview.capabilities') }}</h3>
      <span class="min-w-0 flex-1" aria-hidden="true" />
      <NuxtLink :to="insightsTo" data-site-overview-insights class="site-overview-ecosystem-link inline-flex shrink-0 items-center gap-1">
        {{ t('siteOverview.fullEcosystem') }}<PhArrowUpRight class="site-detail-icon" aria-hidden="true" />
      </NuxtLink>
    </div>
    <p v-if="state === 'unavailable'" data-site-capabilities-unavailable class="site-overview-empty mt-3">{{ t('siteOverview.unavailable') }}</p>
    <button v-if="state === 'unavailable'" data-site-insights-retry type="button" :disabled="retrying" class="site-detail-text-link mt-2" @click="emit('retry')">{{ t(retrying ? 'siteIntelligence.loading' : 'siteIntelligence.retry') }}</button>
    <div data-site-overview-capability-grid class="site-overview-capability-grid mt-5 grid min-w-0 gap-5 md:grid-cols-3">
      <section v-for="group in groups" :key="group.key" :data-site-capability-group="group.key" class="site-overview-capability-group min-w-0">
        <h4 class="site-overview-capability-group-title flex items-center gap-2">{{ group.label }}</h4>
        <dl class="mt-1 space-y-1.5">
          <div v-for="item in group.items" :key="item.key" :data-site-capability="item.key" :data-site-capability-state="item.state" class="site-overview-capability-row flex min-w-0 flex-wrap items-baseline justify-between gap-x-2 gap-y-1">
            <dt class="min-w-0 break-words">{{ item.label }}</dt>
            <dd class="site-overview-state inline-flex items-baseline gap-2" :data-tone="item.tone">
              <span aria-hidden="true" class="site-overview-dot shrink-0" />{{ item.stateLabel }}
            </dd>
          </div>
        </dl>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import type { RouteLocationRaw } from 'vue-router'
import { PhArrowUpRight } from '@phosphor-icons/vue'
defineProps<{ groups: SiteOverviewPresentation['capabilityGroups']; state: SiteOverviewPresentation['capabilityState']; insightsTo: RouteLocationRaw; retrying: boolean }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()
</script>
