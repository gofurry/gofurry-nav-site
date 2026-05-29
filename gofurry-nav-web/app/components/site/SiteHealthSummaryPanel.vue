<template>
  <div class="rounded-xl bg-orange-50 p-5">
    <div class="grid grid-cols-1 gap-4 text-sm lg:grid-cols-3">
      <div class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ t('site.healthSummary.siteStatus') }}</h4>
        <div class="flex items-stretch rounded-lg bg-orange-100 p-3 md:flex-1">
          <div class="flex w-full flex-col justify-center gap-3">
            <div class="grid grid-cols-2 gap-2">
              <div
                v-for="item in siteSummaryHighlights"
                :key="item.label"
                class="rounded-md bg-orange-50 p-2"
              >
                <div class="mb-1 text-[11px] text-gray-500">{{ item.label }}</div>
                <div class="break-words text-xs font-semibold text-gray-800">{{ item.value }}</div>
              </div>
            </div>
            <div
              v-for="target in targetEntries"
              :key="target.target"
              class="flex flex-1 items-center rounded-md bg-orange-50 p-2"
            >
              <ul v-if="target.reason_messages?.length || target.reason_codes?.length" class="w-full space-y-1 text-xs text-gray-600">
                <li v-for="reason in targetReasons(target)" :key="`${target.target}:${reason}`">{{ reason }}</li>
              </ul>
              <div v-else class="text-xs text-gray-500">{{ t('site.healthSummary.none') }}</div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="targetSummary" class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ targetSectionTitle }}</h4>
        <div class="flex items-stretch rounded-lg bg-orange-100 p-3 md:flex-1">
          <div
            v-if="protocolEntries.length"
            class="grid h-full w-full grid-cols-1 grid-rows-3 gap-2 min-[1680px]:grid-cols-3 min-[1680px]:grid-rows-1"
          >
            <div
              v-for="[protocol, protocolSummary] in protocolEntries"
              :key="protocol"
              class="flex min-h-0 flex-col justify-center rounded-md bg-orange-50 p-2 text-xs"
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

      <div class="md:flex md:flex-col">
        <h4 class="mb-2 text-sm font-bold text-gray-500">{{ securityHeadersTitle }}</h4>
        <div class="rounded-lg bg-orange-100 p-3 md:flex-1">
          <div v-if="securityHeaders.length" class="grid h-full grid-cols-1 grid-rows-6 gap-2 sm:grid-cols-2 sm:grid-rows-3">
            <div
              v-for="item in securityHeaders"
              :key="item.label"
              class="flex min-h-10 items-center justify-between gap-3 rounded-md bg-orange-50 px-3 py-2 text-xs"
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
import type { SiteHealthSummary, TargetHealthSummary } from '~/types/nav'

const props = defineProps<{
  siteSummary: SiteHealthSummary | null
  targetSummary: TargetHealthSummary | null
  currentTarget?: string
  securityHeaders?: { label: string; ok: boolean }[]
}>()

const { locale, t } = useI18n()

const isEnglish = computed(() => locale.value === 'en')
const targetEntries = computed(() => {
  const targets = props.siteSummary?.targets ?? []
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
const siteSummaryHighlights = computed(() => {
  return [
    { label: isEnglish.value ? 'Target' : '当前目标', value: props.currentTarget || props.targetSummary?.target || '-' },
    { label: isEnglish.value ? 'Observed' : '最近观测', value: formatTime(props.targetSummary?.observed_at) },
  ]
})
const securityHeadersTitle = computed(() => (isEnglish.value ? 'Security Headers' : '安全响应头'))
const targetSectionTitle = computed(() => (isEnglish.value ? 'Current Check' : '当前采集'))
const observedAtText = computed(() => (isEnglish.value ? 'Time' : '时间'))
const yesText = computed(() => (isEnglish.value ? 'Yes' : '是'))
const noText = computed(() => (isEnglish.value ? 'No' : '否'))

function statusText(status: string) {
  return t(`site.healthSummary.statuses.${status || 'unknown'}`)
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

</script>
