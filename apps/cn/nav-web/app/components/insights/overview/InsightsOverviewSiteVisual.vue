<template>
  <NuxtLink :to="localePath(siteEntityPath(item.site_id))" class="overview-site-visual flex min-w-0 flex-col items-center gap-2" :aria-label="$t('insights.overviewVisuals.viewSite', { name: item.name })" data-overview-site-visual :data-site-id="item.site_id">
    <ManagedAssetImage v-if="!failed" :object-key="item.visual.asset" fallback="" :alt="item.name" width="48" height="48" loading="lazy" class="overview-site-visual__icon block" @exhausted="failed = true" />
    <span v-else class="overview-site-visual__initials flex items-center justify-center" role="img" :aria-label="$t('insights.overviewVisuals.iconUnavailable', { name: item.name })" data-site-logo-fallback>{{ initials }}</span>
    <span class="overview-site-visual__name block w-full">{{ item.name }}</span>
  </NuxtLink>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import ManagedAssetImage from '@/components/common/ManagedAssetImage.vue'
import type { InsightSiteVisual } from '@/types/insights'
import { siteEntityPath } from '@/utils/siteRoutes'
const props = defineProps<{ item: InsightSiteVisual }>()
const localePath = useLocalePath()
const failed = ref(false)
const initials = computed(() => Array.from(props.item.name.trim()).slice(0, 2).join(''))
</script>
