<template>
  <div class="updates-year-group">
    <button
      type="button"
      class="updates-year-group__divider updates-year-group__toggle"
      :class="{ 'is-expanded': expanded }"
      @click="$emit('toggle')"
    >
      <span class="updates-year-group__label">{{ group.year }}</span>
      <span class="updates-year-group__meta">{{ yearSummary }}</span>
      <span class="updates-year-group__chevron" aria-hidden="true" />
    </button>

    <ol v-if="expanded" class="updates-year-group__entries">
      <li
        v-for="item in visibleItems"
        :key="item.id"
      >
        <UpdatesTimelineEntry
          :item="item"
          :latest="item.id === latestId"
          :latest-tag="latestTag"
          :locale-code="localeCode"
          :unavailable-label="unavailableLabel"
        />
      </li>

      <li v-if="hasMore" class="updates-year-group__load-more-wrap">
        <button type="button" class="updates-year-group__load-more" @click="$emit('loadMore')">
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
