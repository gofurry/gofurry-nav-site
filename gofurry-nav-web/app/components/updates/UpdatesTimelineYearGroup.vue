<template>
  <div class="updates-year-group">
    <button
      type="button"
      class="year-divider year-toggle"
      :class="{ 'is-expanded': expanded }"
      @click="$emit('toggle')"
    >
      <span class="year-divider-text">{{ group.year }}</span>
      <span class="year-divider-meta">{{ yearSummary }}</span>
      <span class="year-divider-chevron" aria-hidden="true" />
    </button>

    <ol v-if="expanded" class="year-entries">
      <li
        v-for="item in visibleItems"
        :key="item.id"
        class="timeline-entry-wrap"
      >
        <UpdatesTimelineEntry
          :item="item"
          :latest="item.id === latestId"
          :latest-tag="latestTag"
          :locale-code="localeCode"
          :unavailable-label="unavailableLabel"
        />
      </li>

      <li v-if="hasMore" class="year-load-more-wrap">
        <button type="button" class="year-load-more" @click="$emit('loadMore')">
          {{ loadMoreLabel }}
        </button>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
import type { NavUpdateNotice } from '~/types/nav'

defineEmits<{
  toggle: []
  loadMore: []
}>()

defineProps<{
  group: {
    year: string
    items: NavUpdateNotice[]
  }
  expanded: boolean
  visibleItems: NavUpdateNotice[]
  hasMore: boolean
  latestId: number | null
  latestTag: string
  loadMoreLabel: string
  yearSummary: string
  localeCode: string
  unavailableLabel: string
}>()
</script>
