<template>
  <button type="button" class="pattern-option" :data-pattern-id="pattern.id" :aria-pressed="selected" :title="name" @click="$emit('select')">
    <span class="pattern-option__preview"><span :style="maskStyle" /></span>
    <span class="pattern-option__name">{{ name }}</span>
    <PhCheck v-if="selected" class="pattern-option__check" :size="13" weight="bold" aria-hidden="true" />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { PhCheck } from '@phosphor-icons/vue'
import { useThemeStore } from '~/stores/theme'
import type { BackgroundPattern } from '~/types/nav'
import { DEFAULT_PATTERN_URL } from '~/utils/backgroundPreferences'

const props = defineProps<{ pattern: BackgroundPattern; selected: boolean }>()
defineEmits<{ select: [] }>()
const { locale } = useI18n()
const theme = useThemeStore()
const name = computed(() => locale.value === 'en' ? props.pattern.name_en : props.pattern.name)
const asset = useManagedAsset(() => props.pattern.object_key, DEFAULT_PATTERN_URL, true)
const maskStyle = computed(() => ({
  backgroundColor: theme.theme === 'dark' ? props.pattern.dark_color : props.pattern.light_color,
  // Thumbnails stay legible even when the published default is very faint.
  opacity: Math.max(.3, theme.theme === 'dark' ? props.pattern.dark_opacity : props.pattern.light_opacity),
  maskImage: `url("${asset.src.value}")`, WebkitMaskImage: `url("${asset.src.value}")`,
}))
</script>

<style scoped>
.pattern-option { position: relative; flex: 0 0 calc((100% - 1.2rem) / 3); min-width: 90px; padding: 0; overflow: hidden; border: 1px solid var(--gf-border); border-radius: var(--gf-radius-sm); background: transparent; color: var(--gf-text-main); cursor: pointer; scroll-snap-align: start; }
.pattern-option:hover { border-color: var(--gf-accent); background: var(--gf-surface-hover); }
.pattern-option[aria-pressed='true'] { border-color: var(--gf-accent); box-shadow: 0 0 0 1px var(--gf-accent); }
.pattern-option:focus-visible { outline: 2px solid var(--gf-accent); outline-offset: 2px; }
.pattern-option__preview { position: relative; display: block; height: 70px; background: var(--gf-page-background); }
.pattern-option__preview > span { position: absolute; inset: 0; mask-size: 64px; -webkit-mask-size: 64px; mask-repeat: repeat; -webkit-mask-repeat: repeat; }
.pattern-option__name { display: block; overflow: hidden; padding: .45rem .4rem; font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.pattern-option__check { position: absolute; top: 5px; right: 5px; padding: 3px; width: 20px; height: 20px; border-radius: 50%; background: var(--gf-accent); color: var(--gf-modal-bg); }
@media (max-width: 400px) { .pattern-option { flex-basis: calc((100% - .6rem) / 2); } }
</style>
