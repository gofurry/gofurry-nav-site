<template>
  <div class="rounded-xl bg-orange-50 p-5">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="font-semibold">{{ t('site.healthSummary.title') }}</h3>
        <p class="mt-1 text-xs text-gray-500">{{ t('site.healthSummary.observationOnly') }}</p>
      </div>
      <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(displayStatus)]">
        {{ statusText(displayStatus) }}
      </span>
    </div>

    <div class="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
      <div>
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.siteStatus') }}</h4>
        <div class="rounded-lg bg-orange-100 p-3">
          <div class="mb-2">
            <span class="font-bold">{{ t('site.healthSummary.state') }}:</span>
            {{ stateText(siteSummary?.state) }}
          </div>
          <div>
            <span class="font-bold">{{ t('site.healthSummary.targetCount') }}:</span>
            {{ siteSummary?.target_count ?? 0 }}
          </div>
          <div v-if="statusCountEntries.length" class="mt-2 flex flex-wrap gap-2">
            <span
              v-for="[status, count] in statusCountEntries"
              :key="status"
              class="rounded-full bg-orange-50 px-2 py-0.5 text-xs"
            >
              {{ statusText(status) }} {{ count }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="targetSummary">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.currentTarget') }}</h4>
        <div class="rounded-lg bg-orange-100 p-3">
          <div class="break-all font-mono text-xs">{{ targetSummary.target }}</div>
          <div class="mt-2">
            <span class="font-bold">{{ t('site.healthSummary.status') }}:</span>
            {{ statusText(targetSummary.status) }}
          </div>
          <div v-if="protocolEntries.length" class="mt-2 flex flex-wrap gap-2">
            <span
              v-for="[protocol, protocolSummary] in protocolEntries"
              :key="protocol"
              class="rounded-full bg-orange-50 px-2 py-0.5 text-xs"
            >
              {{ protocol.toUpperCase() }} {{ protocolSummary.stale ? t('site.healthSummary.stale') : protocolSummary.status }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="reasonMessages.length" class="mt-4 rounded-lg bg-orange-100 p-3 text-sm">
      <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.reasons') }}</h4>
      <ul class="space-y-1">
        <li v-for="reason in reasonMessages" :key="reason">{{ reason }}</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { HealthStatus, SiteHealthSummary, TargetHealthSummary } from '~/types/nav'

const props = defineProps<{
  siteSummary: SiteHealthSummary | null
  targetSummary: TargetHealthSummary | null
}>()

const { t } = useI18n()

const displayStatus = computed(() => props.targetSummary?.status || props.siteSummary?.status || 'unknown')
const statusCountEntries = computed(() => Object.entries(props.siteSummary?.status_counts ?? {}).filter(([, count]) => count > 0))
const protocolEntries = computed(() => Object.entries(props.targetSummary?.protocols ?? {}))
const reasonMessages = computed(() => [
  ...(props.targetSummary?.reason_messages ?? []),
  ...(props.siteSummary?.reason_messages ?? []),
])

function statusText(status: string) {
  return t(`site.healthSummary.statuses.${status || 'unknown'}`)
}

function stateText(state?: string) {
  return t(`site.healthSummary.states.${state || 'missing'}`)
}

function statusClass(status: HealthStatus | string) {
  switch (status) {
    case 'healthy':
      return 'bg-green-100 text-green-800'
    case 'warning':
      return 'bg-yellow-100 text-yellow-800'
    case 'degraded':
      return 'bg-orange-200 text-orange-900'
    case 'down':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}
</script>
