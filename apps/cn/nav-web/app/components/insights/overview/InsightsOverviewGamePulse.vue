<template>
  <section class="overview-games" aria-labelledby="overview-games-title" data-overview-games>
    <div class="overview-section-heading">
      <div><p class="overview-kicker">{{ $t('insights.editorial.gamesKicker') }}</p><h2 id="overview-games-title">{{ $t('insights.editorial.gamesTitle') }}</h2></div>
      <NuxtLink :to="localePath('/insights/games')" class="overview-text-link">{{ $t('insights.editorial.explore') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
    <p v-if="!panel" class="overview-unavailable">{{ $t('insights.emptyStates.unavailable') }}</p>
    <div v-else class="overview-pulse grid gap-4" data-overview-pulse>
      <template v-for="slot in slots" :key="slot.role">
        <NuxtLink v-if="slot.game" :to="localePath(`/games/${encodeURIComponent(slot.game.id)}`)" class="overview-pulse__item flex min-w-0 flex-col items-start gap-2" :data-pulse="slot.role">
          <span class="overview-kicker">{{ $t(slot.role === 'players' ? 'insights.overviewHero.playerObservation' : `insights.editorial.pulse.${slot.role}`) }}</span>
          <strong>{{ slot.game.name }}</strong>
          <template v-if="slot.role === 'players' && observation">
            <span>{{ number(observation.count) }} {{ $t('insights.editorial.playersUnit') }}</span>
            <time :datetime="observation.collectedAt">{{ observation.label }}</time>
            <span v-if="validPeak(slot.game.online_count.peak_count)" class="overview-pulse__detail">{{ $t('game.panel.onlinePeak') }} {{ number(slot.game.online_count.peak_count!) }}</span>
          </template>
          <span v-else-if="slot.role === 'discount' && discount" class="overview-pulse__detail">{{ formatMinorAmount(discount.final_amount, discount.currency, locale) }} · {{ $t('insights.editorial.discount', { percent: discount.discount_percent }) }} · {{ $t('insights.editorial.usRegion') }} / {{ discount.currency }}</span>
          <span v-else-if="slot.role === 'latest' && slot.game.release_date" class="overview-pulse__detail">{{ slot.game.release_date }}</span>
        </NuxtLink>
        <div v-else :data-pulse="slot.role"><p class="overview-kicker">{{ $t(slot.role === 'players' ? 'insights.overviewHero.playerObservation' : `insights.editorial.pulse.${slot.role}`) }}</p><p class="overview-unavailable">{{ $t('insights.emptyStates.unavailable') }}</p></div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GameV2PanelRecord } from '@/types/game'
import { pulseDiscountPrice } from '@/utils/insightGamePulse'
import { overviewPlayerObservation, selectOverviewPulse } from '@/utils/insightOverviewPresentation'
import { formatMinorAmount } from '@/utils/insightPrices'

const props = defineProps<{ panel: GameV2PanelRecord | null, evaluatedAt: number }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const pulse = computed(() => selectOverviewPulse(props.panel, props.evaluatedAt))
const slots = computed(() => (['players', 'discount', 'latest'] as const).map(role => ({ role, game: pulse.value[role] })))
const observation = computed(() => pulse.value.players ? overviewPlayerObservation(pulse.value.players, props.evaluatedAt) : null)
const discount = computed(() => pulse.value.discount ? pulseDiscountPrice(pulse.value.discount) : null)
const number = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
const validPeak = (value: number | undefined) => typeof value === 'number' && Number.isFinite(value) && value >= 0
</script>
