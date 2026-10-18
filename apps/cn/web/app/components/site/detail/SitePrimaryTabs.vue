<template>
  <div data-site-primary-tabs role="tablist" :aria-label="t('siteDetail.navigation')" class="site-detail-tabs sticky top-0 z-30 flex" :class="{ 'overflow-x-auto': hasSimilar }">
    <button
      v-for="(tab, index) in siteDetailTabs"
      :id="'site-tab-' + tab"
      :key="tab"
      :data-site-primary-tab="tab"
      role="tab"
      type="button"
      :aria-selected="!auxActive && active === tab"
      aria-controls="site-workspace"
      :tabindex="!auxActive && active === tab ? 0 : -1"
      class="site-detail-tab min-w-0" :class="hasSimilar ? 'shrink-0 whitespace-nowrap' : 'flex-1 sm:flex-none'"
      @click="emit('select', tab)"
      @keydown="onKey($event, index)"
    >{{ t('siteDetail.tabs.' + tab) }}</button>
    <button v-if="hasSimilar" id="site-tab-similar" data-site-aux-tab="similar" role="tab" type="button" aria-controls="site-similar-workspace"
      :aria-selected="auxActive" :tabindex="auxActive ? 0 : -1" class="site-detail-tab min-w-0 shrink-0 whitespace-nowrap xl:hidden"
      @click="emit('similar')" @keydown="onKey($event, 4)">{{ t('siteDetail.similarSites') }}</button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { siteDetailTabs, type SiteDetailTab } from '~/utils/siteDetailRouteState'
const props = defineProps<{ active: SiteDetailTab; hasSimilar?: boolean; auxActive?: boolean; desktop?: boolean }>()
const emit = defineEmits<{ select: [tab: SiteDetailTab]; similar: [] }>()
const { t } = useI18n()
function onKey(event: KeyboardEvent, index: number) {
  const tabs = [...siteDetailTabs, ...(props.hasSimilar && !props.desktop ? ['similar' as const] : [])]
  const count = tabs.length
  const next = event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count
    : event.key === 'Home' ? 0 : event.key === 'End' ? count - 1 : -1
  if (next < 0) return
  event.preventDefault()
  const tab = tabs[next]!
  if (tab === 'similar') emit('similar')
  else emit('select', tab)
  const button = document.getElementById('site-tab-' + tab)
  button?.focus({ preventScroll: true })
  if (props.hasSimilar && !props.desktop) button?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'instant' })
}
</script>
