<template>
  <section class="min-w-0" :aria-labelledby="titleId">
    <div class="flex min-w-0 items-center justify-between gap-3">
      <h3 :id="titleId" class="site-overview-title">{{ title }}</h3>
      <div v-if="state === 'ready'" data-site-change-modes class="hidden gap-1 md:flex" role="group" :aria-label="t('insights.entity.timelineView')">
        <button v-for="option in modes" :key="option" type="button" :data-site-change-mode="option" :aria-pressed="mode === option"
          class="site-detail-segment" @click="mode = option">{{ t('insights.entity.timelineModes.' + option) }}</button>
      </div>
    </div>
    <p v-if="state !== 'ready'" class="site-detail-note mt-3">{{ emptyText }}</p>
    <ol v-else :data-site-overview-change-grid="variant === 'overview' ? '' : undefined" data-site-change-stream :data-mode="mode"
      class="mt-4 grid min-w-0" :class="mode === 'compact' ? 'gap-1.5 md:grid-cols-3 md:gap-x-8 md:gap-y-7' : 'gap-1.5'">
      <li v-for="(item, index) in items" :key="item.key" :data-site-change="variant === 'overview' ? '' : undefined"
        :data-site-insight-change="variant === 'insights' ? '' : undefined" :data-site-overview-change-card="variant === 'overview' ? '' : undefined"
        :data-site-change-type="item.type" :data-site-change-category="item.category" :data-connector="connector(index, items.length)"
        :style="compactPlacement(index)" class="site-detail-change-item relative min-w-0"
        :class="mode === 'compact' ? 'md:col-[var(--timeline-column)] md:row-[var(--timeline-row)]' : 'flex flex-wrap items-baseline justify-between gap-x-5 gap-y-2'">
        <div class="min-w-0"><SiteChangeCategory :category="item.category" :label="item.categoryLabel" /><p class="site-overview-change-event mt-1 break-words">{{ item.label }}</p></div>
        <time :datetime="item.dateTime" :data-precise="item.precise" class="site-detail-note mt-2 block shrink-0">{{ item.when }}</time>
        <SiteSequenceArrow v-if="mode === 'compact'" :direction="connector(index, items.length)" class="hidden md:block" />
      </li>
    </ol>
  </section>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { compactPlacement, connector } from '~/utils/serpentineSequence'
import SiteChangeCategory from './SiteChangeCategory.vue'
import SiteSequenceArrow from './SiteSequenceArrow.vue'
defineProps<{ title: string; titleId: string; variant: 'overview' | 'insights'; state: string; emptyText: string;
  items: Array<{ key: string; type?: string; category: string; categoryLabel: string; label: string; when: string; dateTime: string; precise: boolean }> }>()
const { t } = useI18n(), mode = ref<'compact' | 'list'>('compact'), modes = ['compact', 'list'] as const
</script>
