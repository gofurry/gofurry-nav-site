<template>
  <section class="insights-section insights-recent-changes" aria-labelledby="insights-changes-title">
    <div class="insights-section__heading">
      <div>
        <p class="insights-eyebrow">{{ $t('insights.overview.recentDescription') }}</p>
        <h2 id="insights-changes-title">{{ $t('insights.changes.title') }}</h2>
      </div>
    </div>

    <p v-if="unavailable" class="insights-empty-state">
      {{ $t('insights.emptyStates.unavailable') }}
    </p>
    <p v-else-if="items.length === 0" class="insights-empty-state">
      {{ $t('insights.emptyStates.changesEmpty') }}
    </p>
    <div v-else class="insights-change-list">
      <NuxtLink
        v-for="item in items"
        :key="`${item.domain}:${item.entity.id}:${item.type}:${item.occurred_at || item.date}`"
        :to="localePath(entityPath(item))"
        class="insights-change"
        data-change-link
      >
        <span class="insights-change__domain">{{ $t(`insights.changes.${item.domain}`) }}</span>
        <span class="insights-change__body">
          <strong>{{ item.entity.name }}</strong>
          <span>{{ eventLabel(item.type) }}</span>
        </span>
        <time :datetime="item.occurred_at || item.date">{{ formatWhen(item) }}</time>
      </NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InsightFeedItem } from '@/types/insights'
import { siteDetailPath } from '@/utils/siteRoutes'

defineProps<{
  items: InsightFeedItem[]
  unavailable?: boolean
}>()

const localePath = useLocalePath()
const { locale, t } = useI18n()

const knownEvents: Record<string, string> = {
  'site.ipv6.enabled': 'siteIpv6Enabled',
  'site.ipv6.disabled': 'siteIpv6Disabled',
  'site.tls13.enabled': 'siteTls13Enabled',
  'site.tls13.disabled': 'siteTls13Disabled',
  'site.security_txt.added': 'siteSecurityTxtAdded',
  'site.security_txt.removed': 'siteSecurityTxtRemoved',
  'site.primary_target.changed': 'sitePrimaryTargetChanged',
  'site.tls_certificate.changed': 'siteTlsCertificateChanged',
  'game.free.enabled': 'gameFreeEnabled',
  'game.free.disabled': 'gameFreeDisabled',
  'game.windows.added': 'gameWindowsAdded',
  'game.windows.removed': 'gameWindowsRemoved',
  'game.linux.added': 'gameLinuxAdded',
  'game.linux.removed': 'gameLinuxRemoved',
  'game.release.available': 'gameReleaseAvailable',
  'game.release.withdrawn': 'gameReleaseWithdrawn',
  'game.release.changed': 'gameReleaseChanged',
  'game.price.increased': 'gamePriceIncreased',
  'game.price.decreased': 'gamePriceDecreased',
  'game.price.state_changed': 'gamePriceStateChanged',
  'game.price.currency_changed': 'gamePriceCurrencyChanged',
  'game.discount.started': 'gameDiscountStarted',
  'game.discount.ended': 'gameDiscountEnded',
  'game.discount.changed': 'gameDiscountChanged',
}

function eventLabel(type: string) {
  const key = knownEvents[type]
  return key ? t(`insights.changes.events.${key}`) : t('insights.changes.events.unknown')
}

function entityPath(item: InsightFeedItem) {
  return item.domain === 'site' ? siteDetailPath(item.entity.id) : `/games/${encodeURIComponent(String(item.entity.id))}`
}

function formatWhen(item: InsightFeedItem) {
  if (!item.occurred_at) return item.date
  const value = new Date(item.occurred_at)
  if (Number.isNaN(value.getTime())) return item.date
  return new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(value)
}
</script>
