<template>
  <div data-site-overview class="site-overview min-w-0">
    <SiteOverviewHealth :health="presentation.health" />
    <section v-if="presentation.attention.length" data-site-overview-attention class="site-overview-attention mt-5" aria-labelledby="site-attention-title">
      <h3 id="site-attention-title" class="site-overview-title">{{ t('siteOverview.attention') }}</h3>
      <ul class="mt-2 space-y-2">
        <li v-for="item in presentation.attention" :key="item.key" :data-site-attention-target="item.target" class="site-overview-attention__item break-words">
          {{ item.message }}
        </li>
      </ul>
    </section>
    <div data-site-overview-columns class="mt-6 grid min-w-0 gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <SiteCapabilitySnapshot :groups="presentation.capabilityGroups" :state="presentation.capabilityState" />
      <SiteRecentChanges :items="presentation.recentChanges" :state="presentation.changesState" />
    </div>
    <NuxtLink :to="insightsTo" data-site-overview-insights class="gf-button gf-button--ghost mt-5">
      {{ t('siteOverview.fullEcosystem') }} <span aria-hidden="true">→</span>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import SiteOverviewHealth from './SiteOverviewHealth.vue'
import SiteCapabilitySnapshot from './SiteCapabilitySnapshot.vue'
import SiteRecentChanges from './SiteRecentChanges.vue'
defineProps<{ presentation: SiteOverviewPresentation; insightsTo: RouteLocationRaw }>()
const { t } = useI18n()
</script>
