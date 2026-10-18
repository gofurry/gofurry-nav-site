<template>
  <SiteChangeStream data-site-recent-changes :data-site-changes-state="state" variant="overview" title-id="site-recent-changes-title"
    :title="t('siteOverview.changes')" :items="displayItems" :state="state" :empty-text="t(state === 'unavailable' ? 'siteOverview.unavailable' : 'siteOverview.changesEmpty')" />
</template>
<script setup lang="ts">
import { computed } from 'vue'
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import SiteChangeStream from './SiteChangeStream.vue'
const props = defineProps<{ items: SiteOverviewPresentation['recentChanges']; state: SiteOverviewPresentation['changesState'] }>()
const { t } = useI18n()
const displayItems = computed(() => props.items.map(item => ({ ...item, when: item.when + (item.precise ? ' ' + t('siteOverview.utc') : '') })))
</script>
