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
      <p class="error-page__code error-page__step" style="--gf-error-step-delay: 1350ms">{{ statusCode }}</p>
      <h1 id="error-title" class="error-page__title error-page__step" style="--gf-error-step-delay: 1650ms">
        {{ t(`errorPage.${variant}.title`) }}
      </h1>
      <p class="error-page__copy">
        <span class="error-page__line1 error-page__step" style="--gf-error-step-delay: 1950ms">
          {{ t(`errorPage.${variant}.line1`) }}
        </span>
        <span class="error-page__line2 error-page__step" style="--gf-error-step-delay: 2200ms">
          {{ t(`errorPage.${variant}.line2`) }}
        </span>
      </p>
      <div class="error-page__actions error-page__step" style="--gf-error-step-delay: 2550ms">
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

// Keep the enhanced experience hidden from the first SSR paint, while making
// the complete error page visible when JavaScript is disabled.
useHead({
  noscript: [{
    innerHTML: '<style>.error-page .error-page__step,.error-page .error-page__artwork-layer.is-active{opacity:1}</style>',
  }],
})

type Theme = 'light' | 'dark'
const artworks: { theme: Theme; src: string }[] = [
  { theme: 'light', src: '/web/404/illustration_001_16_9_light.avif' },
  { theme: 'dark', src: '/web/404/illustration_001_16_9_dark.avif' },
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
  position: relative;
  isolation: isolate;
  display: flex;
  flex: 1;
  align-items: center;
  overflow: hidden;
  padding: 64px clamp(28px, 7vw, 120px) 100px;
}

.error-page__artwork,
.error-page__artwork-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.error-page__artwork { z-index: -1; }
.error-page__image { width: 100%; height: 100%; }
.error-page__content { width: min(100%, 560px); }
.error-page__code { margin: 0; }
.error-page__title { margin: 26px 0 0; }
.error-page__copy { margin: 24px 0 0; }
.error-page__copy span { display: block; }
.error-page__line2 { margin-top: 3px; }
.error-page__actions { display: flex; gap: 12px; margin-top: 30px; }

@media (min-width: 768px) and (max-width: 1199px) {
  .error-page { padding-inline: 5vw; }
  .error-page__content { width: 50%; max-width: 500px; }
}

@media (min-width: 768px) and (max-width: 899px) {
  .error-page__content { width: 48%; }
  .error-page__actions { flex-wrap: wrap; }
}

@media (max-width: 767px) {
  .error-page { display: block; padding: 0 0 36px; }
  .error-page__artwork { position: relative; height: 34dvh; }
  .error-page__content { width: 100%; max-width: 560px; padding: 8px 28px 0; }
  .error-page__title { margin-top: 18px; }
  .error-page__copy { margin-top: 18px; }
  .error-page__actions { margin-top: 24px; }
}

@media (max-width: 479px) {
  .error-page__actions { flex-direction: column; }
  .error-page__actions .gf-button { width: 100%; }
}
</style>
