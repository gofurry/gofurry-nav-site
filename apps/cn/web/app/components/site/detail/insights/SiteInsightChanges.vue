<template>
  <SiteChangeStream data-site-insight-changes :data-site-changes-state="state" variant="insights" title-id="site-insight-changes-title"
    :title="t('siteIntelligence.recentChanges')" :items="items" :state="state" :empty-text="t(state === 'unavailable' ? 'siteIntelligence.unavailable' : 'siteIntelligence.noChanges')" />
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import SiteChangeStream from '../SiteChangeStream.vue'
defineProps<{ items: SiteInsightsPresentation['changes']; state: SiteInsightsPresentation['changesState'] }>()
const { t } = useI18n()
</script>
