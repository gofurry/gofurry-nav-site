<template>
  <div v-if="visibleSites.length" class="site-icon-strip flex items-center gap-2 overflow-x-auto overflow-y-hidden">
    <div v-for="item in visibleSites" :key="item.id" class="shrink-0">
      <a
          :href="toExternalUrl(item.url)"
          target="_blank"
          rel="noopener noreferrer"
          class="site-icon-strip__link flex h-10 w-10 items-center justify-center"
          :title="item.name"
          @click="handleVisit(item)"
      >
        <img
            v-if="!failedIcons[item.id]"
            :src="`https://favicon.im/${toExternalUrl(item.url)}?larger=true`"
            :alt="item.name"
            class="site-icon-strip__image"
            loading="lazy"
            @error="markIconFailed(item.id)"
        />
        <div
            v-else
            class="site-icon-strip__fallback"
        >
          {{ item.name.slice(0, 1).toUpperCase() }}
        </div>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'

export interface SiteStripItem {
  id: string
  name: string
  url: string
}

const props = withDefaults(defineProps<{
  sites: SiteStripItem[]
  maxItems?: number
}>(), {
  maxItems: 8,
})

const visibleSites = computed(() => props.sites.slice(0, props.maxItems))
const failedIcons = ref<Record<string, boolean>>({})

function markIconFailed(id: string) {
  failedIcons.value = {
    ...failedIcons.value,
    [id]: true,
  }
}

function handleVisit(site: SiteStripItem) {
  recordRecentSite(site)
}
</script>
