<template>
  <header class="insights-domain-header" data-domain-header>
    <h1>{{ $t(domain === 'site' ? 'insights.sites.title' : 'insights.games.title') }}</h1>
    <p class="insights-domain-header__intro">{{ $t(`insights.domain.${domain}.description`) }}</p>
    <div class="insights-domain-header__facts">
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
