<template>
  <div data-site-observation-nav role="tablist" :aria-label="t('siteObservation.navigation')" class="site-observation-nav flex min-w-0 overflow-x-auto">
    <button v-for="(view, index) in siteObservationViews" :id="'observation-tab-' + view" :key="view" :data-site-observation-tab="view"
      type="button" role="tab" :aria-selected="active === view" aria-controls="site-observation-panel" :tabindex="active === view ? 0 : -1"
      class="site-observation-tab shrink-0" @click="emit('select', view)" @keydown="onKey($event, index)">
      {{ t('siteObservation.views.' + view) }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { siteObservationViews, type SiteObservationView } from '~/utils/siteDetailRouteState'
defineProps<{ active: SiteObservationView }>()
const emit = defineEmits<{ select: [view: SiteObservationView] }>()
const { t } = useI18n()
function onKey(event: KeyboardEvent, index: number) {
  const count = siteObservationViews.length
  const next = event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count
    : event.key === 'Home' ? 0 : event.key === 'End' ? count - 1 : -1
  if (next < 0) return
  event.preventDefault()
  const view = siteObservationViews[next]!
  emit('select', view)
  const button = document.getElementById('observation-tab-' + view)
  button?.focus({ preventScroll: true })
  if (button?.parentElement) button.parentElement.scrollLeft = button.offsetLeft - button.parentElement.offsetLeft
}
</script>
