<template>
  <p v-if="!panel" class="insights-game-pulse__unavailable">{{ $t('insights.emptyStates.unavailable') }}</p>
  <div v-else class="insights-game-pulse grid items-start gap-5 min-[761px]:grid-cols-[minmax(0,7fr)_minmax(0,5fr)] min-[761px]:gap-6 min-[1101px]:gap-x-10">
    <NuxtLink v-if="pulse.players" :to="localePath(`/games/${encodeURIComponent(pulse.players.id)}`)" class="insights-game-pulse__item insights-game-pulse__item--lead flex min-w-0 flex-col gap-4" data-pulse="players">
      <InsightEntityMedia domain="game" :entity="entity(pulse.players)" :data-pulse-source="source(pulse.players)" />
      <span class="insights-game-pulse__body relative flex min-w-0 flex-col gap-[7px]">
        <span class="insights-game-pulse__kicker">{{ $t('insights.editorial.pulse.players') }}</span>
        <strong>{{ pulse.players.name }}</strong>
        <span class="insights-game-pulse__count">{{ number(pulse.players.online_count.count) }} <small>{{ $t('insights.editorial.playersUnit') }}</small></span>
        <span class="insights-game-pulse__detail">{{ $t('insights.editorial.latestObservation') }}</span>
        <span v-if="typeof pulse.players.online_count.peak_count === 'number' && Number.isFinite(pulse.players.online_count.peak_count)" class="insights-game-pulse__detail">{{ $t('game.panel.onlinePeak') }} {{ number(pulse.players.online_count.peak_count) }}</span>
      </span>
    </NuxtLink>
    <div v-else class="insights-game-pulse__empty" data-pulse="players"><span class="insights-game-pulse__kicker">{{ $t('insights.editorial.pulse.players') }}</span><p>{{ $t('insights.emptyStates.unavailable') }}</p></div>
    <div class="insights-game-pulse__secondary flex min-w-0 flex-col gap-6">
      <template v-for="slot in secondarySlots" :key="slot.role">
        <NuxtLink v-if="slot.game" :to="localePath(`/games/${encodeURIComponent(slot.game.id)}`)" class="insights-game-pulse__item flex min-w-0 items-start gap-3.5 min-[1101px]:gap-[18px]" :data-pulse="slot.role">
          <InsightEntityMedia domain="game" :entity="entity(slot.game)" :data-pulse-source="source(slot.game)" />
          <span class="insights-game-pulse__body flex min-w-0 flex-1 flex-col gap-[7px]">
            <span class="insights-game-pulse__kicker">{{ $t(`insights.editorial.pulse.${slot.role}`) }}</span>
            <strong>{{ slot.game.name }}</strong>
            <span v-if="slot.role === 'discount' && discount" class="insights-game-pulse__detail">
              <b>{{ formatMinorAmount(discount.final_amount, discount.currency, locale) }}</b>
              · {{ $t('insights.editorial.discount', { percent: discount.discount_percent }) }} · {{ $t('insights.editorial.usRegion') }} / {{ discount.currency }}
            </span>
            <span v-else-if="slot.role === 'latest' && slot.game.release_date" class="insights-game-pulse__detail">{{ slot.game.release_date }}</span>
          </span>
        </NuxtLink>
        <div v-else class="insights-game-pulse__empty" :data-pulse="slot.role"><span class="insights-game-pulse__kicker">{{ $t(`insights.editorial.pulse.${slot.role}`) }}</span><p>{{ $t('insights.emptyStates.unavailable') }}</p></div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GameV2ListItem, GameV2PanelRecord } from '@/types/game'
import type { InsightEntityRef } from '@/types/insights'
import { pulseDiscountPrice, selectGamePulse } from '@/utils/insightGamePulse'
import { formatMinorAmount } from '@/utils/insightPrices'
import InsightEntityMedia from '../entity/InsightEntityMedia.vue'

const props = defineProps<{ panel: GameV2PanelRecord | null }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const pulse = computed(() => selectGamePulse(props.panel))
const secondarySlots = computed(() => (['discount', 'latest'] as const).map(role => ({ role, game: pulse.value[role] })))
const discount = computed(() => pulse.value.discount ? pulseDiscountPrice(pulse.value.discount) : null)
const number = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
// Layout hint only; selection and image failure still belong to their existing owners.
const source = (game: GameV2ListItem) => game.header_url ? 'header' : game.capsule_url ? 'capsule' : 'none'
function entity(game: GameV2ListItem): InsightEntityRef {
  return { id: Number(game.id), name: game.name, visual: { kind: 'game_header', asset: game.header_url || game.capsule_url || null } }
}
</script>
