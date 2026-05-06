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
            <span
              class="inline-flex items-center gap-1 rounded-full border border-orange-200 bg-white/70 px-3 py-1 text-xs font-medium text-gray-600 backdrop-blur"
            >
              <strong class="text-orange-600">{{ (site.viewCount ?? 0).toLocaleString() }}</strong>
              <span>{{ t('common.visits') }}</span>
            </span>

            <a
              :href="`https://${site.domain}`"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex transform items-center gap-2 rounded-full bg-orange-200 px-4 py-2 text-sm font-semibold transition-transform duration-500 hover:bg-orange-300"
            >
              <img src="@/assets/svgs/go.svg" alt="go" class="h-5 w-5 opacity-90" />
              {{ t('site.overview.goto') }}
            </a>
          </div>
        </div>

        <div class="group relative mb-3 flex items-center">
          <span
            class="cursor-pointer font-mono text-gray-500 hover:text-orange-400"
            @click="copyToClipboard(site.domain)"
          >
            {{ site.domain }}
          </span>
          <span class="ml-2 text-xs text-gray-400 opacity-0 transition-opacity group-hover:opacity-100">
            {{ t('common.copy') }}
          </span>

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
