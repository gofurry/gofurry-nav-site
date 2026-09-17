<template>
  <section class="resource-route gf-modal__section" :aria-labelledby="titleId">
    <h3 :id="titleId" class="gf-modal__label">{{ title }}</h3>
    <fieldset class="resource-route__modes">
      <legend class="sr-only">{{ title }}</legend>
      <label v-for="option in options" :key="option.value" :class="{ selected: mode === option.value }">
        <input type="radio" :name="titleId" :value="option.value" :checked="mode === option.value" @change="emit('update:mode', option.value)">
        <span>{{ option.label }}</span>
      </label>
    </fieldset>
    <dl class="resource-route__summary">
      <div><dt>{{ t('resourceRouting.preferred') }}</dt><dd>{{ preferred }} <span v-if="mode !== 'auto'" class="gf-chip gf-chip--muted">{{ t('resourceRouting.pinned') }}</span></dd></div>
      <div><dt>{{ t('resourceRouting.recommendation') }}</dt><dd>{{ recommendation || t('resourceRouting.never') }}</dd></div>
    </dl>
    <div class="resource-route__measurements" aria-live="polite">
      <div v-for="route in routes" :key="route.label">
        <span>{{ route.label }}</span>
        <span>{{ route.state === 'success' && route.ms !== null ? t('resourceRouting.latency', { ms: Math.round(route.ms) }) : t('resourceRouting.' + route.state) }}<small v-if="route.state === 'stale' && route.ms !== null"> · {{ Math.round(route.ms) }} ms</small></span>
      </div>
    </div>
    <p v-if="warning" class="resource-route__warning">{{ t('resourceRouting.pinnedWarning') }}</p>
    <p v-if="description" class="gf-modal__help">{{ description }}</p>
    <div class="resource-route__footer">
      <span class="gf-modal__help">{{ t('resourceRouting.checked') }}: {{ checkedAt ? new Date(checkedAt).toLocaleString(locale) : t('resourceRouting.never') }}</span>
      <button type="button" class="gf-button gf-button--surface" :disabled="probing || cooldownSeconds > 0" @click="emit('probe')">
        {{ probing ? t('resourceRouting.probing') : cooldownSeconds > 0 ? t('resourceRouting.cooldown', { seconds: cooldownSeconds }) : t('resourceRouting.retest') }}
      </button>
    </div>
    <details v-if="details?.length" class="resource-route__details">
      <summary>{{ t('resourceRouting.details') }}</summary>
      <div v-for="group in details" :key="group.label"><strong>{{ group.label }}</strong><ul><li v-for="host in group.hosts" :key="host">{{ host }}</li></ul></div>
    </details>
  </section>
</template>

<script setup lang="ts">
import { useId } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ProbeState } from '~/utils/resourceRouting'
defineProps<{
  title: string
  mode: string
  options: { value: string; label: string }[]
  preferred: string
  recommendation: string | null
  routes: { label: string; ms: number | null; state: ProbeState }[]
  checkedAt: number | null
  probing: boolean
  cooldownSeconds: number
  warning: boolean
  description?: string
  details?: { label: string; hosts: string[] }[]
}>()
const emit = defineEmits<{ 'update:mode': [value: string]; probe: [] }>()
const { t, locale } = useI18n()
const titleId = useId()
</script>

<style scoped>
.resource-route { display: grid; gap: .8rem; min-width: 0; }
.resource-route__modes { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .4rem; border: 0; padding: 0; margin: 0; }
.resource-route__modes label { display: flex; align-items: center; justify-content: center; gap: .35rem; padding: .65rem .3rem; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); color: var(--gf-text-muted); cursor: pointer; font-size: .8rem; }
.resource-route__modes label:hover { background: var(--gf-surface-hover); }
.resource-route__modes label.selected { background: var(--gf-accent-soft); border-color: var(--gf-accent); color: var(--gf-accent); }
.resource-route__modes input { accent-color: var(--gf-accent); margin: 0; flex-shrink: 0; }
.resource-route__modes label:focus-within, .resource-route button:focus-visible, summary:focus-visible { outline: 2px solid var(--gf-accent); outline-offset: 3px; }
.resource-route__summary { display: grid; gap: .5rem; font-size: .8rem; margin: 0; }
.resource-route__summary > div, .resource-route__measurements > div { display: flex; flex-wrap: wrap; justify-content: space-between; gap: .4rem; }
.resource-route__summary dt { color: var(--gf-text-muted); }
.resource-route__summary dd { margin: 0; color: var(--gf-text-main); }
.resource-route__measurements { display: grid; gap: .5rem; padding: .7rem; border-radius: var(--gf-radius-sm); background: var(--gf-surface); font-size: .8rem; color: var(--gf-text-main); font-variant-numeric: tabular-nums; }
.resource-route__warning { margin: 0; font-size: .8rem; color: var(--gf-text-muted); }
.resource-route__footer { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: .5rem; }
.resource-route__footer button:disabled { background: transparent; color: var(--gf-text-muted); opacity: 1; }
.resource-route__details { font-size: .78rem; color: var(--gf-text-muted); overflow-wrap: anywhere; }
.resource-route__details summary { cursor: pointer; padding: .3rem 0; }
.resource-route__details strong { display: block; margin-top: .6rem; }
.resource-route__details ul { padding-left: 1.2rem; margin: .3rem 0; }
</style>
