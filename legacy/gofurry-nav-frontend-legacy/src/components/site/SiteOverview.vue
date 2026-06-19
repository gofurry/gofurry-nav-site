<template>
  <section
      class="rounded-xl p-6 transition-all"
  >
    <div class="flex flex-col md:flex-row items-start md:items-center gap-6 relative">
      <!-- LOGO -->
      <div
          class="w-20 h-20 rounded-lg bg-orange-100 flex items-center justify-center
           overflow-hidden hover:scale-[1.05] transform transition-transform duration-500"
      >
        <img
            :src="logoSrc"
            alt="站点LOGO"
            class="w-full h-full object-contain"
            @error="onImageError"
        />
      </div>

      <!-- 文字 -->
      <div class="flex-1 w-full">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-2">
          <!-- 标题与标签 -->
          <div class="flex flex-col sm:flex-row sm:items-center gap-2 items-start">
            <h1 class="text-2xl font-bold mr-2">{{ site.name }}</h1>
            <div class="flex flex-wrap gap-2 w-auto">
              <span
                  v-for="(tag, index) in tags"
                  :key="index"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium whitespace-nowrap"
                  :class="tagClass(index)"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <a
              :href="`https://${site.domain}`"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-full
               bg-orange-200 hover:bg-orange-300
               transition-transform duration-500 self-start sm:self-auto transform"
          >
            <img src="@/assets/svgs/go.svg" alt="go" class="w-5 h-5 opacity-90" />
            {{ t('site.overview.goto') }}
          </a>

        </div>

        <!-- 域名可复制 -->
        <div class="flex items-center mb-3 group relative">
          <span
              class="text-gray-500 font-mono cursor-pointer hover:text-orange-400"
              @click="copyToClipboard(site.domain)"
          >
            {{ site.domain }}
          </span>
          <span
              class="ml-2 text-xs text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            {{ t('common.copy') }}
          </span>

          <!-- 浮动提示 -->
          <transition name="fade">
            <div
                v-if="copied"
                class="absolute left-0 -top-6 bg-gray-800 text-white text-xs px-2 py-0.5 rounded shadow-sm"
            >
              {{ t('common.copied') }}
            </div>
          </transition>
        </div>

        <!-- 简介 -->
        <p
            class="text-gray-600 mb-2 line-clamp-2 hover:line-clamp-none transition-all duration-300"
        >
          {{ site.description }}
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import {i18n} from "@/main.ts";

const t = (key: string) => i18n.global.t(key)

interface SiteOverviewProps {
  site: {
    name: string
    icon?: string
    domain: string
    welfare?: boolean
    nsfw?: boolean
    description: string
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

// 修改tags计算属性，使用i18n翻译标签
const tags = computed(() => {
  const list: string[] = []
  list.push(props.site.nsfw ? t('site.overview.nsfw') : t('site.overview.sfw'))
  list.push(props.site.welfare ? t('site.overview.welfare') : t('site.overview.non-welfare'))
  return list
})

function tagClass(index: number) {
  const colors = [
    'bg-orange-100 text-orange-800',
    'bg-amber-100 text-amber-800'
  ]
  return colors[index % colors.length]
}
</script>

<style scoped>

</style>
