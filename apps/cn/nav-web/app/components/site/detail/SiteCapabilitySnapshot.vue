<template>
  <section data-site-capability-snapshot :data-site-capabilities-state="state" class="min-w-0" aria-labelledby="site-capability-title">
    <h3 id="site-capability-title" class="site-overview-title">{{ t('siteOverview.capabilities') }}</h3>
    <p v-if="state === 'unavailable'" data-site-capabilities-unavailable class="site-overview-empty mt-3">{{ t('siteOverview.unavailable') }}</p>
    <section v-for="group in groups" :key="group.key" :data-site-capability-group="group.key" class="site-capability-group mt-4">
      <h4 class="site-detail-label">{{ group.label }}</h4>
      <dl class="mt-1">
        <div v-for="item in group.items" :key="item.key" :data-site-capability="item.key" :data-site-capability-state="item.state" class="site-capability-row flex min-w-0 flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
          <dt class="min-w-0 break-words">{{ item.label }}</dt>
          <dd class="site-overview-state inline-flex items-baseline gap-2" :data-tone="item.tone">
            <span aria-hidden="true" class="site-overview-dot shrink-0" />{{ item.stateLabel }}
          </dd>
        </div>
      </dl>
    </section>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
defineProps<{ groups: SiteOverviewPresentation['capabilityGroups']; state: SiteOverviewPresentation['capabilityState'] }>()
const { t } = useI18n()
</script>
