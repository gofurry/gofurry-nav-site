<template>
  <section
    class="error-page"
    :class="{ 'is-enhanced': enhanced, 'has-entered': enterStarted }"
    aria-labelledby="error-title"
  >
    <div class="error-page__artwork" aria-hidden="true">
      <div
        v-for="art in artworks"
        :key="art.theme"
        class="error-page__artwork-layer"
        :class="{
          'is-active': themeStore.theme === art.theme,
          'is-ready': artReady[art.theme],
        }"
      >
        <img
          :ref="element => setArtworkRef(element, art.theme)"
          class="error-page__image"
          :class="{ 'is-entering': enhanced && artReady[art.theme] && initialTheme === art.theme }"
          :src="art.src"
          alt=""
          aria-hidden="true"
          decoding="async"
          :fetchpriority="themeStore.theme === art.theme ? 'high' : 'low'"
          @load="prepareArtwork(art.theme)"
        >
      </div>
    </div>

    <div class="error-page__content">
      <p class="error-page__code error-page__step" style="--delay: 1350ms">{{ statusCode }}</p>
      <h1 id="error-title" class="error-page__title error-page__step" style="--delay: 1650ms">
        {{ t(`errorPage.${variant}.title`) }}
      </h1>
      <p class="error-page__copy">
        <span class="error-page__line1 error-page__step" style="--delay: 1950ms">
          {{ t(`errorPage.${variant}.line1`) }}
        </span>
        <span class="error-page__line2 error-page__step" style="--delay: 2200ms">
          {{ t(`errorPage.${variant}.line2`) }}
        </span>
      </p>
      <div class="error-page__actions error-page__step" style="--delay: 2550ms">
        <button
          class="gf-button gf-button--primary"
          type="button"
          @click="variant === 'serverError' ? emit('retry') : emit('home')"
        >
          <PhArrowClockwise v-if="variant === 'serverError'" :size="18" aria-hidden="true" />
          <PhHouse v-else :size="18" aria-hidden="true" />
          {{ t(`errorPage.${variant}.${variant === 'serverError' ? 'retry' : 'home'}`) }}
        </button>
        <button
          class="gf-button gf-button--surface"
          type="button"
          @click="variant === 'serverError' ? emit('home') : emit('back')"
        >
          <PhHouse v-if="variant === 'serverError'" :size="18" aria-hidden="true" />
          <PhArrowLeft v-else :size="18" aria-hidden="true" />
          {{ t(`errorPage.${variant}.${variant === 'serverError' ? 'home' : 'back'}`) }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ComponentPublicInstance } from 'vue'
import { PhArrowClockwise, PhArrowLeft, PhHouse } from '@phosphor-icons/vue'
import { useThemeStore } from '@/stores/theme'

defineProps<{
  statusCode: number
  variant: 'notFound' | 'serverError' | 'generic'
}>()
const emit = defineEmits<{ home: []; back: []; retry: [] }>()
const { t } = useI18n()
const themeStore = useThemeStore()

// Hide the staggered content from the first SSR paint, while keeping a
// readable error page when JavaScript is disabled.
useHead({
  noscript: [{ innerHTML: '<style>.error-page__step { opacity: 1 !important; }</style>' }],
})

type Theme = 'light' | 'dark'
const artworks: { theme: Theme; src: string }[] = [
  { theme: 'light', src: 'https://qcdn.go-furry.com/web/404/illustration_001_16_9_light.avif' },
  { theme: 'dark', src: 'https://qcdn.go-furry.com/web/404/illustration_001_16_9_dark.avif' },
]
const artworkElements: Partial<Record<Theme, HTMLImageElement>> = {}
const artReady = reactive<Record<Theme, boolean>>({ light: false, dark: false })
const enhanced = ref(false)
const enterStarted = ref(false)
const initialTheme = ref<Theme>('light')
let fallbackTimer: ReturnType<typeof setTimeout> | undefined
let disposed = false

function setArtworkRef(element: Element | ComponentPublicInstance | null, theme: Theme) {
  if (import.meta.client && element instanceof HTMLImageElement) artworkElements[theme] = element
}

function startEnter() {
  if (!enhanced.value || enterStarted.value || disposed) return
  enterStarted.value = true
  clearTimeout(fallbackTimer)
}

async function prepareArtwork(theme: Theme) {
  const element = artworkElements[theme]
  if (!element?.complete || !element.naturalWidth || artReady[theme]) return
  // Cached images may finish before mount. Decode is an enhancement: a loaded
  // image remains usable in browsers where decode rejects.
  try { await element.decode() } catch { /* retain the successfully loaded image */ }
  if (disposed) return
  artReady[theme] = true
  if (themeStore.theme === theme) startEnter()
}

watch(() => themeStore.theme, theme => {
  if (artReady[theme]) startEnter()
})

onMounted(() => {
  initialTheme.value = themeStore.theme
  enhanced.value = true
  fallbackTimer = setTimeout(startEnter, 900)
  for (const art of artworks) void prepareArtwork(art.theme)
  if (artReady[themeStore.theme]) startEnter()
})

onBeforeUnmount(() => {
  disposed = true
  clearTimeout(fallbackTimer)
})
</script>

<style scoped>
.error-page {
  --error-background: #f5e9d9;
  --error-text: #483728;
  --error-muted: #756453;
  --error-accent: #ac582c;
  position: relative;
  isolation: isolate;
  display: flex;
  flex: 1;
  align-items: center;
  overflow: hidden;
  padding: 64px clamp(28px, 7vw, 120px) 100px;
  background: var(--error-background);
  color: var(--error-text);
}

:global(html.dark .error-page) {
  --error-background: #0d1625;
  --error-text: #f2e6d6;
  --error-muted: #b7b6bb;
  --error-accent: #edac6c;
}

.error-page__artwork,
.error-page__artwork-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.error-page__artwork {
  z-index: -1;
  mask-image: linear-gradient(to right, transparent 36%, #000 68%);
}

.error-page__artwork-layer {
  opacity: 0;
  transition: opacity 300ms ease;
}

.error-page__artwork-layer.is-active { opacity: 1; }
.is-enhanced .error-page__artwork-layer:not(.is-ready) { opacity: 0; }

.error-page__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
}

.error-page__image.is-entering {
  animation: error-art-enter 1800ms cubic-bezier(.65, 0, .35, 1) both;
}

.error-page__content { width: min(100%, 560px); }

.error-page__code {
  margin: 0;
  color: var(--error-accent);
  font-size: clamp(80px, 6.5vw, 96px);
  font-weight: 800;
  line-height: .96;
  letter-spacing: -.04em;
}

.error-page__title {
  margin: 26px 0 0;
  font-size: clamp(34px, 2.8vw, 40px);
  font-weight: 750;
  line-height: 1.3;
  text-wrap: balance;
}

.error-page__copy { margin: 24px 0 0; }
.error-page__copy span { display: block; line-height: 1.7; }
.error-page__line1,
.error-page__line2 { color: var(--error-muted); font-size: 16px; }
.error-page__line2 { margin-top: 3px; }

.error-page__actions { display: flex; gap: 12px; margin-top: 30px; }
.error-page__actions .gf-button {
  min-height: 46px;
  gap: 8px;
  border-radius: 11px;
  padding: 10px 22px;
  font-size: 15px;
  box-shadow: none;
}
.error-page__actions .gf-button:hover { transform: none; }
.error-page__actions .gf-button--primary {
  border-color: #ae5628;
  background: #ae5628;
  color: #fff;
}
.error-page__actions .gf-button--primary:hover { border-color: #b45d2c; background: #b45d2c; }
.error-page__actions .gf-button--surface {
  border-color: color-mix(in srgb, var(--error-text) 24%, transparent);
  background: transparent;
  color: var(--error-text);
}
.error-page__actions .gf-button--surface:hover {
  border-color: color-mix(in srgb, var(--error-text) 40%, transparent);
  background: color-mix(in srgb, var(--error-text) 6%, transparent);
}
.error-page__actions .gf-button:focus-visible {
  outline: 3px solid var(--error-accent);
  outline-offset: 4px;
}

.error-page__step { opacity: 0; }
.has-entered .error-page__step {
  animation: error-content-enter 650ms cubic-bezier(.22, 1, .36, 1) var(--delay) both;
}
.error-page__actions:focus-within {
  opacity: 1 !important;
  transform: none !important;
}

@keyframes error-art-enter {
  from { opacity: .58; }
  to { opacity: 1; }
}
@keyframes error-content-enter {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (min-width: 768px) and (max-width: 1199px) {
  .error-page { padding-inline: 5vw; }
  .error-page__content { width: 50%; max-width: 500px; }
  .error-page__image { object-position: 60% center; }
  .error-page__title { font-size: 34px; }
}

@media (min-width: 768px) and (max-width: 899px) {
  .error-page__content { width: 48%; }
  .error-page__title { font-size: 30px; }
  .error-page__actions { flex-wrap: wrap; }
}

@media (max-width: 767px) {
  .error-page { display: block; padding: 0 0 36px; }
  .error-page__artwork { position: relative; height: 34dvh; mask-image: none; }
  .error-page__artwork::after {
    content: "";
    position: absolute;
    inset: 68% 0 0;
    background: linear-gradient(transparent, var(--error-background));
  }
  .error-page__image { object-position: 72% center; }
  .error-page__content { width: 100%; max-width: 560px; padding: 8px 28px 0; }
  .error-page__code { font-size: 64px; }
  .error-page__title { margin-top: 18px; font-size: 28px; }
  .error-page__copy { margin-top: 18px; }
  .error-page__line1,
  .error-page__line2 { font-size: 15px; }
  .error-page__actions { margin-top: 24px; }
}

@media (max-width: 479px) {
  .error-page__actions { flex-direction: column; }
  .error-page__actions .gf-button { width: 100%; }
}

@media (prefers-reduced-motion: reduce) {
  .error-page .error-page__step,
  .error-page .error-page__image {
    opacity: 1;
    filter: none;
    transform: none;
    animation: none;
  }
  .error-page__artwork-layer,
  .error-page__actions .gf-button { transition: none; }
}
</style>
