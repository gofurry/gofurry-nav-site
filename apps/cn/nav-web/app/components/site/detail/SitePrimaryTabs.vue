<template>
  <div data-site-primary-tabs role="tablist" :aria-label="t('siteDetail.navigation')" class="site-detail-tabs sticky top-0 z-30 flex">
    <button
      v-for="(tab, index) in siteDetailTabs"
      :id="'site-tab-' + tab"
      :key="tab"
      :data-site-primary-tab="tab"
      role="tab"
      type="button"
      :aria-selected="active === tab"
      aria-controls="site-workspace"
      :tabindex="active === tab ? 0 : -1"
      class="site-detail-tab min-w-0 flex-1 sm:flex-none"
      @click="emit('select', tab)"
      @keydown="onKey($event, index)"
    >{{ t('siteDetail.tabs.' + tab) }}</button>
  </div>
</template>

<script setup lang="ts">
import { siteDetailTabs, type SiteDetailTab } from '~/utils/siteDetailRouteState'
defineProps<{ active: SiteDetailTab }>()
const emit = defineEmits<{ select: [tab: SiteDetailTab] }>()
const { t } = useI18n()
function onKey(event: KeyboardEvent, index: number) {
  const count = siteDetailTabs.length
  const next = event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count
    : event.key === 'Home' ? 0 : event.key === 'End' ? count - 1 : -1
  if (next < 0) return
  event.preventDefault()
  const tab = siteDetailTabs[next]!
  emit('select', tab)
  document.getElementById('site-tab-' + tab)?.focus({ preventScroll: true })
}
</script>
