<template>
  <section class="rounded-xl p-6 transition-all">
    <div class="relative flex flex-col items-start gap-6 md:flex-row md:items-center">
      <div
        class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-lg bg-orange-100 transition-transform duration-500 hover:scale-[1.05]"
      >
        <img
          :src="logoSrc"
          alt="站点LOGO"
          class="h-full w-full object-contain"
          @error="onImageError"
        />
      </div>

      <div class="w-full flex-1">
        <div class="mb-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-col items-start gap-2 sm:flex-row sm:items-center">
            <h1 class="mr-2 text-2xl font-bold">{{ site.name }}</h1>
            <div class="flex w-auto flex-wrap gap-2">
              <span
                v-for="(tag, index) in tags"
                :key="index"
                class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium whitespace-nowrap"
                :class="tagClass(index)"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-2 self-start sm:self-auto">
            <a
              :href="`https://${site.domain}`"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex transform items-center gap-2 rounded-lg bg-orange-200/60 px-4 py-2 text-sm font-semibold hover:bg-orange-200 duration-500"
            >
              <img src="@/assets/svgs/go.svg" alt="go" class="h-5 w-5 opacity-90" />
              {{ t('site.overview.goto') }}
            </a>
          </div>
        </div>

        <div class="relative mb-3 flex items-center">
          <div class="flex justify-between items-center w-full">
            <div
              class="group/domain relative"
              @mouseenter="openDomainCard"
              @mouseleave="scheduleCloseDomainCard"
            >
              <div class="flex items-center">
                <span
                  class="cursor-pointer font-mono text-gray-500 duration-300 hover:text-orange-400 dark:text-gray-400 dark:hover:text-orange-300"
                  @click="copyToClipboard(site.domain)"
                >
                  {{ site.domain }}
                </span>
                <span class="duration-300 ml-2 text-xs text-gray-400 opacity-0 transition-opacity group-hover/domain:opacity-100 dark:text-gray-500">
                  {{ t('common.copy') }}
                </span>
              </div>

              <transition name="domain-card">
                <div
                  v-if="showDomainCard"
                  class="absolute left-0 top-full z-20 w-[min(22rem,calc(100vw-4rem))] pt-3"
                  @mouseenter="openDomainCard"
                  @mouseleave="scheduleCloseDomainCard"
                >
                  <div class="rounded-xl bg-orange-100/95 p-3 text-sm text-orange-950 backdrop-blur-md dark:bg-stone-900/95 dark:text-orange-50">
                    <div class="mb-2 px-1 text-xs font-medium text-orange-500 dark:text-orange-300">
                      {{ domainListTitle }}
                    </div>
                    <div class="flex flex-col gap-1">
                      <NuxtLink
                        v-for="domain in switchableDomains"
                        :key="domain"
                        :to="domainLink(domain)"
                        class="rounded-lg px-3 py-2 font-mono text-xs text-gray-700 transition duration-200 hover:bg-orange-200/70 hover:text-orange-700 dark:text-gray-200 dark:hover:bg-orange-500/15 dark:hover:text-orange-200"
                        :class="{ 'bg-orange-200/60 text-orange-700 dark:bg-orange-500/20 dark:text-orange-100': domain === site.domain }"
                      >
                        {{ domain }}
                      </NuxtLink>
                    </div>
                  </div>
                </div>
              </transition>
            </div>
            <div>
              <span
                class="flex shrink-0 items-center gap-1 text-xs text-orange-500"
              >
                <strong>{{ t('common.visits') }}: </strong>
                <div>{{ (site.viewCount ?? 0).toLocaleString() }}</div>
              </span>
            </div>
          </div>
          
          

          <transition name="fade">
            <div
              v-if="copied"
              class="absolute -top-6 left-0 rounded bg-gray-800 px-2 py-0.5 text-xs text-white shadow-sm"
            >
              {{ t('common.copied') }}
            </div>
          </transition>
        </div>

        <p class="mb-2 line-clamp-2 text-gray-600 transition-all duration-300 hover:line-clamp-none">
          {{ site.description }}
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/main'

const t = (key: string) => i18n.global.t(key)

interface SiteOverviewProps {
  site: {
    name: string
    icon?: string
    domain: string
    welfare?: boolean
    nsfw?: boolean
    description: string
    viewCount?: number
  }
  enableDomainSwitcher?: boolean
  domainOptions?: string[]
  domainRouteSuffix?: string
  siteId?: string | number
}

const props = defineProps<SiteOverviewProps>()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const logoSrc = computed(() => {
  const icon = props.site.icon || defaultLogo
  return `${logoPrefix ? `${logoPrefix}/` : ''}${icon}`
})

function onImageError(event: Event) {
  const target = event.target as HTMLImageElement
  target.src = `${logoPrefix ? `${logoPrefix}/` : ''}${defaultLogo}`
}

const copied = ref(false)
function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

const showDomainCard = ref(false)
let domainCardCloseTimer: ReturnType<typeof setTimeout> | null = null

const switchableDomains = computed(() => {
  if (!props.enableDomainSwitcher) {
    return []
  }

  const seen = new Set<string>()
  const domains: string[] = []

  for (const domain of props.domainOptions ?? []) {
    const value = domain.trim()
    if (!value || seen.has(value)) {
      continue
    }

    seen.add(value)
    domains.push(value)
  }

  if (props.site.domain && !seen.has(props.site.domain)) {
    domains.unshift(props.site.domain)
  }

  return domains
})

const domainListTitle = computed(() => {
  return i18n.global.locale.value === 'en' ? 'Collected domains' : '采集域名'
})

function openDomainCard() {
  if (!props.enableDomainSwitcher || switchableDomains.value.length <= 1) {
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
  }, 180)
}

function domainLink(domain: string) {
  const siteId = props.siteId ? String(props.siteId) : ''
  const suffix = props.domainRouteSuffix ?? ''
  return `/site/${encodeURIComponent(siteId)}/${encodeURIComponent(domain)}${suffix}`
}

const tags = computed(() => {
  const list: string[] = []
  list.push(props.site.nsfw ? t('site.overview.nsfw') : t('site.overview.sfw'))
  list.push(props.site.welfare ? t('site.overview.welfare') : t('site.overview.non-welfare'))
  return list
})

function tagClass(index: number) {
  const colors = ['bg-orange-100 text-orange-800', 'bg-amber-100 text-amber-800']
  return colors[index % colors.length]
}
</script>

<style scoped>
.domain-card-enter-active,
.domain-card-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}

.domain-card-enter-from,
.domain-card-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
