<template>
  <section class="intelligence-panel">
    <h2>{{ title }}</h2>
    <p v-if="items.length === 0" class="insights-empty-state">{{ $t('insights.certificateIntelligence.noItems') }}</p>
    <div v-else class="intelligence-table-wrap">
      <table class="intelligence-table">
        <thead>
          <tr>
            <th>{{ $t('insights.certificateIntelligence.site') }}</th>
            <th>{{ $t('insights.certificateIntelligence.target') }}</th>
            <th>{{ mode === 'expiry' ? $t('insights.certificateIntelligence.expiry') : $t('insights.certificateIntelligence.issue') }}</th>
            <th>{{ $t('insights.certificateIntelligence.notAfter') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="`${mode}-${item.site.id}`">
            <td><NuxtLink :to="localePath(`/site/${item.site.id}`)">{{ item.site.name || `#${item.site.id}` }}</NuxtLink></td>
            <td>{{ item.target }}</td>
            <td>{{ itemLabel(item) }}</td>
            <td>{{ formatTimestamp(item.not_after) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CertificateInsightItem } from '@/types/insights'

const props = defineProps<{ title: string; items: CertificateInsightItem[]; mode: 'expiry' | 'verification' }>()
const { locale, t } = useI18n()
const localePath = useLocalePath()

function itemLabel(item: CertificateInsightItem) {
  const value = props.mode === 'expiry' ? item.expiry_status : item.verification_issue
  return value ? t(`insights.certificateIntelligence.values.${value}`) : '—'
}

function formatTimestamp(value: string | null) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', {
    dateStyle: 'medium', timeStyle: 'short', timeZone: 'UTC',
  }).format(new Date(value))
}
</script>
