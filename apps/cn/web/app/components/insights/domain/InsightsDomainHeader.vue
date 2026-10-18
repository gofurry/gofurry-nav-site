<template>
  <header class="insights-domain-header -mt-2 mb-5 min-[1101px]:-mt-4 min-[1101px]:mb-6" data-domain-header>
    <h1 class="sr-only">{{ $t(domain === 'site' ? 'insights.sites.title' : 'insights.games.title') }}</h1>
    <div class="insights-domain-header__facts flex flex-col items-start gap-2 min-[761px]:flex-row min-[761px]:items-baseline min-[761px]:justify-between min-[761px]:gap-4">
      <p><strong data-domain-count>{{ entityCount === null ? '—' : new Intl.NumberFormat(locale).format(entityCount) }}</strong> {{ $t(`insights.domain.${domain}.count`) }}</p>
      <p class="insights-domain-header__snapshot">{{ $t('insights.editorial.snapshot') }} <time v-if="generatedAt" :datetime="generatedAt">{{ formatOverviewSnapshot(generatedAt, locale) }}</time><span v-else>—</span></p>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InsightDomain } from '@/types/insights'
import { formatOverviewSnapshot } from '@/utils/insightOverview'
defineProps<{ domain: InsightDomain, entityCount: number | null, generatedAt: string | null }>()
const { locale } = useI18n()
</script>
