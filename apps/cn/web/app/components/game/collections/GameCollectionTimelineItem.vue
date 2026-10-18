<template>
  <NuxtLink :to="localePath(`/games/${item.game_id}`)" class="game-collection-card game-collection-timeline-card relative z-10 flex h-full min-w-0 flex-col overflow-hidden">
    <div class="game-collection-timeline-card__media overflow-hidden">
      <SteamAssetImage :src="item.header_url" :alt="item.name" loading="lazy" class="game-collection-image" />
    </div>
    <div class="flex min-w-0 flex-1 flex-col gap-2 px-1 pb-1 pt-3">
      <div class="flex min-w-0 items-center justify-between gap-2">
        <p class="game-collection-time min-w-0 truncate" :title="timeLabel">{{ timeLabel }}</p>
        <div v-if="item.primary_tag || item.secondary_tag" class="flex min-w-0 max-w-[55%] shrink-0 gap-1">
          <span v-if="item.primary_tag" class="game-collection-timeline-tag game-collection-timeline-tag--primary min-w-0 truncate" :title="item.primary_tag.name">{{ item.primary_tag.name }}</span>
          <span v-if="item.secondary_tag" class="game-collection-timeline-tag min-w-0 truncate" :title="item.secondary_tag.name">{{ item.secondary_tag.name }}</span>
        </div>
      </div>
      <h3 class="game-collection-title line-clamp-2">{{ item.name }}</h3>
      <p class="game-collection-copy line-clamp-2">{{ item.summary }}</p>
      <div class="game-collection-timeline-metrics mt-auto flex min-w-0 items-center gap-3 overflow-hidden whitespace-nowrap">
        <span v-if="item.rating" class="flex min-w-0 items-center gap-1" role="img" :aria-label="ratingLabel" :title="ratingLabel">
          <PhStar aria-hidden="true" /><span class="truncate">{{ ratingAverage }} · {{ number(item.rating.count) }}</span>
        </span>
        <span v-if="item.online" class="flex min-w-0 items-center gap-1" role="img" :aria-label="onlineLabel" :title="onlineLabel">
          <PhPulse aria-hidden="true" /><span class="truncate">{{ number(item.online.count) }}</span>
        </span>
        <span v-if="item.community_count && item.community_count > 0" class="flex min-w-0 items-center gap-1" role="img" :aria-label="communityLabel" :title="communityLabel">
          <PhUsersThree aria-hidden="true" /><span class="truncate">{{ number(item.community_count) }}</span>
        </span>
      </div>
    </div>
  </NuxtLink>
</template>
<script setup lang="ts">
import { PhStar, PhPulse, PhUsersThree } from '@phosphor-icons/vue'
import type { GameCollectionTimelineItem } from '~/types/game'
import SteamAssetImage from '~/components/common/SteamAssetImage.vue'
import { collectionTime } from '~/utils/gameCollectionPresentation'
const props = defineProps<{ item: GameCollectionTimelineItem }>()
const { t, locale } = useI18n()
const localePath = useLocalePath()
const numberFormat = computed(() => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN'))
const number = (value: number) => numberFormat.value.format(value)
const timeLabel = computed(() => collectionTime(props.item, locale.value, t))
const ratingAverage = computed(() => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', { maximumFractionDigits: 1 }).format(props.item.rating?.average ?? 0))
const ratingLabel = computed(() => t('game.collections.metrics.rating', { average: ratingAverage.value, count: number(props.item.rating?.count ?? 0) }))
const onlineLabel = computed(() => {
  const count = number(props.item.online?.count ?? 0)
  const observedAt = new Date(props.item.online?.collected_at ?? '')
  if (!Number.isFinite(observedAt.getTime())) return t('game.collections.metrics.online', { count })
  // Explicit UTC keeps SSR and hydration labels identical across host time zones.
  const time = new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', {
    timeZone: 'UTC', year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', hourCycle: 'h23', timeZoneName: 'short',
  }).format(observedAt)
  return t('game.collections.metrics.onlineAt', { count, time })
})
const communityLabel = computed(() => t('game.collections.metrics.communities', { count: number(props.item.community_count ?? 0) }))
</script>
