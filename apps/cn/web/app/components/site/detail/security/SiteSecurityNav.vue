<template>
  <div data-site-security-nav role="tablist" :aria-label="t('siteSecurity.navigation')" class="site-detail-subnav flex min-w-0 overflow-x-auto">
    <button v-for="(view, index) in siteSecurityViews" :id="'security-tab-' + view" :key="view" :data-site-security-tab="view"
      type="button" role="tab" :aria-selected="active === view" aria-controls="site-security-panel" :tabindex="active === view ? 0 : -1"
      class="site-detail-segment shrink-0" @click="emit('select', view)" @keydown="onKey($event, index)">
      {{ t('siteSecurity.views.' + view) }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { siteSecurityViews, type SiteSecurityView } from '~/utils/siteDetailRouteState'
defineProps<{ active: SiteSecurityView }>()
const emit = defineEmits<{ select: [view: SiteSecurityView] }>()
const { t } = useI18n()
function onKey(event: KeyboardEvent, index: number) {
  const count = siteSecurityViews.length
  const next = event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count
    : event.key === 'Home' ? 0 : event.key === 'End' ? count - 1 : -1
  if (next < 0) return
  event.preventDefault()
  const view = siteSecurityViews[next]!
  emit('select', view)
  const button = document.getElementById('security-tab-' + view)
  button?.focus({ preventScroll: true })
  if (button?.parentElement) button.parentElement.scrollLeft = button.offsetLeft - button.parentElement.offsetLeft
}
</script>
