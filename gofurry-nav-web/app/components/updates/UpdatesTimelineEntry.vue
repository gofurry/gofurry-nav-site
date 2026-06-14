<template>
  <article
    class="updates-entry"
    :class="{
      'is-latest': latest,
    }"
    tabindex="0"
  >
    <div class="entry-marker" aria-hidden="true" />

    <time class="entry-stamp" :datetime="item.published_at">
      <span class="entry-month">{{ monthDayLabel }}</span>
      <span class="entry-time">{{ clockLabel }}</span>
    </time>

    <div class="entry-copy">
      <div class="entry-heading">
        <h2>{{ item.title }}</h2>
        <span v-if="latest" class="entry-tag">{{ latestTag }}</span>
      </div>
      <p class="entry-body">{{ item.body }}</p>
      <p class="entry-meta">{{ fullDateLabel }}</p>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { NavUpdateNotice } from '~/types/nav'
import {
  formatUpdatesClock,
  formatUpdatesFullDate,
  formatUpdatesMonthDay,
} from '~/utils/updatesDate'

const props = defineProps<{
  item: NavUpdateNotice
  latest: boolean
  latestTag: string
  localeCode: string
  unavailableLabel: string
}>()

const monthDayLabel = computed(() => formatUpdatesMonthDay(props.item.published_at, props.localeCode))
const clockLabel = computed(() => formatUpdatesClock(props.item.published_at, props.localeCode))
const fullDateLabel = computed(() => (
  formatUpdatesFullDate(props.item.published_at, props.localeCode, props.unavailableLabel)
))
</script>

<style scoped>
.updates-entry {
  position: relative;
  display: grid;
  grid-template-columns: 126px minmax(0, 1fr);
  gap: clamp(18px, 4vw, 34px);
  padding: 18px 0 28px;
  border-bottom: 1px solid var(--updates-entry-border);
  background: var(--updates-entry-bg);
  outline: none;
}

.entry-marker {
  position: absolute;
  top: 26px;
  left: -29px;
  width: 11px;
  height: 11px;
  border: 2px solid var(--updates-entry-marker-border);
  background: var(--updates-entry-marker-bg);
  transform: rotate(45deg);
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.updates-entry:hover .entry-marker,
.updates-entry:focus-visible .entry-marker {
  background: var(--updates-entry-marker-active-bg);
  box-shadow: 0 0 0 9px var(--updates-entry-marker-ring);
  transform: rotate(45deg) scale(1.08);
}

.updates-entry.is-latest .entry-marker {
  background: var(--updates-entry-marker-active-bg);
  animation: marker-pulse 2400ms ease-in-out infinite;
}

.entry-stamp {
  display: grid;
  align-content: start;
  gap: 6px;
  padding-top: 4px;
  padding-left: 14px;
  padding-right: 14px;
  color: var(--updates-entry-stamp);
  font-variant-numeric: tabular-nums;
}

.entry-month {
  font-size: clamp(1.3rem, 3vw, 1.9rem);
  font-weight: 820;
  line-height: 1;
}

.entry-time {
  font-size: 0.82rem;
}

.entry-copy {
  min-width: 0;
  padding-left: 18px;
  padding-right: min(8vw, 48px);
  transition: transform 180ms ease;
}

.updates-entry:hover .entry-copy,
.updates-entry:focus-visible .entry-copy {
  transform: translateX(4px);
}

.entry-heading {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 16px;
}

.entry-heading h2 {
  margin: 0;
  color: var(--updates-entry-heading);
  font-size: clamp(1.38rem, 2.4vw, 2.2rem);
  font-weight: 830;
  line-height: 1.2;
  text-shadow: var(--updates-entry-heading-shadow);
  overflow-wrap: anywhere;
}

.updates-entry.is-latest .entry-heading h2 {
  font-size: clamp(1.6rem, 2.8vw, 2.5rem);
}

.entry-tag {
  flex: 0 0 auto;
  border: 1px solid var(--updates-entry-tag-border);
  background: var(--updates-entry-tag-bg);
  padding: 0.28rem 0.54rem;
  color: var(--updates-entry-tag-text);
  font-size: 0.72rem;
  font-weight: 900;
  box-shadow: var(--updates-entry-tag-shadow);
  text-transform: uppercase;
}

.entry-body {
  max-width: 820px;
  margin: 18px 0 0;
  color: var(--updates-entry-body);
  font-size: 1rem;
  line-height: 1.9;
  white-space: pre-line;
  overflow-wrap: anywhere;
}

.entry-meta {
  margin: 18px 0 0;
  color: var(--updates-entry-meta);
  font-size: 0.84rem;
  font-variant-numeric: tabular-nums;
}

.updates-entry:hover,
.updates-entry:focus-visible {
  background: var(--updates-entry-bg-hover);
}

.entry-month {
  color: var(--updates-entry-month);
}

.entry-time {
  color: var(--updates-entry-time);
}

@keyframes marker-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 var(--updates-entry-marker-pulse);
  }
  50% {
    box-shadow: 0 0 0 12px var(--updates-entry-marker-pulse-clear);
  }
}

@media (max-width: 720px) {
  .updates-entry {
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
  }

  .entry-marker {
    top: 18px;
    left: -25px;
  }

  .entry-stamp {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding-left: 8px;
    padding-right: 0;
  }

  .entry-copy {
    padding-left: 8px;
    padding-right: 0;
  }

  .entry-heading {
    flex-wrap: wrap;
  }

  .updates-entry:hover .entry-copy,
  .updates-entry:focus-visible .entry-copy {
    transform: translateY(-2px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .updates-entry.is-latest .entry-marker {
    animation: none;
  }

  .entry-copy,
  .entry-marker {
    transition: none;
  }
}
</style>
