<template>
  <article
    class="updates-entry"
    :class="{
      'is-latest': latest,
      'is-dark-theme': dark,
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
  dark: boolean
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
  border-bottom: 1px solid rgba(62, 50, 41, 0.12);
  outline: none;
}

.entry-marker {
  position: absolute;
  top: 26px;
  left: -29px;
  width: 11px;
  height: 11px;
  border: 2px solid #0f766e;
  background: rgba(255, 251, 247, 0.92);
  transform: rotate(45deg);
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.updates-entry:hover .entry-marker,
.updates-entry:focus-visible .entry-marker {
  background: #0f766e;
  box-shadow: 0 0 0 9px rgba(15, 118, 110, 0.12);
  transform: rotate(45deg) scale(1.08);
}

.updates-entry.is-latest .entry-marker {
  background: #0f766e;
  animation: marker-pulse 2400ms ease-in-out infinite;
}

.entry-stamp {
  display: grid;
  align-content: start;
  gap: 6px;
  padding-top: 4px;
  padding-left: 14px;
  padding-right: 14px;
  color: rgba(32, 24, 21, 0.54);
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
  color: rgba(31, 23, 19, 0.82);
  font-size: clamp(1.38rem, 2.4vw, 2.2rem);
  font-weight: 830;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.updates-entry.is-latest .entry-heading h2 {
  font-size: clamp(1.6rem, 2.8vw, 2.5rem);
}

.entry-tag {
  flex: 0 0 auto;
  border: 1px solid rgba(15, 118, 110, 0.22);
  background: rgba(15, 118, 110, 0.08);
  padding: 0.28rem 0.54rem;
  color: #0f766e;
  font-size: 0.72rem;
  font-weight: 900;
  text-transform: uppercase;
}

.entry-body {
  max-width: 820px;
  margin: 18px 0 0;
  color: rgba(32, 24, 21, 0.66);
  font-size: 1rem;
  line-height: 1.9;
  white-space: pre-line;
  overflow-wrap: anywhere;
}

.entry-meta {
  margin: 18px 0 0;
  color: rgba(32, 24, 21, 0.38);
  font-size: 0.84rem;
  font-variant-numeric: tabular-nums;
}

.updates-entry.is-dark-theme {
  border-color: rgba(151, 224, 236, 0.18);
  background: linear-gradient(90deg, rgba(6, 20, 29, 0.18), rgba(6, 20, 29, 0.04));
}

.updates-entry.is-dark-theme:hover,
.updates-entry.is-dark-theme:focus-visible {
  background: linear-gradient(90deg, rgba(9, 28, 40, 0.3), rgba(9, 28, 40, 0.1));
}

.updates-entry.is-dark-theme .entry-heading h2 {
  color: rgba(226, 238, 242, 0.84);
  text-shadow: 0 1px 18px rgba(0, 0, 0, 0.22);
}

.updates-entry.is-dark-theme .entry-stamp,
.updates-entry.is-dark-theme .entry-meta {
  color: rgba(174, 205, 212, 0.68);
}

.updates-entry.is-dark-theme .entry-month {
  color: rgba(215, 233, 238, 0.82);
}

.updates-entry.is-dark-theme .entry-time {
  color: rgba(182, 214, 221, 0.72);
}

.updates-entry.is-dark-theme .entry-body {
  color: rgba(204, 223, 228, 0.76);
}

.updates-entry.is-dark-theme .entry-marker {
  border-color: rgba(153, 245, 250, 0.96);
  background: rgba(10, 24, 34, 0.96);
}

.updates-entry.is-dark-theme:hover .entry-marker,
.updates-entry.is-dark-theme:focus-visible .entry-marker,
.updates-entry.is-dark-theme.is-latest .entry-marker {
  background: #9af8fb;
}

.updates-entry.is-dark-theme:hover .entry-marker,
.updates-entry.is-dark-theme:focus-visible .entry-marker {
  box-shadow: 0 0 0 9px rgba(154, 248, 251, 0.14);
}

.updates-entry.is-dark-theme .entry-tag {
  color: rgba(225, 242, 246, 0.84);
  border-color: rgba(154, 248, 251, 0.3);
  background: rgba(154, 248, 251, 0.12);
  box-shadow: inset 0 0 0 1px rgba(247, 254, 255, 0.04);
}

@keyframes marker-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(15, 118, 110, 0.22);
  }
  50% {
    box-shadow: 0 0 0 12px rgba(15, 118, 110, 0);
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
