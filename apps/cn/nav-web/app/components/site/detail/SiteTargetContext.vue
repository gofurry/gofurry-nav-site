<template>
  <aside data-site-target-context :aria-label="t('siteDetail.currentTarget')" class="site-detail-context min-w-0 self-start xl:sticky xl:order-2">
    <div class="flex min-w-0 items-start justify-between gap-3 xl:flex-col">
      <div class="min-w-0">
        <h2 class="site-detail-label">{{ t('siteDetail.currentTarget') }}</h2>
        <p class="site-detail-target-name mt-1 break-all">{{ presentation.target }}</p>
        <p v-if="presentation.targetRelation" class="site-detail-note mt-1 break-words">{{ presentation.targetRelation }}</p>
      </div>
      <span class="site-detail-status shrink-0" :data-tone="presentation.tone">{{ presentation.statusLabel }}</span>
    </div>
    <div class="mt-3"><SiteTargetSelector :targets="presentation.targetList" :selected="selected" @select="emit('select', $event)" /></div>
    <p v-if="pending" data-site-target-pending role="status" class="site-detail-note mt-3 break-words">{{ t('siteDetail.loadingTarget', { target: selected }) }}</p>
    <div :aria-busy="pending" class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2 xl:grid-cols-1">
      <dl class="grid grid-cols-3 gap-2 xl:grid-cols-1">
        <div v-for="protocol in presentation.protocolStates" :key="protocol.protocol" :data-site-protocol="protocol.protocol" class="min-w-0 xl:flex xl:justify-between xl:gap-2">
          <dt class="site-detail-label">{{ protocol.label }}</dt>
          <dd class="site-detail-protocol-value break-words xl:text-right" :data-tone="protocol.tone">
            {{ protocol.statusLabel }} <span class="site-detail-note block">{{ protocol.duration }}</span>
          </dd>
        </div>
      </dl>
      <div class="min-w-0">
        <div class="flex flex-wrap items-baseline gap-x-2 xl:block">
          <div class="site-detail-label">{{ t('siteDetail.observed') }}</div>
          <p class="site-detail-note break-words xl:mt-1">{{ presentation.observedAt }}</p>
        </div>
        <div data-site-infrastructure class="hidden xl:block">
          <div class="site-detail-label mt-3">{{ t('siteDetail.infrastructure') }}</div>
          <ul v-if="presentation.edgeProviderHints.length" class="site-detail-note mt-1 space-y-2">
            <li v-for="hint in presentation.edgeProviderHints" :key="hint.provider + hint.type" class="break-words" :title="hint.evidence">
              {{ hint.provider }} · {{ hint.type }}<br>{{ t('siteDetail.confidence') }}: {{ hint.confidence }}
            </li>
          </ul>
          <p v-else class="site-detail-note mt-1">{{ t('siteDetail.noInfrastructure') }}</p>
        </div>
      </div>
    </div>
    <p v-if="siteScope" data-site-insights-scope class="site-detail-scope mt-4">{{ t('siteDetail.siteScope') }}</p>
  </aside>
</template>

<script setup lang="ts">
import SiteTargetSelector from './SiteTargetSelector.vue'
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
defineProps<{ presentation: SiteTargetPresentation; pending: boolean; selected: string; siteScope: boolean }>()
const emit = defineEmits<{ select: [target: string] }>()
const { t } = useI18n()
</script>
