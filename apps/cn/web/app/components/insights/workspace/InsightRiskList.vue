<template>
  <p v-if="!items.length" class="insights-empty-state">{{ $t('insights.certificateIntelligence.noItems') }}</p>
  <ul v-else class="insight-risk-list">
    <li v-for="item in items" :key="item.site.id">
      <NuxtLink :to="localePath('/site/' + item.site.id)" class="insight-risk-row" data-workspace-entity>
        <InsightEntityMedia domain="site" :entity="item.site" />
        <div class="insight-risk-row__identity"><strong>{{ item.site.name || '#' + item.site.id }}</strong><span>{{ item.target }}</span></div>
        <div class="insight-risk-row__status">
          <strong>{{ itemLabel(item) }}</strong>
          <span v-if="mode === 'expiry'">{{ $t('insights.workspace.daysToExpiry', { count: item.days_to_expiry ?? '—' }) }}</span>
        </div>
        <div class="insight-risk-row__time">
          <span>{{ mode === 'expiry' ? $t('insights.certificateIntelligence.notAfter') : $t('insights.workspace.observedAt') }}</span>
          <time :datetime="(mode === 'expiry' ? item.not_after : item.observed_at) || undefined">{{ formatWorkspaceTimestamp(mode === 'expiry' ? item.not_after : item.observed_at, locale) }}</time>
        </div>
      </NuxtLink>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
import type { CertificateInsightItem } from '@/types/insights'
import { formatWorkspaceTimestamp } from '@/utils/insightWorkspace'
const props = defineProps<{ items: CertificateInsightItem[]; mode: 'expiry' | 'verification' }>()
const { locale, t } = useI18n()
const localePath = useLocalePath()
function itemLabel(item: CertificateInsightItem) {
  const value = props.mode === 'expiry' ? item.expiry_status : item.verification_issue
  return value ? t('insights.certificateIntelligence.values.' + value) : '—'
}
</script>
