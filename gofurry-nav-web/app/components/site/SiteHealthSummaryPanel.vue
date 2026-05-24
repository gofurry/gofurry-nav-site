<template>
  <div class="rounded-xl bg-orange-50 p-5">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="font-semibold">{{ t('site.healthSummary.title') }}</h3>
      </div>
      <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(displayStatus)]">
        {{ statusText(displayStatus) }}
      </span>
    </div>

    <div class="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
      <div class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.siteStatus') }}</h4>
        <div class="rounded-lg bg-orange-100 p-3 md:flex-1">
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
          <div v-if="targetEntries.length" class="mt-4 space-y-2">
            <div
              v-for="target in targetEntries"
              :key="target.target"
              class="rounded-md bg-orange-50 p-2"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="break-all font-mono text-xs">{{ target.target }}</span>
                <span :class="['rounded-full px-2 py-0.5 text-xs font-semibold', statusClass(target.status)]">
                  {{ statusText(target.status) }}
                </span>
              </div>
              <ul v-if="target.reason_messages?.length || target.reason_codes?.length" class="mt-2 space-y-1 text-xs text-gray-600">
                <li v-for="reason in targetReasons(target)" :key="`${target.target}:${reason}`">{{ reason }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div v-if="targetSummary" class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.currentTarget') }}</h4>
        <div class="rounded-lg bg-orange-100 p-3 md:flex-1">
          <div class="break-all font-mono text-xs">{{ targetSummary.target }}</div>
          <div class="mt-2">
            <span class="font-bold">{{ t('site.healthSummary.status') }}:</span>
            {{ statusText(targetSummary.status) }}
          </div>
          <div v-if="protocolEntries.length" class="mt-2 flex flex-wrap gap-2">
            <div
              v-for="[protocol, protocolSummary] in protocolEntries"
              :key="protocol"
              class="rounded-md bg-orange-50 p-2 text-xs"
            >
              <div class="mb-1 font-semibold">
                {{ protocol.toUpperCase() }} {{ protocolSummary.stale ? t('site.healthSummary.stale') : statusText(protocolSummary.status) }}
              </div>
              <div>{{ t('site.healthSummary.duration') }}: {{ protocolSummary.duration_ms }}ms</div>
              <div>{{ t('site.healthSummary.observedAt') }}: {{ formatTime(protocolSummary.observed_at) }}</div>
              <div>{{ t('site.healthSummary.staleAfter') }}: {{ protocolSummary.stale_after_seconds }}s</div>
              <div v-if="protocolSummary.error_code">
                {{ t('site.healthSummary.errorCode') }}: {{ protocolSummary.error_code }}
              </div>
            </div>
          </div>
        </div>
      </div>
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
const targetEntries = computed(() => props.siteSummary?.targets ?? [])
const protocolEntries = computed(() => Object.entries(props.targetSummary?.protocols ?? {}))

function statusText(status: string) {
  return t(`site.healthSummary.statuses.${status || 'unknown'}`)
}

function stateText(state?: string) {
  return t(`site.healthSummary.states.${state || 'missing'}`)
}

function targetReasons(target: { reason_messages?: string[]; reason_codes?: string[] }) {
  const messages = target.reason_messages ?? []
  if (messages.length) {
    return messages
  }
  return target.reason_codes ?? []
}

function formatTime(value?: string) {
  if (!value) {
    return t('site.healthSummary.none')
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function statusClass(status: HealthStatus | string) {
  switch (status) {
    case 'healthy':
    case 'success':
      return 'bg-green-100 text-green-800'
    case 'warning':
      return 'bg-yellow-100 text-yellow-800'
    case 'degraded':
      return 'bg-orange-200 text-orange-900'
    case 'down':
    case 'failure':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}
</script>
