<template>
  <section class="detail-hero">
    <div class="relative z-10 flex min-w-0 flex-col gap-5 md:flex-row md:items-center">
      <div class="flex shrink-0 justify-center md:self-center">
        <div class="logo-shell">
          <img
            v-if="icon"
            :src="logoUrl(icon)"
            :alt="siteName"
            class="h-20 w-20 rounded-lg object-contain"
          >
          <div v-else class="flex h-20 w-20 items-center justify-center rounded-lg text-2xl font-bold text-slate-700">
            GF
          </div>
        </div>
      </div>

      <div class="min-w-0 flex-1">
        <div class="flex min-w-0 flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0 flex-1">
            <div class="hero-title-row">
              <h1 class="mr-2 break-words text-xl font-black tracking-normal text-slate-950 md:text-2xl">
                {{ siteName }}
              </h1>
              <div class="flex flex-wrap items-center gap-2.5">
                <span
                  v-for="badge in badges"
                  :key="badge.label"
                  class="detail-pill"
                  :class="badge.class"
                >
                  {{ badge.label }}
                </span>
              </div>
            </div>

            <div class="hero-meta-row">
              <div
                class="group/domain relative w-fit"
                @pointerenter="openDomainCard"
                @pointerleave="scheduleCloseDomainCard"
                @focusin="openDomainCard"
                @focusout="scheduleCloseDomainCard"
              >
                <button
                  type="button"
                  class="flex items-center font-mono text-sm text-slate-500 transition-colors duration-500 hover:text-orange-500"
                  @click="copyToClipboard(domain)"
                >
                  <span>{{ domain }}</span>
                  <span class="ml-2 text-xs text-slate-400 opacity-0 transition-opacity duration-500 group-hover/domain:opacity-100">
                    {{ t('common.copy') }}
                  </span>
                </button>

                <transition name="domain-card">
                  <div
                    v-show="showDomainCard"
                    class="absolute left-0 top-full z-30 w-[min(22rem,calc(100vw-3rem))] pt-3"
                    @pointerenter="openDomainCard"
                    @pointerleave="scheduleCloseDomainCard"
                  >
                    <div class="absolute left-0 top-0 h-3 w-full" />
                    <div class="domain-popover">
                      <div class="mb-2 px-1 text-xs font-semibold text-orange-500">
                        {{ label('采集域名', 'Collected domains') }}
                      </div>
                      <div class="domain-list-scroll flex max-h-72 flex-col gap-1 overflow-y-auto pr-1">
                        <NuxtLink
                          v-for="item in switchableDomains"
                          :key="item"
                          :to="domainLink(item)"
                          class="rounded-lg px-3 py-2 font-mono text-xs text-slate-700 transition-colors duration-500 hover:bg-orange-100/80 hover:text-orange-700"
                          :class="{ 'bg-orange-100/80 text-orange-700': item === domain }"
                        >
                          {{ item }}
                        </NuxtLink>
                      </div>
                    </div>
                  </div>
                </transition>

                <transition name="fade">
                  <div
                    v-if="copied"
                    class="absolute -top-7 left-0 rounded bg-slate-900 px-2 py-0.5 text-xs text-white"
                  >
                    {{ t('common.copied') }}
                  </div>
                </transition>
              </div>
              <div class="hero-mobile-views">
                {{ label('浏览量', 'Views') }}: {{ viewCount ?? 0 }}
              </div>
            </div>
          </div>

          <div class="hero-actions">
            <a
              v-if="visitUrl"
              :href="visitUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="visit-button duration-500"
            >
              <img src="@/assets/svgs/go.svg" alt="" class="h-5 w-5 opacity-90">
              {{ label('访问网站', 'Visit site') }}
            </a>
            <div class="hidden text-xs text-orange-500 lg:block">
              {{ label('浏览量', 'Views') }}: {{ viewCount ?? 0 }}
            </div>
          </div>
        </div>

        <p class="max-w-6xl text-sm leading-7 text-slate-700 md:text-base">
          {{ info || '-' }}
        </p>

        <div v-if="keywords.length" class="flex flex-wrap gap-2">
          <span
            v-for="keyword in keywords"
            :key="keyword"
            class="keyword-chip"
          >
            {{ keyword }}
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { i18n } from '@/main'
import type { SiteHeroBadge } from './detailTypes'

const props = defineProps<{
  badges: SiteHeroBadge[]
  domain: string
  icon?: string
  info?: string
  keywords: string[]
  logoPrefix: string
  siteId: string | number
  siteName: string
  switchableDomains: string[]
  viewCount?: number
  visitUrl: string
}>()

const copied = ref(false)
const showDomainCard = ref(false)
let domainCardCloseTimer: ReturnType<typeof setTimeout> | null = null

function copyToClipboard(text: string) {
  if (!text) {
    return
  }

  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    void navigator.clipboard.writeText(text)
  }
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 1800)
}

function openDomainCard() {
  if (props.switchableDomains.length <= 1) {
    return
  }

  if (domainCardCloseTimer) {
    clearTimeout(domainCardCloseTimer)
    domainCardCloseTimer = null
  }

  showDomainCard.value = true
}

function scheduleCloseDomainCard() {
  if (domainCardCloseTimer) {
    clearTimeout(domainCardCloseTimer)
  }

  domainCardCloseTimer = setTimeout(() => {
    showDomainCard.value = false
  }, 320)
}

function domainLink(domain: string) {
  return `/site/${encodeURIComponent(String(props.siteId))}/${encodeURIComponent(domain)}`
}

function logoUrl(icon: string) {
  if (/^https?:\/\//i.test(icon)) {
    return icon
  }
  return `${props.logoPrefix}${icon}`
}

function t(key: string) {
  return i18n.global.t(key)
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.detail-hero {
  position: relative;
  overflow: visible;
  border-radius: 0.5rem;
  background:
    radial-gradient(circle at 18% 24%, rgba(251, 140, 47, 0.20), transparent 26%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.74), rgba(255, 242, 219, 0.62));
  padding: clamp(1rem, 2vw, 1.6rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18);
}

.logo-shell {
  border-radius: 0.5rem;
  background: #ffedd5;
  padding: 0.45rem;
}

.hero-title-row {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.55rem;
}

.detail-pill {
  border-radius: 999px;
  padding: 0.25rem 0.7rem;
  font-size: 0.75rem;
  font-weight: 400;
  white-space: nowrap;
}

.detail-pill-edge {
  background: rgba(255, 237, 213, 0.78);
  color: #9a3412;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.20);
}

.detail-pill-sfw {
  background: #fed7aa;
  color: #c2410c;
}

.detail-pill-risk {
  background: rgba(254, 226, 226, 0.80);
  color: #b91c1c;
}

.detail-pill-welfare {
  background: #fde68a;
  color: #b45309;
}

.keyword-chip {
  border-radius: 999px;
  background: #fed7aa;
  padding: 0.25rem 0.75rem;
  color: #9a3412;
  font-size: 0.75rem;
  font-weight: 400;
}

.hero-meta-row {
  position: relative;
  margin-top: 0.5rem;
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.hero-mobile-views {
  margin-left: auto;
  flex: 0 0 auto;
  color: #ea580c;
  font-size: 0.75rem;
  font-weight: 600;
  text-align: right;
  white-space: nowrap;
}

.hero-actions {
  display: flex;
  flex-shrink: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.6rem;
}

.domain-popover {
  border-radius: 1rem;
  background: rgba(255, 247, 237, 0.96);
  padding: 0.75rem;
  color: #1f2937;
  backdrop-filter: blur(14px);
}

.domain-card-enter-active,
.domain-card-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}

.domain-card-enter-from,
.domain-card-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 160ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.domain-list-scroll {
  scrollbar-width: thin;
}

.visit-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  background: rgba(254, 215, 170, 0.62);
  padding: 0.55rem 1rem;
  color: #111827;
  font-size: 0.875rem;
  font-weight: 700;
}

.visit-button:hover {
  background: rgba(254, 215, 170, 0.88);
  color: #111827;
}

@media (min-width: 640px) {
  .hero-title-row {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.65rem;
  }
}

@media (min-width: 1024px) {
  .hero-meta-row {
    width: fit-content;
    justify-content: flex-start;
  }

  .hero-mobile-views {
    display: none;
  }

  .hero-actions {
    align-items: flex-end;
    margin-left: auto;
    padding-left: 2rem;
  }
}
</style>
