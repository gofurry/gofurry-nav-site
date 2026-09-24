<template>
  <div data-resource-routing>
    <p class="gf-modal__help">{{ t('resourceRouting.help') }}</p>
    <ResourceRouteSection
      :title="t('resourceRouting.gofurry')" :mode="assetMode" :options="assetOptions"
      :preferred="assetLabel(resolveAssetPreferred(assetMode, asset.recommendation.value))"
      :recommendation="asset.recommendation.value ? assetLabel(asset.recommendation.value) : null"
      :routes="assetRoutes" :checked-at="asset.diagnostics.value?.checkedAt || null"
      :probing="asset.probing.value" :cooldown-seconds="cooldown(asset.lastManualAt.value)"
      :warning="assetMode !== 'auto' && failed(assetRoutes.find(route => route.value === assetMode)?.state)"
      @update:mode="assetMode = normalizeAssetMode($event)" @probe="asset.probe(true)"
    />
    <ResourceRouteSection
      :title="t('resourceRouting.steam')" :mode="steamMode" :options="steamOptions"
      :preferred="steamLabel(resolveSteamPreferred(steamMode, steam.recommendation.value, locale))"
      :recommendation="steam.recommendation.value ? steamLabel(steam.recommendation.value) : null"
      :routes="steamRoutes" :checked-at="steam.diagnostics.value?.checkedAt || null"
      :probing="steam.probing.value" :cooldown-seconds="cooldown(steam.lastManualAt.value)"
      :warning="steamMode !== 'auto' && failed(steamRoutes.find(route => route.value === steamMode)?.state)"
      :description="t('resourceRouting.steamProbeHint')" :details="steamDetails"
      @update:mode="steamMode = normalizeSteamMode($event)" @probe="steam.probe(true)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ResourceRouteSection from './ResourceRouteSection.vue'
import { normalizeAssetMode, resolveAssetPreferred, type AssetCDN } from '~/utils/managedAssets'
import { normalizeSteamMode, resolveSteamPreferred, STEAM_SHARED_CDN_GROUP_PREFIXES, type SteamSharedCdnGroup } from '~/utils/steamAssets'
import { isFresh, MANUAL_PROBE_COOLDOWN_MS, type ProbeState } from '~/utils/resourceRouting'

const { t, locale } = useI18n()
const { $assetCDN: asset, $steamAssetRoute: steam } = useNuxtApp()
const assetMode = ref(asset.mode.value)
const steamMode = ref(steam.mode.value)
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => { timer = setInterval(() => { now.value = Date.now() }, 1000) })
onUnmounted(() => clearInterval(timer))
const cooldown = (last: number) => Math.max(0, Math.ceil((last + MANUAL_PROBE_COOLDOWN_MS - now.value) / 1000))
const assetLabel = (provider: AssetCDN) => t('resourceRouting.' + provider)
const steamLabel = (group: SteamSharedCdnGroup) => t('resourceRouting.' + group)
const assetOptions = computed(() => [{ value: 'auto', label: t('resourceRouting.auto') }, ...(['primary', 'mirror'] as const).map(value => ({ value, label: assetLabel(value) }))])
const steamOptions = computed(() => [{ value: 'auto', label: t('resourceRouting.auto') }, ...(['china', 'global'] as const).map(value => ({ value, label: steamLabel(value) }))])
const state = (probing: boolean, checkedAt?: number, status?: ProbeState, stale = false): ProbeState =>
  probing ? 'probing' : checkedAt === undefined ? 'never' : stale || !isFresh(checkedAt, now.value) ? 'stale' : status || 'never'
const failed = (value?: ProbeState) => value === 'failed' || value === 'timeout'
const assetRoutes = computed(() => (['primary', 'mirror'] as const).map(value => ({
  value, label: assetLabel(value), ms: asset.diagnostics.value?.[value === 'primary' ? 'primaryMs' : 'mirrorMs'] ?? null,
  state: state(asset.probing.value, asset.diagnostics.value?.checkedAt, asset.diagnostics.value?.[value === 'primary' ? 'primaryState' : 'mirrorState'], asset.diagnostics.value?.stale),
})))
const steamRoutes = computed(() => (['china', 'global'] as const).map(value => ({
  value, label: steamLabel(value), ms: steam.diagnostics.value?.[value].ms ?? null,
  state: state(steam.probing.value, steam.diagnostics.value?.checkedAt, steam.diagnostics.value?.[value].state, steam.diagnostics.value?.stale),
})))
const steamDetails = computed(() => (['china', 'global'] as const).map(group => ({
  label: steamLabel(group), hosts: STEAM_SHARED_CDN_GROUP_PREFIXES[group].map(prefix => new URL(prefix).host),
})))
defineExpose({ save() { asset.saveMode(assetMode.value); steam.saveMode(steamMode.value) } })
</script>
