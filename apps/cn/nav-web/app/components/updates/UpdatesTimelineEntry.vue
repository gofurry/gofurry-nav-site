<template>
  <article
    class="updates-entry"
    :class="{
      'is-latest': latest,
    }"
    tabindex="0"
  >
    <div class="updates-entry__marker" aria-hidden="true" />

    <time class="updates-entry__stamp" :datetime="item.published_at">
      <span class="updates-entry__month">{{ monthDayLabel }}</span>
      <span class="updates-entry__time">{{ clockLabel }}</span>
    </time>

    <div class="updates-entry__copy">
      <div class="updates-entry__heading">
        <h2>{{ item.title }}</h2>
        <span v-if="latest" class="updates-entry__tag">{{ latestTag }}</span>
      </div>
      <p class="updates-entry__body">{{ item.body }}</p>
      <p class="updates-entry__meta">{{ fullDateLabel }}</p>
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
