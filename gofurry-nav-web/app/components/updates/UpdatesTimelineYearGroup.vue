<template>
  <div class="updates-year-group" :class="{ 'is-dark-theme': dark }">
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
          :dark="dark"
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
  dark: boolean
}>()
</script>

<style scoped>
.updates-year-group {
  margin-bottom: 10px;
}

.year-divider {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  gap: 14px;
  margin: 8px 0 14px;
  padding-left: 28px;
}

.year-divider::before {
  content: "";
  position: absolute;
  top: 50%;
  left: -34px;
  width: 18px;
  height: 1px;
  background: rgba(15, 118, 110, 0.5);
}

.year-divider-text {
  color: rgba(32, 24, 21, 0.42);
  font-size: 0.76rem;
  font-weight: 700;
  text-transform: uppercase;
}

.year-divider-meta {
  color: rgba(32, 24, 21, 0.36);
  font-size: 0.78rem;
}

.year-toggle {
  border: 0;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.year-divider-chevron {
  position: relative;
  flex: 0 0 auto;
  width: 10px;
  height: 10px;
  margin-left: auto;
}

.year-divider-chevron::before,
.year-divider-chevron::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 6px;
  height: 1px;
  background: rgba(32, 24, 21, 0.42);
  transition: transform 180ms ease;
}

.year-divider-chevron::before {
  left: 0;
  transform: translateY(-50%) rotate(45deg);
}

.year-divider-chevron::after {
  right: 0;
  transform: translateY(-50%) rotate(-45deg);
}

.year-toggle.is-expanded .year-divider-chevron::before {
  transform: translateY(-50%) rotate(-45deg);
}

.year-toggle.is-expanded .year-divider-chevron::after {
  transform: translateY(-50%) rotate(45deg);
}

.year-entries {
  margin: 0;
  padding: 0;
  list-style: none;
}

.year-load-more-wrap {
  padding: 18px 0 8px 28px;
}

.year-load-more {
  border: 1px solid rgba(15, 118, 110, 0.18);
  background: rgba(255, 251, 247, 0.5);
  padding: 0.58rem 0.9rem;
  color: #0f766e;
  font-size: 0.82rem;
  font-weight: 700;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease;
}

.year-load-more:hover,
.year-load-more:focus-visible {
  border-color: rgba(15, 118, 110, 0.34);
  background: rgba(255, 251, 247, 0.82);
  transform: translateY(-1px);
}

.updates-year-group.is-dark-theme .year-divider-text,
.updates-year-group.is-dark-theme .year-divider-meta {
  color: rgba(174, 205, 212, 0.68);
}

.updates-year-group.is-dark-theme .year-divider-text {
  color: rgba(210, 231, 236, 0.8);
}

.updates-year-group.is-dark-theme .year-divider-meta {
  color: rgba(166, 199, 206, 0.64);
}

.updates-year-group.is-dark-theme .year-divider::before {
  background: rgba(127, 240, 247, 0.56);
}

.updates-year-group.is-dark-theme .year-divider-chevron::before,
.updates-year-group.is-dark-theme .year-divider-chevron::after {
  background: rgba(224, 248, 252, 0.86);
}

.updates-year-group.is-dark-theme .year-load-more {
  border-color: rgba(154, 248, 251, 0.28);
  background: rgba(7, 24, 34, 0.74);
  color: rgba(223, 239, 243, 0.82);
  box-shadow: inset 0 0 0 1px rgba(247, 254, 255, 0.04);
}

.updates-year-group.is-dark-theme .year-load-more:hover,
.updates-year-group.is-dark-theme .year-load-more:focus-visible {
  border-color: rgba(154, 248, 251, 0.44);
  background: rgba(10, 31, 44, 0.9);
  color: rgba(232, 245, 248, 0.9);
}

@media (max-width: 720px) {
  .year-divider {
    padding-left: 18px;
  }

  .year-divider::before {
    left: -22px;
    width: 14px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .year-divider-chevron::before,
  .year-divider-chevron::after,
  .year-load-more {
    transition: none;
  }
}
</style>
