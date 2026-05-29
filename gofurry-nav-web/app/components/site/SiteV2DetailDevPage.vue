<template>
  <div class="min-h-screen bg-[#f7f1e8] text-slate-900 transition-colors dark:bg-[#090d14] dark:text-slate-100">
    <div v-if="pending" class="flex min-h-[60vh] items-center justify-center text-sm text-slate-500 dark:text-slate-400">
      {{ text.loading }}
    </div>

    <div v-else-if="error || !viewData" class="mx-auto flex min-h-[60vh] max-w-3xl items-center justify-center px-4">
      <div class="w-full rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-950/30 dark:text-red-200">
        {{ text.loadFailed }}
      </div>
    </div>

    <div v-else class="mx-auto flex w-full max-w-[1500px] flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
      <section class="rounded-lg border border-orange-200/70 bg-white/78 p-5 shadow-sm dark:border-white/10 dark:bg-slate-900/82">
        <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0 flex-1">
            <div class="mb-3 flex flex-wrap items-center gap-2">
              <span class="rounded-full bg-orange-100 px-3 py-1 text-xs font-semibold text-orange-800 dark:bg-orange-400/15 dark:text-orange-200">
                V2 DEV
              </span>
              <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(targetSummary?.status || siteSummary?.status)]">
                {{ statusText(targetSummary?.status || siteSummary?.status) }}
              </span>
              <span v-if="targetSummary?.state" class="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                {{ stateText(targetSummary.state) }}
              </span>
            </div>

            <div class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center">
              <div class="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-orange-100 dark:bg-slate-800">
                <img :src="logoSrc" alt="site logo" class="h-full w-full object-contain" @error="onLogoError" />
              </div>
              <div class="min-w-0">
                <h1 class="break-words text-2xl font-bold sm:text-3xl">{{ site.name }}</h1>
                <div class="mt-2 flex flex-wrap items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
                  <button class="break-all font-mono text-orange-700 transition hover:text-orange-500 dark:text-orange-200" @click="copyText(selectedTarget)">
                    {{ selectedTarget }}
                  </button>
                  <span v-if="site.country">/ {{ site.country }}</span>
                  <span>{{ text.visits }} {{ formatNumber(site.view_count) }}</span>
                </div>
              </div>
            </div>

            <p v-if="site.info" class="mt-4 max-w-5xl text-sm leading-6 text-slate-600 dark:text-slate-300">
              {{ site.info }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2 lg:justify-end">
            <NuxtLink
              :to="stableDetailPath"
              class="rounded-lg border border-orange-200 px-4 py-2 text-sm font-semibold text-orange-800 transition hover:bg-orange-50 dark:border-orange-400/25 dark:text-orange-200 dark:hover:bg-orange-400/10"
            >
              {{ text.oldPage }}
            </NuxtLink>
            <a
              :href="visitUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-lg bg-orange-500 px-4 py-2 text-sm font-semibold text-white transition hover:bg-orange-600"
            >
              {{ text.openSite }}
            </a>
          </div>
        </div>
      </section>

      <section v-if="targets.length" class="rounded-lg border border-orange-200/70 bg-white/78 p-4 shadow-sm dark:border-white/10 dark:bg-slate-900/82">
        <div class="mb-3 flex items-center justify-between gap-3">
          <h2 class="text-base font-semibold">{{ text.targets }}</h2>
          <span class="text-xs text-slate-500 dark:text-slate-400">{{ targets.length }}</span>
        </div>
        <div class="flex flex-wrap gap-2">
          <NuxtLink
            v-for="target in targets"
            :key="target.target"
            :to="targetDevPath(target.target)"
            :class="[
              'rounded-lg border px-3 py-2 text-left text-xs transition',
              target.target === selectedTarget
                ? 'border-orange-400 bg-orange-100 text-orange-900 dark:border-orange-300/50 dark:bg-orange-400/15 dark:text-orange-100'
                : 'border-slate-200 bg-white text-slate-600 hover:border-orange-300 hover:text-orange-700 dark:border-white/10 dark:bg-slate-950/40 dark:text-slate-300 dark:hover:border-orange-300/40'
            ]"
          >
            <span class="block break-all font-mono">{{ target.target }}</span>
            <span class="mt-1 block text-[11px] opacity-75">
              {{ target.registered ? text.registered : text.summaryOnly }} / {{ statusText(target.status) }}
            </span>
          </NuxtLink>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-5 xl:grid-cols-[1fr_1.15fr]">
        <div class="rounded-lg border border-orange-200/70 bg-white/78 p-5 shadow-sm dark:border-white/10 dark:bg-slate-900/82">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h2 class="text-base font-semibold">{{ text.siteSummary }}</h2>
            <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(siteSummary?.status)]">
              {{ statusText(siteSummary?.status) }}
            </span>
          </div>

          <div class="grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
            <MetricTile :label="text.targetCount" :value="formatNumber(siteSummary?.target_count)" />
            <MetricTile :label="text.healthy" :value="formatNumber(siteSummary?.status_counts?.healthy)" />
            <MetricTile :label="text.warning" :value="formatNumber(siteSummary?.status_counts?.warning)" />
            <MetricTile :label="text.down" :value="formatNumber(siteSummary?.status_counts?.down)" />
          </div>

          <div v-if="siteSummaryTargets.length" class="mt-4 divide-y divide-slate-200/70 dark:divide-white/10">
            <div v-for="target in siteSummaryTargets" :key="target.target" class="py-3">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="break-all font-mono text-sm">{{ target.target }}</span>
                <span :class="['rounded-full px-2.5 py-1 text-xs font-semibold', statusClass(target.status)]">
                  {{ statusText(target.status) }}
                </span>
              </div>
              <div v-if="target.reason_codes?.length" class="mt-2 flex flex-wrap gap-1.5">
                <span v-for="code in target.reason_codes" :key="`${target.target}:${code}`" class="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                  {{ reasonLabel(code) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-lg border border-orange-200/70 bg-white/78 p-5 shadow-sm dark:border-white/10 dark:bg-slate-900/82">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h2 class="text-base font-semibold">{{ text.currentTarget }}</h2>
            <span :class="['rounded-full px-3 py-1 text-xs font-semibold', statusClass(targetSummary?.status)]">
              {{ statusText(targetSummary?.status) }}
            </span>
          </div>

          <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
            <div v-for="[protocol, item] in protocolEntries" :key="protocol" class="rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-white/10 dark:bg-slate-950/35">
              <div class="mb-2 flex items-center justify-between gap-2">
                <span class="text-xs font-bold uppercase tracking-wide text-slate-500 dark:text-slate-400">{{ protocol }}</span>
                <span :class="['rounded-full px-2 py-0.5 text-[11px] font-semibold', statusClass(item.status)]">
                  {{ item.stale ? text.stale : statusText(item.status) }}
                </span>
              </div>
              <div class="space-y-1 text-xs text-slate-600 dark:text-slate-300">
                <div>{{ text.duration }} {{ formatDuration(item.duration_ms) }}</div>
                <div>{{ text.observedAt }} {{ formatTime(item.observed_at) }}</div>
                <div v-if="item.error_code">{{ text.errorCode }} {{ item.error_code }}</div>
              </div>
            </div>
          </div>

          <div v-if="targetReasonCodes.length" class="mt-4 flex flex-wrap gap-1.5">
            <span v-for="code in targetReasonCodes" :key="code" class="rounded-full bg-orange-100 px-2.5 py-1 text-xs text-orange-800 dark:bg-orange-400/15 dark:text-orange-100">
              {{ reasonLabel(code) }}
            </span>
          </div>

          <div v-if="targetSummary?.edge_provider_hints?.length" class="mt-4">
            <h3 class="mb-2 text-sm font-semibold text-slate-500 dark:text-slate-400">{{ text.edgeHints }}</h3>
            <div class="flex flex-wrap gap-2">
              <span v-for="hint in targetSummary.edge_provider_hints" :key="`${hint.provider}:${hint.hint_type}`" class="rounded-lg bg-slate-100 px-3 py-2 text-xs dark:bg-slate-800">
                {{ providerText(hint.provider) }} / {{ hint.hint_type }} / {{ hint.confidence }}
              </span>
            </div>
          </div>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <DataPanel :title="text.pingPanel" :status="coreEnvelope('ping')?.status">
          <MetricGrid :items="pingMetrics" />
        </DataPanel>

        <DataPanel :title="text.httpPanel" :status="coreEnvelope('http')?.status">
          <MetricGrid :items="httpMetrics" />
          <KeyValueList class="mt-4" :items="httpHeaderItems" :empty-text="text.none" />
        </DataPanel>

        <DataPanel :title="text.tlsPanel" :status="coreEnvelope('http')?.status">
          <MetricGrid :items="tlsMetrics" />
        </DataPanel>
      </section>

      <section class="grid grid-cols-1 gap-5 xl:grid-cols-[1.1fr_0.9fr]">
        <DataPanel :title="text.dnsPanel" :status="coreEnvelope('dns')?.status">
          <div v-if="dnsRiskFlags.length" class="mb-4 flex flex-wrap gap-2">
            <span v-for="flag in dnsRiskFlags" :key="flag" class="rounded-full bg-yellow-100 px-2.5 py-1 text-xs text-yellow-800 dark:bg-yellow-400/15 dark:text-yellow-100">
              {{ flag }}
            </span>
          </div>
          <MetricGrid :items="dnsMetrics" />
          <div v-if="dnsRecordGroups.length" class="mt-4 space-y-4">
            <div v-for="group in dnsRecordGroups" :key="group.type">
              <h3 class="mb-2 text-sm font-semibold text-slate-500 dark:text-slate-400">{{ group.type }}</h3>
              <div class="overflow-hidden rounded-lg border border-slate-200 dark:border-white/10">
                <div v-for="record in group.records" :key="record.key" class="grid gap-2 border-b border-slate-200 px-3 py-2 text-xs last:border-b-0 dark:border-white/10 md:grid-cols-[80px_1fr_90px]">
                  <span class="font-semibold">{{ record.type }}</span>
                  <span class="break-all font-mono">{{ record.value }}</span>
                  <span class="text-slate-500 dark:text-slate-400">TTL {{ record.ttl }}</span>
                </div>
              </div>
            </div>
          </div>
        </DataPanel>

        <DataPanel :title="text.pageInfoPanel" :status="coreEnvelope('http')?.status">
          <KeyValueList :items="pageInfoItems" :empty-text="text.none" />
          <div v-if="securityHeaderItems.length" class="mt-4">
            <h3 class="mb-2 text-sm font-semibold text-slate-500 dark:text-slate-400">{{ text.securityHeaders }}</h3>
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <div
                v-for="item in securityHeaderItems"
                :key="item.label"
                class="flex items-center justify-between gap-2 rounded-lg bg-slate-50 px-3 py-2 text-xs dark:bg-slate-950/35"
              >
                <span>{{ item.label }}</span>
                <span :class="item.ok ? 'text-green-600 dark:text-green-300' : 'text-slate-400'">
                  {{ item.ok ? text.yes : text.no }}
                </span>
              </div>
            </div>
          </div>
        </DataPanel>
      </section>

      <section class="grid grid-cols-1 gap-5 xl:grid-cols-[0.95fr_1.05fr]">
        <DataPanel :title="text.lightProbePanel">
          <div v-if="lightProbeEntries.length" class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div v-for="probe in lightProbeEntries" :key="probe.protocol" class="rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-white/10 dark:bg-slate-950/35">
              <div class="mb-2 flex items-center justify-between gap-2">
                <span class="text-sm font-semibold">{{ protocolName(probe.protocol) }}</span>
                <span :class="['rounded-full px-2 py-0.5 text-[11px] font-semibold', statusClass(probe.status)]">
                  {{ statusText(probe.status) }}
                </span>
              </div>
              <KeyValueList :items="probe.items" :empty-text="text.none" compact />
            </div>
          </div>
          <EmptyLine v-else :text="text.none" />
        </DataPanel>

        <DataPanel :title="text.trendPanel">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div v-for="window in trendWindows" :key="window.name" class="rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-white/10 dark:bg-slate-950/35">
              <h3 class="mb-3 text-sm font-semibold">{{ window.name }}</h3>
              <div class="space-y-3">
                <div v-for="protocol in window.protocols" :key="`${window.name}:${protocol.name}`" class="text-xs">
                  <div class="mb-1 font-semibold uppercase text-slate-500 dark:text-slate-400">{{ protocol.name }}</div>
                  <div class="grid grid-cols-2 gap-2">
                    <span>{{ text.count }} {{ protocol.count }}</span>
                    <span>{{ text.successRate }} {{ protocol.successRate }}</span>
                    <span>{{ text.avgDuration }} {{ protocol.avgDuration }}</span>
                    <span>{{ text.p95Duration }} {{ protocol.p95Duration }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <EmptyLine v-if="!trendWindows.length" :text="text.none" />
        </DataPanel>
      </section>

      <section class="grid grid-cols-1 gap-5 xl:grid-cols-[1fr_1fr]">
        <DataPanel :title="text.changePanel">
          <div v-if="changeEvents.length" class="divide-y divide-slate-200 dark:divide-white/10">
            <div v-for="event in changeEvents" :key="event.key" class="py-3 text-sm">
              <div class="mb-1 flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-semibold uppercase text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                  {{ event.protocol }}
                </span>
                <span class="font-semibold">{{ event.field }}</span>
                <span class="text-xs text-slate-500 dark:text-slate-400">{{ formatTime(event.detectedAt) }}</span>
              </div>
              <div class="grid gap-2 text-xs md:grid-cols-2">
                <div class="rounded-lg bg-slate-50 p-2 dark:bg-slate-950/35">{{ text.oldValue }} {{ event.oldValue }}</div>
                <div class="rounded-lg bg-slate-50 p-2 dark:bg-slate-950/35">{{ text.newValue }} {{ event.newValue }}</div>
              </div>
            </div>
          </div>
          <EmptyLine v-else :text="text.none" />
        </DataPanel>

        <DataPanel :title="text.historyPanel">
          <div class="space-y-4">
            <div v-for="history in observationHistories" :key="history.protocol">
              <h3 class="mb-2 text-sm font-semibold uppercase text-slate-500 dark:text-slate-400">{{ history.protocol }}</h3>
              <div v-if="history.items.length" class="overflow-hidden rounded-lg border border-slate-200 dark:border-white/10">
                <div
                  v-for="item in history.items"
                  :key="`${history.protocol}:${item.observed_at}:${item.job_id || ''}`"
                  class="grid gap-2 border-b border-slate-200 px-3 py-2 text-xs last:border-b-0 dark:border-white/10 md:grid-cols-[90px_100px_1fr]"
                >
                  <span :class="['w-fit rounded-full px-2 py-0.5 font-semibold', statusClass(item.status)]">{{ statusText(item.status) }}</span>
                  <span>{{ formatDuration(item.duration_ms) }}</span>
                  <span class="text-slate-500 dark:text-slate-400">{{ formatTime(item.observed_at) }}</span>
                </div>
              </div>
              <EmptyLine v-else :text="text.none" />
            </div>
          </div>
        </DataPanel>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme'
import type {
  CollectorEnvelope,
  SiteHealthSummary,
  SiteV2DetailResponse,
  TargetChangesResponse,
  TargetHealthSummary,
  TargetLatestResponse,
  TargetObservationsResponse,
  TargetTrendResponse,
} from '~/types/nav'

type Primitive = string | number | boolean | null | undefined
type MetricItem = { label: string; value: Primitive; accent?: boolean }
type KeyValueItem = { label: string; value: Primitive | string[] }

interface SiteV2DevData {
  detail: SiteV2DetailResponse
  siteSummary: SiteHealthSummary | null
  targetSummary: TargetHealthSummary | null
  latestCore: TargetLatestResponse | null
  lightProbeState: TargetLatestResponse | null
  trend: TargetTrendResponse | null
  changes: TargetChangesResponse | null
  observations: Record<'ping' | 'http' | 'dns', TargetObservationsResponse | null>
}

const MetricTile = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: '-' },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg bg-slate-50 p-3 dark:bg-slate-950/35' }, [
      h('div', { class: 'text-[11px] font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400' }, props.label),
      h('div', { class: 'mt-1 text-lg font-bold' }, String(props.value ?? '-')),
    ])
  },
})

const DataPanel = defineComponent({
  props: {
    title: { type: String, required: true },
    status: { type: String, default: '' },
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'rounded-lg border border-orange-200/70 bg-white/78 p-5 shadow-sm dark:border-white/10 dark:bg-slate-900/82' }, [
      h('div', { class: 'mb-4 flex items-center justify-between gap-3' }, [
        h('h2', { class: 'text-base font-semibold' }, props.title),
        props.status ? h('span', { class: statusClass(props.status) + ' rounded-full px-3 py-1 text-xs font-semibold' }, statusText(props.status)) : null,
      ]),
      slots.default?.(),
    ])
  },
})

const MetricGrid = defineComponent({
  props: {
    items: { type: Array as PropType<MetricItem[]>, default: () => [] },
  },
  setup(props) {
    return () => props.items.length
      ? h('div', { class: 'grid grid-cols-1 gap-3 sm:grid-cols-2' }, props.items.map((item) => (
        h('div', { class: item.accent ? 'rounded-lg bg-orange-50 p-3 dark:bg-orange-400/10' : 'rounded-lg bg-slate-50 p-3 dark:bg-slate-950/35' }, [
          h('div', { class: 'text-[11px] font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400' }, item.label),
          h('div', { class: 'mt-1 break-words font-semibold' }, displayValue(item.value)),
        ])
      )))
      : h(EmptyLine, { text: '-' })
  },
})

const KeyValueList = defineComponent({
  props: {
    items: { type: Array as PropType<KeyValueItem[]>, default: () => [] },
    emptyText: { type: String, default: '-' },
    compact: { type: Boolean, default: false },
  },
  setup(props) {
    return () => props.items.length
      ? h('div', { class: props.compact ? 'space-y-1.5 text-xs' : 'space-y-2 text-sm' }, props.items.map((item) => (
        h('div', { class: 'grid gap-1 sm:grid-cols-[150px_1fr]' }, [
          h('span', { class: 'font-semibold text-slate-500 dark:text-slate-400' }, item.label),
          h('span', { class: 'break-words font-mono text-slate-700 dark:text-slate-200' }, Array.isArray(item.value) ? item.value.join(', ') : displayValue(item.value)),
        ])
      )))
      : h(EmptyLine, { text: props.emptyText })
  },
})

const EmptyLine = defineComponent({
  props: {
    text: { type: String, default: '-' },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-400 dark:bg-slate-950/35 dark:text-slate-500' }, props.text)
  },
})

const route = useRoute()
const { locale } = useI18n()
const themeStore = useThemeStore()
const navV2Api = useApi('navV2')
const runtimeConfig = useRuntimeConfig()

const siteId = computed(() => Number(route.params.id || 0))
const routeTarget = computed(() => extractRouteParam(route.params.domain) || extractDevTargetFromPath(route.path))
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
const logoPrefix = computed(() => runtimeConfig.public.siteLogoPrefixUrl || '')
const defaultLogo = computed(() => runtimeConfig.public.siteDefaultLogo || `${logoPrefix.value}defaultLogo.svg`)

const { data, pending, error, refresh } = await useAsyncData<SiteV2DevData>(
  () => `site-v2-dev:${siteId.value}:${routeTarget.value}:${lang.value}`,
  async () => {
    if (!siteId.value || !routeTarget.value) {
      throw new Error('invalid route')
    }

    const detail = await navV2Api<SiteV2DetailResponse>(`/nav/sites/${siteId.value}/detail`, {
      query: {
        lang: lang.value,
        target: routeTarget.value,
        payload_mode: 'preview',
      },
    })

    const target = detail.selected_target || routeTarget.value
    const encodedTarget = encodeURIComponent(target)
    const targetBase = `/nav/sites/${siteId.value}/targets/${encodedTarget}`

    const [
      siteSummary,
      targetSummary,
      latestCore,
      lightProbeState,
      trend,
      changes,
      pingObservations,
      httpObservations,
      dnsObservations,
    ] = await Promise.all([
      safeRequest(() => navV2Api<SiteHealthSummary>(`/nav/sites/${siteId.value}/summary`)),
      safeRequest(() => navV2Api<TargetHealthSummary>(`${targetBase}/summary`)),
      safeRequest(() => navV2Api<TargetLatestResponse>(`${targetBase}/latest`, { query: { payload_mode: 'preview' } })),
      safeRequest(() => navV2Api<TargetLatestResponse>(`${targetBase}/light-probes`, { query: { payload_mode: 'preview' } })),
      safeRequest(() => navV2Api<TargetTrendResponse>(`${targetBase}/trend`)),
      safeRequest(() => navV2Api<TargetChangesResponse>(`${targetBase}/changes`)),
      safeRequest(() => navV2Api<TargetObservationsResponse>(`${targetBase}/observations`, { query: { protocol: 'ping', limit: 8, payload_mode: 'preview' } })),
      safeRequest(() => navV2Api<TargetObservationsResponse>(`${targetBase}/observations`, { query: { protocol: 'http', limit: 8, payload_mode: 'preview' } })),
      safeRequest(() => navV2Api<TargetObservationsResponse>(`${targetBase}/observations`, { query: { protocol: 'dns', limit: 8, payload_mode: 'preview' } })),
    ])

    return {
      detail,
      siteSummary: siteSummary ?? detail.site_summary ?? null,
      targetSummary: targetSummary ?? detail.target_summary ?? null,
      latestCore: latestCore ?? detail.latest_core ?? null,
      lightProbeState: lightProbeState ?? detail.light_probe_state ?? null,
      trend: trend ?? detail.derived?.trend ?? null,
      changes: changes ?? detail.derived?.changes ?? null,
      observations: {
        ping: pingObservations,
        http: httpObservations,
        dns: dnsObservations,
      },
    }
  },
  {
    watch: [siteId, routeTarget, lang],
  }
)

const text = computed(() => {
  const en = locale.value === 'en'
  return en
    ? {
        loading: 'Loading site v2 detail...',
        loadFailed: 'Failed to load v2 site detail.',
        oldPage: 'Current page',
        openSite: 'Open site',
        visits: 'Visits',
        targets: 'Collector targets',
        registered: 'registered',
        summaryOnly: 'summary only',
        siteSummary: 'Site summary',
        currentTarget: 'Current target',
        targetCount: 'Targets',
        healthy: 'Healthy',
        warning: 'Warning',
        down: 'Down',
        stale: 'Stale',
        duration: 'Duration',
        observedAt: 'Observed',
        errorCode: 'Error',
        edgeHints: 'Edge hints',
        pingPanel: 'Ping observation',
        httpPanel: 'HTTP observation',
        tlsPanel: 'TLS observation',
        dnsPanel: 'DNS observation',
        pageInfoPanel: 'Page metadata',
        lightProbePanel: 'Light probes',
        trendPanel: 'Trend',
        changePanel: 'Changes',
        historyPanel: 'Observation history',
        securityHeaders: 'Security headers',
        yes: 'Yes',
        no: 'No',
        none: 'None',
        count: 'Count',
        successRate: 'Success',
        avgDuration: 'Avg',
        p95Duration: 'P95',
        oldValue: 'Old:',
        newValue: 'New:',
      }
    : {
        loading: '正在加载 v2 站点详情...',
        loadFailed: '站点 v2 详情加载失败。',
        oldPage: '当前详情页',
        openSite: '打开站点',
        visits: '访问量',
        targets: '采集目标',
        registered: '已登记',
        summaryOnly: '仅摘要',
        siteSummary: '站点摘要',
        currentTarget: '当前采集目标',
        targetCount: '目标数',
        healthy: '健康',
        warning: '需关注',
        down: '不可用',
        stale: '过期',
        duration: '耗时',
        observedAt: '观测时间',
        errorCode: '错误码',
        edgeHints: '边缘服务线索',
        pingPanel: 'Ping 观测',
        httpPanel: 'HTTP 观测',
        tlsPanel: 'TLS 观测',
        dnsPanel: 'DNS 观测',
        pageInfoPanel: '页面元信息',
        lightProbePanel: '低频轻探测',
        trendPanel: '趋势摘要',
        changePanel: '变化事件',
        historyPanel: '观测历史',
        securityHeaders: '安全响应头',
        yes: '是',
        no: '否',
        none: '暂无数据',
        count: '数量',
        successRate: '成功率',
        avgDuration: '平均',
        p95Duration: 'P95',
        oldValue: '旧值：',
        newValue: '新值：',
      }
})

const viewData = computed(() => data.value)
const detail = computed(() => viewData.value?.detail)
const site = computed(() => detail.value?.site ?? {
  id: siteId.value,
  name: '',
  info: '',
  icon: null,
  country: null,
  nsfw: '0',
  welfare: '0',
  view_count: 0,
})
const targets = computed(() => detail.value?.targets ?? [])
const selectedTarget = computed(() => detail.value?.selected_target || routeTarget.value)
const siteSummary = computed(() => viewData.value?.siteSummary)
const targetSummary = computed(() => viewData.value?.targetSummary)
const latestCore = computed(() => viewData.value?.latestCore)
const lightProbeState = computed(() => viewData.value?.lightProbeState)
const protocolEntries = computed(() => Object.entries(targetSummary.value?.protocols ?? {}))
const targetReasonCodes = computed(() => targetSummary.value?.reason_codes ?? [])
const siteSummaryTargets = computed(() => siteSummary.value?.targets ?? [])
const coreProtocols = computed(() => latestCore.value?.protocols ?? detail.value?.latest_core?.protocols ?? {})
const lightProtocols = computed(() => lightProbeState.value?.protocols ?? detail.value?.light_probe_state?.protocols ?? {})
const httpPayload = computed(() => envelopePayload(coreEnvelope('http')))
const pingPayload = computed(() => envelopePayload(coreEnvelope('ping')))
const dnsPayload = computed(() => envelopePayload(coreEnvelope('dns')))
const logoSrc = computed(() => {
  const icon = site.value.icon || ''
  if (!icon) {
    return defaultLogo.value
  }
  if (/^https?:\/\//i.test(icon)) {
    return icon
  }
  return `${logoPrefix.value.replace(/\/$/, '')}/${icon}`
})
const selectedTargetMeta = computed(() => targets.value.find((item) => item.target === selectedTarget.value))
const visitUrl = computed(() => `${selectedTargetMeta.value?.tls === '0' ? 'http' : 'https'}://${selectedTarget.value}`)
const stableDetailPath = computed(() => `/site/${siteId.value}/${encodeURIComponent(selectedTarget.value)}`)

const pingMetrics = computed<MetricItem[]>(() => [
  { label: 'ICMP', value: stringValue(pingPayload.value.icmp_status) },
  { label: 'Avg RTT', value: msValue(pingPayload.value.avg_rtt_ms), accent: true },
  { label: 'Min RTT', value: msValue(pingPayload.value.min_rtt_ms) },
  { label: 'Max RTT', value: msValue(pingPayload.value.max_rtt_ms) },
  { label: 'Jitter', value: msValue(pingPayload.value.jitter_ms) },
  { label: 'Loss', value: percentValue(pingPayload.value.loss_rate) },
  { label: 'Packets', value: `${formatNumber(pingPayload.value.packets_recv)}/${formatNumber(pingPayload.value.packets_sent)}` },
  { label: 'Resolved IP', value: stringValue(pingPayload.value.selected_ip || pingPayload.value.resolved_ip) },
])

const httpMetrics = computed<MetricItem[]>(() => [
  { label: 'Status', value: numberValue(httpPayload.value.status_code), accent: true },
  { label: 'Response', value: msValue(httpPayload.value.response_time_ms) },
  { label: 'DNS Lookup', value: msValue(httpPayload.value.dns_lookup_ms) },
  { label: 'TCP Connect', value: msValue(httpPayload.value.tcp_connect_ms) },
  { label: 'TLS Handshake', value: msValue(httpPayload.value.tls_handshake_ms) },
  { label: 'TTFB', value: msValue(httpPayload.value.ttfb_ms) },
  { label: 'Transfer', value: msValue(httpPayload.value.transfer_ms) },
  { label: 'Body', value: formatBytes(Number(httpPayload.value.body_read_bytes ?? 0)) },
  { label: 'Protocol', value: stringValue(httpPayload.value.http_protocol) },
  { label: 'Remote IP', value: stringValue(httpPayload.value.remote_ip) },
  { label: 'Content-Type', value: stringValue(httpPayload.value.content_type) },
  { label: 'Final URL', value: stringValue(httpPayload.value.final_url || httpPayload.value.url) },
])

const tlsMetrics = computed<MetricItem[]>(() => [
  { label: 'Collected', value: boolText(httpPayload.value.cert_collected) },
  { label: 'Verified', value: boolText(httpPayload.value.cert_verified), accent: true },
  { label: 'Handshake', value: stringValue(httpPayload.value.tls_handshake) },
  { label: 'Verify Error', value: stringValue(httpPayload.value.verify_error_category || httpPayload.value.verify_error) },
  { label: 'TLS Version', value: stringValue(httpPayload.value.tls_version) },
  { label: 'Cipher', value: stringValue(httpPayload.value.cipher_suite) },
  { label: 'Not Before', value: dateValue(httpPayload.value.cert_not_before) },
  { label: 'Not After', value: dateValue(httpPayload.value.cert_not_after || httpPayload.value.cert_expiry) },
  { label: 'Days Left', value: numberValue(httpPayload.value.cert_days_left) },
  { label: 'Issuer', value: stringValue(httpPayload.value.cert_issuer_cn || httpPayload.value.cert_issuer) },
  { label: 'SAN Count', value: numberValue(httpPayload.value.cert_san_count) },
  { label: 'Chain Length', value: numberValue(httpPayload.value.cert_chain_length) },
])

const dnsRiskFlags = computed(() => stringArray(dnsPayload.value.risk_flags))
const dnsMetrics = computed<MetricItem[]>(() => [
  { label: 'A', value: numberValue(dnsPayload.value.ipv4_count) },
  { label: 'AAAA', value: numberValue(dnsPayload.value.ipv6_count) },
  { label: 'CNAME Depth', value: numberValue(dnsPayload.value.cname_chain_depth) },
  { label: 'CNAME Terminal', value: stringValue(dnsPayload.value.cname_terminal) },
  { label: 'TTL Spread', value: numberValue(dnsPayload.value.ttl_spread) },
  { label: 'Record Budget', value: boolText(dnsPayload.value.record_budget_exhausted) },
  { label: 'MX Hosts', value: stringArray(dnsPayload.value.mx_hosts).join(', ') || '-' },
  { label: 'NS Hosts', value: stringArray(dnsPayload.value.name_server_hosts).join(', ') || '-' },
])

const dnsRecordGroups = computed(() => {
  const groups: { type: string; records: { key: string; type: string; value: string; ttl: string }[] }[] = []
  for (const type of ['A', 'AAAA', 'CNAME', 'MX', 'NS', 'TXT', 'CAA', 'SOA']) {
    const rows = arrayValue(dnsPayload.value[type]).slice(0, 8)
    if (!rows.length) {
      continue
    }
    groups.push({
      type,
      records: rows.map((row, index) => {
        const item = asRecord(row)
        return {
          key: `${type}:${index}:${stringValue(item.value)}`,
          type: stringValue(item.type || type),
          value: stringValue(item.value || item.host || item.ns || item.mbox),
          ttl: displayValue(item.ttl),
        }
      }),
    })
  }
  return groups
})

const httpHeaderItems = computed<KeyValueItem[]>(() => {
  const headers = asRecord(httpPayload.value.headers)
  const keys = ['server', 'content-type', 'content-language', 'cache-control', 'etag', 'last-modified', 'vary', 'content-encoding', 'x-robots-tag', 'alt-svc', 'x-powered-by']
  return keys
    .map((key) => ({ label: key, value: headerValue(headers, key) }))
    .filter((item) => item.value !== '-')
})

const pageInfoItems = computed<KeyValueItem[]>(() => {
  const meta = asRecord(httpPayload.value.meta)
  const og = asRecord(httpPayload.value.open_graph || httpPayload.value.openGraph)
  const twitter = asRecord(httpPayload.value.twitter_card || httpPayload.value.twitterCard)
  const cachePolicy = asRecord(httpPayload.value.cache_policy)
  const serverHints = asRecord(httpPayload.value.server_hints || httpPayload.value.serverHints)
  const cookieSummary = asRecord(httpPayload.value.cookie_summary || httpPayload.value.cookieSummary)

  return compactItems([
    { label: 'Title', value: stringValue(httpPayload.value.title) },
    { label: 'Description', value: stringValue(meta.description) },
    { label: 'Keywords', value: stringValue(meta.keywords) },
    { label: 'Author', value: stringValue(meta.author) },
    { label: 'Generator', value: stringValue(meta.generator || serverHints.generator) },
    { label: 'Application', value: stringValue(meta.application_name) },
    { label: 'Theme Color', value: stringValue(meta.theme_color) },
    { label: 'Robots', value: stringValue(meta.robots || httpPayload.value.robots_meta_policy) },
    { label: 'Viewport', value: stringValue(meta.viewport) },
    { label: 'Canonical', value: stringValue(httpPayload.value.canonical_url || httpPayload.value.canonicalUrl) },
    { label: 'HTML Lang', value: stringValue(httpPayload.value.html_lang || httpPayload.value.htmlLang) },
    { label: 'OpenGraph', value: stringValue(og.title || og.description || og.image) },
    { label: 'Twitter Card', value: stringValue(twitter.card || twitter.title || twitter.description) },
    { label: 'Cookie', value: cookieSummary.count == null ? '-' : `${cookieSummary.count} / Secure ${formatNumber(cookieSummary.secure_count)} / HttpOnly ${formatNumber(cookieSummary.http_only_count)}` },
    { label: 'Cache-Control', value: stringValue(cachePolicy.cache_control) },
    { label: 'ETag', value: stringValue(cachePolicy.etag) },
  ])
})

const securityHeaderItems = computed(() => {
  const compactSummary = asRecord(httpPayload.value.security_headers)
  if (Object.keys(compactSummary).length) {
    return Object.entries({
      HSTS: compactSummary.strict_transport_security,
      CSP: compactSummary.content_security_policy,
      'X-Frame-Options': compactSummary.x_frame_options,
      'X-Content-Type-Options': compactSummary.x_content_type_options,
      'Referrer-Policy': compactSummary.referrer_policy,
      'Permissions-Policy': compactSummary.permissions_policy,
    }).map(([label, value]) => ({ label, ok: Boolean(value) }))
  }

  const summary = asRecord(httpPayload.value.security_header_summary)
  return Object.entries({
    HSTS: asRecord(summary.hsts).present,
    CSP: asRecord(summary.content_security_policy).present,
    'X-Frame-Options': asRecord(summary.x_frame_options).present,
    'X-Content-Type-Options': asRecord(summary.x_content_type_options).present,
    'Referrer-Policy': asRecord(summary.referrer_policy).present,
    'Permissions-Policy': asRecord(summary.permissions_policy).present,
  }).map(([label, value]) => ({ label, ok: Boolean(value) }))
})

const lightProbeEntries = computed(() => Object.entries(lightProtocols.value).map(([protocol, envelope]) => ({
  protocol,
  status: envelope.status,
  items: lightProbeItems(protocol, envelope),
})))

const trendWindows = computed(() => {
  const windows = asRecord(viewData.value?.trend?.windows)
  return Object.entries(windows).map(([name, value]) => {
    const windowValue = asRecord(value)
    const protocols = asRecord(windowValue.protocols)
    return {
      name,
      protocols: Object.entries(protocols).map(([protocol, raw]) => {
        const item = asRecord(raw)
        return {
          name: protocol,
          count: formatNumber(item.observation_count),
          successRate: percentValue(item.success_rate),
          avgDuration: msValue(item.avg_duration_ms),
          p95Duration: msValue(item.p95_duration_ms),
        }
      }),
    }
  }).filter((item) => item.protocols.length)
})

const changeEvents = computed(() => arrayValue(viewData.value?.changes?.events).slice(0, 12).map((raw, index) => {
  const item = asRecord(raw)
  return {
    key: stringValue(item.event_id || index),
    protocol: stringValue(item.protocol),
    field: stringValue(item.field),
    oldValue: displayValue(item.old_value),
    newValue: displayValue(item.new_value),
    detectedAt: stringValue(item.detected_at),
  }
}))

const observationHistories = computed(() => (['ping', 'http', 'dns'] as const).map((protocol) => ({
  protocol,
  items: viewData.value?.observations[protocol]?.items ?? [],
})))

onMounted(() => {
  themeStore.initTheme()
})

useSeoMeta({
  title: () => `${site.value.name || selectedTarget.value} - GoFurry`,
  description: () => site.value.info?.slice(0, 160) || selectedTarget.value,
})

function coreEnvelope(protocol: string) {
  return coreProtocols.value[protocol]
}

function targetDevPath(target: string) {
  return `/site/${siteId.value}/${encodeURIComponent(target)}/dev`
}

function onLogoError(event: Event) {
  const target = event.target as HTMLImageElement
  target.src = defaultLogo.value
}

async function safeRequest<T>(request: () => Promise<T>): Promise<T | null> {
  try {
    return await request()
  } catch {
    return null
  }
}

function extractRouteParam(value: unknown): string {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string') {
    return ''
  }

  try {
    return decodeURIComponent(rawValue).trim()
  } catch {
    return rawValue.trim()
  }
}

function extractDevTargetFromPath(path: string): string {
  const parts = path.split('/').filter(Boolean)
  const siteIndex = parts.findIndex((part) => part === 'site')
  const targetPart = siteIndex >= 0 ? parts[siteIndex + 2] : ''
  if (!targetPart || targetPart === 'dev') {
    return ''
  }
  return extractRouteParam(targetPart)
}

function asRecord(value: unknown): Record<string, any> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {}
  }
  return value as Record<string, any>
}

function arrayValue(value: unknown): any[] {
  return Array.isArray(value) ? value : []
}

function envelopePayload(envelope?: CollectorEnvelope): Record<string, any> {
  return asRecord(envelope?.payload)
}

function stringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.map((item) => String(item)).filter(Boolean)
}

function statusText(status?: string) {
  const map: Record<'zh' | 'en', Record<string, string>> = {
    zh: {
      ready: '可用',
      missing: '缺失',
      stale: '已过期',
      healthy: '健康',
      warning: '需关注',
      degraded: '降级',
      unknown: '未知',
      down: '不可用',
      success: '成功',
      failure: '失败',
      skipped: '跳过',
    },
    en: {
      ready: 'Ready',
      missing: 'Missing',
      stale: 'Stale',
      healthy: 'Healthy',
      warning: 'Warning',
      degraded: 'Degraded',
      unknown: 'Unknown',
      down: 'Down',
      success: 'Success',
      failure: 'Failure',
      skipped: 'Skipped',
    },
  }
  const langKey: 'zh' | 'en' = locale.value === 'en' ? 'en' : 'zh'
  const dict = map[langKey]
  return dict[status || 'unknown'] || status || dict.unknown
}

function stateText(state?: string) {
  return statusText(state)
}

function statusClass(status?: string) {
  switch (status) {
    case 'healthy':
    case 'success':
    case 'ready':
      return 'bg-green-100 text-green-800 dark:bg-green-400/15 dark:text-green-200'
    case 'warning':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-400/15 dark:text-yellow-100'
    case 'degraded':
      return 'bg-orange-100 text-orange-900 dark:bg-orange-400/15 dark:text-orange-100'
    case 'down':
    case 'failure':
      return 'bg-red-100 text-red-800 dark:bg-red-400/15 dark:text-red-100'
    case 'stale':
    case 'missing':
    case 'unknown':
    default:
      return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
  }
}

function reasonLabel(code: string) {
  const zh: Record<string, string> = {
    http_missing_or_stale: 'HTTP 缺失或过期',
    http_failed: 'HTTP 失败',
    dns_failed: 'DNS 失败',
    dns_missing_or_stale: 'DNS 缺失或过期',
    dns_failed_but_http_ok: 'DNS 异常但 HTTP 可访问',
    ping_failed_but_http_ok: 'Ping 异常但 HTTP 可访问',
    dns_risk_private_ip: 'DNS 私网 IP',
    dns_risk_low_ttl: 'DNS TTL 偏低',
    dns_risk_nxdomain_with_answer: 'NXDOMAIN 带响应',
    dns_risk_ptr_empty: 'PTR 为空',
    tls_verify_expired: 'TLS 证书过期',
    tls_verify_not_yet_valid: 'TLS 证书尚未生效',
    tls_verify_hostname_mismatch: 'TLS 域名不匹配',
    tls_verify_unknown_authority: 'TLS 证书链不受信',
    tls_verify_incompatible_usage: 'TLS 用途不兼容',
    tls_verify_other: 'TLS 校验异常',
    tls_cert_expired: 'TLS 证书已过期',
    tls_cert_expiring_soon: 'TLS 证书即将过期',
    no_target_summary: '无目标摘要',
    all_targets_down: '全部目标不可用',
    all_targets_unknown: '全部目标未知',
    some_targets_degraded: '部分目标降级',
    some_targets_warning: '部分目标需关注',
  }
  return locale.value === 'en' ? code : zh[code] || code
}

function providerText(provider: string) {
  return provider.replace(/_/g, ' ')
}

function protocolName(protocol: string) {
  const map: Record<string, string> = {
    rdap: 'RDAP',
    robots: 'robots.txt',
    security_txt: 'security.txt',
    page_assets: 'Page assets',
    port_check: 'Port check',
    waf_canary: 'WAF canary',
  }
  return map[protocol] || protocol
}

function displayValue(value: Primitive | unknown): string {
  if (value == null || value === '') {
    return '-'
  }
  if (typeof value === 'boolean') {
    return boolText(value)
  }
  if (typeof value === 'number') {
    return Number.isFinite(value) ? String(value) : '-'
  }
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => displayValue(item)).join(', ') : '-'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

function stringValue(value: unknown): string {
  const text = displayValue(value)
  return text.length > 220 ? `${text.slice(0, 220)}...` : text
}

function numberValue(value: unknown): string {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return Number.isInteger(value) ? String(value) : value.toFixed(2)
  }
  if (typeof value === 'string' && value.trim() !== '') {
    return value
  }
  return '-'
}

function formatNumber(value: unknown): string {
  const num = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(num) ? num.toLocaleString() : '-'
}

function msValue(value: unknown): string {
  const num = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(num) ? `${num.toFixed(num >= 10 ? 0 : 2)}ms` : '-'
}

function percentValue(value: unknown): string {
  const num = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(num)) {
    return '-'
  }
  const percent = num <= 1 ? num * 100 : num
  return `${percent.toFixed(percent >= 10 ? 1 : 2)}%`
}

function boolText(value: unknown): string {
  if (value === true) {
    return text.value.yes
  }
  if (value === false) {
    return text.value.no
  }
  return '-'
}

function formatDuration(value: unknown): string {
  return msValue(value)
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '-'
  }
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

function formatTime(value?: string) {
  if (!value || value.startsWith('0001-01-01')) {
    return '-'
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function dateValue(value: unknown): string {
  return typeof value === 'string' ? formatTime(value) : '-'
}

function headerValue(headers: Record<string, any>, key: string) {
  const matchedKey = Object.keys(headers).find((item) => item.toLowerCase() === key.toLowerCase())
  if (!matchedKey) {
    return '-'
  }
  const value = headers[matchedKey]
  return Array.isArray(value) ? value.join(', ') : stringValue(value)
}

function compactItems(items: KeyValueItem[]): KeyValueItem[] {
  return items.filter((item) => displayValue(item.value) !== '-')
}

function lightProbeItems(protocol: string, envelope: CollectorEnvelope): KeyValueItem[] {
  const payload = envelopePayload(envelope)
  switch (protocol) {
    case 'rdap':
      return compactItems([
        { label: 'Domain', value: stringValue(payload.registrable_domain) },
        { label: 'Registrar', value: stringValue(payload.registrar) },
        { label: 'Expires', value: dateValue(payload.expires_at) },
        { label: 'Statuses', value: stringArray(payload.statuses) },
        { label: 'Nameservers', value: stringArray(payload.nameservers).slice(0, 6) },
      ])
    case 'robots':
      return compactItems([
        { label: 'Exists', value: boolText(payload.exists) },
        { label: 'Status', value: numberValue(payload.status_code) },
        { label: 'Sitemaps', value: numberValue(payload.sitemap_count) },
        { label: 'Disallow All', value: boolText(payload.global_disallow_all) },
      ])
    case 'security_txt':
      return compactItems([
        { label: 'Exists', value: boolText(payload.exists) },
        { label: 'Path', value: stringValue(payload.path_used) },
        { label: 'Contact', value: stringArray(payload.contact).slice(0, 4) },
        { label: 'Expires', value: stringValue(payload.expires) },
      ])
    case 'page_assets': {
      const icon = asRecord(payload.icon)
      const manifest = asRecord(payload.manifest)
      return compactItems([
        { label: 'Icon', value: boolText(icon.exists) },
        { label: 'Icon Type', value: stringValue(icon.content_type) },
        { label: 'Manifest', value: boolText(manifest.exists) },
        { label: 'App Name', value: stringValue(manifest.name || manifest.short_name) },
        { label: 'Theme', value: stringValue(manifest.theme_color) },
      ])
    }
    case 'port_check':
      return compactItems([
        { label: 'Ports', value: `${formatNumber(payload.ports_checked)}/${formatNumber(payload.ports_configured)}` },
        { label: 'Open', value: formatNumber(payload.open_count) },
        { label: 'Closed', value: formatNumber(payload.closed_count) },
        { label: 'Timeout', value: formatNumber(payload.timeout_count) },
      ])
    case 'waf_canary':
      return compactItems([
        { label: 'Cases', value: `${formatNumber(payload.cases_executed)}/${formatNumber(payload.cases_total)}` },
        { label: 'Blocked', value: formatNumber(payload.blocked_count) },
        { label: 'Matched', value: formatNumber(payload.expected_blocked_matched_count) },
        { label: 'Unexpected Pass', value: formatNumber(payload.unexpected_pass_count) },
      ])
    default:
      return compactItems([
        { label: 'Observed', value: formatTime(envelope.observed_at) },
        { label: 'Duration', value: formatDuration(envelope.duration_ms) },
      ])
  }
}

async function copyText(value: string) {
  if (!import.meta.client || !navigator.clipboard) {
    return
  }
  await navigator.clipboard.writeText(value)
}

function refreshPage() {
  refresh()
}

defineExpose({ refreshPage })
</script>
