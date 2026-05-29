<template>
  <div class="rounded-xl bg-orange-50 p-5">
    <div v-if="!isV2Mode" class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="font-semibold">{{ t('site.healthSummary.title') }}</h3>
      </div>
      <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(displayStatus)]">
        {{ statusText(displayStatus) }}
      </span>
    </div>

    <div :class="['grid grid-cols-1 gap-4 text-sm', isV2Mode ? 'lg:grid-cols-3' : 'md:grid-cols-2']">
      <div class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.siteStatus') }}</h4>
        <div :class="['rounded-lg bg-orange-100 p-3 md:flex-1', isV2Mode ? 'flex items-center justify-center' : '']">
          <div v-if="!isV2Mode" class="mb-2">
            <span class="font-bold">{{ t('site.healthSummary.state') }}:</span>
            {{ stateText(siteSummary?.state) }}
          </div>
          <div v-if="!isV2Mode">
            <span class="font-bold">{{ t('site.healthSummary.targetCount') }}:</span>
            {{ siteSummary?.target_count ?? 0 }}
          </div>
          <div v-if="!isV2Mode && statusCountEntries.length" class="mt-2 flex flex-wrap gap-2">
            <span
              v-for="[status, count] in statusCountEntries"
              :key="status"
              class="rounded-full bg-orange-50 px-2 py-0.5 text-xs"
            >
              {{ statusText(status) }} {{ count }}
            </span>
          </div>
          <div v-if="targetEntries.length" :class="[isV2Mode ? 'w-full max-w-md space-y-2' : 'mt-4 space-y-2']">
            <div
              v-for="target in targetEntries"
              :key="target.target"
              class="rounded-md bg-orange-50 p-2"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="break-all font-mono text-xs">{{ target.target }}</span>
                <span v-if="!isV2Mode" :class="['rounded-full px-2 py-0.5 text-xs font-semibold', statusClass(target.status)]">
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
        <div :class="['rounded-lg bg-orange-100 p-3 md:flex-1', isV2Mode ? 'flex items-center justify-center' : '']">
          <div v-if="!isV2Mode" class="break-all font-mono text-xs">{{ targetSummary.target }}</div>
          <div v-if="!isV2Mode" class="mt-2">
            <span class="font-bold">{{ t('site.healthSummary.status') }}:</span>
            {{ statusText(targetSummary.status) }}
          </div>
          <div v-if="protocolEntries.length" :class="[isV2Mode ? 'flex w-full max-w-md flex-wrap justify-center gap-2' : 'mt-2 flex flex-wrap gap-2']">
            <div
              v-for="[protocol, protocolSummary] in protocolEntries"
              :key="protocol"
              class="rounded-md bg-orange-50 p-2 text-xs"
            >
              <div class="mb-1 font-semibold">
                {{ protocol.toUpperCase() }} {{ protocolSummary.stale ? t('site.healthSummary.stale') : statusText(protocolSummary.status) }}
              </div>
              <div>{{ t('site.healthSummary.duration') }}: {{ protocolSummary.duration_ms }}ms</div>
              <div>{{ observedAtText }}: {{ formatTime(protocolSummary.observed_at) }}</div>
              <div>{{ t('site.healthSummary.staleAfter') }}: {{ protocolSummary.stale_after_seconds }}s</div>
              <div v-if="protocolSummary.error_code">
                {{ t('site.healthSummary.errorCode') }}: {{ protocolSummary.error_code }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="isV2Mode" class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ securityHeadersTitle }}</h4>
        <div class="rounded-lg bg-orange-100 p-3 md:flex-1">
          <div v-if="securityHeaders.length" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <div
              v-for="item in securityHeaders"
              :key="item.label"
              class="flex items-center justify-between gap-3 rounded-md bg-orange-50 px-3 py-2 text-xs"
            >
              <span>{{ item.label }}</span>
              <span :class="item.ok ? 'text-green-700' : 'text-gray-400'">
                {{ item.ok ? yesText : noText }}
              </span>
            </div>
          </div>
          <div v-else class="text-xs text-gray-500">{{ t('site.healthSummary.none') }}</div>
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
  currentTarget?: string
  mode?: 'default' | 'v2'
  securityHeaders?: { label: string; ok: boolean }[]
}>()

const { locale, t } = useI18n()

const isV2Mode = computed(() => props.mode === 'v2')
const isEnglish = computed(() => locale.value === 'en')
const displayStatus = computed(() => props.targetSummary?.status || props.siteSummary?.status || 'unknown')
const statusCountEntries = computed(() => Object.entries(props.siteSummary?.status_counts ?? {}).filter(([, count]) => count > 0))
const targetEntries = computed(() => {
  const targets = props.siteSummary?.targets ?? []
  if (!isV2Mode.value) {
    return targets
  }

  const currentTarget = props.currentTarget || props.targetSummary?.target
  const matchedTargets = targets.filter(target => target.target === currentTarget).slice(0, 1)
  if (matchedTargets.length || !props.targetSummary) {
    return matchedTargets
  }

  return [{
    target: props.targetSummary.target,
    status: props.targetSummary.status,
    reason_codes: props.targetSummary.reason_codes,
    reason_messages: props.targetSummary.reason_messages,
    observed_at: props.targetSummary.observed_at,
  }]
})
const protocolEntries = computed(() => Object.entries(props.targetSummary?.protocols ?? {}))
const securityHeaders = computed(() => props.securityHeaders ?? [])
const securityHeadersTitle = computed(() => (isEnglish.value ? 'Security Headers' : '安全响应头'))
const observedAtText = computed(() => (isV2Mode.value ? (isEnglish.value ? 'Time' : '时间') : t('site.healthSummary.observedAt')))
const yesText = computed(() => (isEnglish.value ? 'Yes' : '是'))
const noText = computed(() => (isEnglish.value ? 'No' : '否'))

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
