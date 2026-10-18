<template>
  <aside data-site-target-context :aria-label="t('siteDetail.currentTarget')" class="site-detail-context min-w-0 self-start">
    <h2 class="site-detail-label mb-2">{{ t('siteDetail.currentTarget') }}</h2>
    <SiteTargetSelector :targets="presentation.targetList" :selected="selected" :display-target="presentation.target" @select="emit('select', $event)" />
    <p v-if="pending" data-site-target-pending role="status" class="site-detail-note mt-3 break-words">{{ t('siteDetail.loadingTarget', { target: selected }) }}</p>
    <div :aria-busy="pending" class="mt-3 min-w-0">
      <dl class="grid min-w-0 gap-x-5 gap-y-1.5 sm:grid-cols-3 xl:grid-cols-1">
        <div v-for="protocol in presentation.protocolStates" :key="protocol.protocol" :data-site-protocol="protocol.protocol" class="site-detail-protocol-row flex min-w-0 items-baseline justify-between gap-2">
          <dt class="site-detail-label site-detail-protocol-heading flex min-w-0 flex-wrap items-center gap-2">
            <span class="site-detail-status-dot shrink-0" :data-tone="protocol.tone" aria-hidden="true" />{{ protocol.label }}
            <span data-site-protocol-status :class="{ 'sr-only': protocol.success }"><span v-if="!protocol.success" aria-hidden="true">· </span>{{ protocol.statusLabel }}</span>
          </dt>
          <dd class="site-detail-protocol-value shrink-0 text-right" :data-tone="protocol.durationTone">{{ protocol.duration }}</dd>
        </div>
      </dl>
      <p class="site-detail-note mt-3 break-words">{{ t('siteDetail.observed') }} <time>{{ presentation.observedAt }}</time></p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import SiteTargetSelector from './SiteTargetSelector.vue'
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
defineProps<{ presentation: SiteTargetPresentation; pending: boolean; selected: string; siteScope: boolean }>()
const emit = defineEmits<{ select: [target: string] }>()
const { t } = useI18n()
</script>
