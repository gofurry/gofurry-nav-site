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
