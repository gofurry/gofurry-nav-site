<template>
  <footer class="bg-slate-800 text-slate-300">
    <div class="mx-auto grid w-full max-w-[1700px] gap-10 px-4 py-10 sm:px-6 md:grid-cols-3">
      <div class="space-y-8">
        <section class="space-y-3">
          <h3 class="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">
            <img :src="compassIcon" alt="sitemap" class="h-4 w-4 opacity-80" />
            {{ t('footer.sections.sitemap') }}
          </h3>
          <a
              :href="sitemapUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="footer-link"
          >
            {{ t('footer.links.sitemapXml') }}
          </a>
        </section>

        <section class="space-y-3">
          <h3 class="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">
            <img :src="apiIcon" alt="open platform" class="h-4 w-4 opacity-80" />
            {{ t('footer.sections.openPlatform') }}
          </h3>
          <div class="flex gap-2">
            <component
                :is="openPlatformApiLink.external ? 'a' : 'NuxtLink'"
                v-bind="openPlatformApiLink.external
                  ? { href: openPlatformApiLink.href, target: '_blank', rel: 'noopener noreferrer' }
                  : { to: openPlatformApiLink.to }"
                class="footer-link"
            >
              {{ t('footer.links.api') }}
            </component>
            <NuxtLink to="/panel" class="footer-link">
              {{ t('sidebar.panel') }}
            </NuxtLink>
            <a
                href="https://op.go-furry.com"
                target="_blank"
                rel="noopener noreferrer"
                class="footer-link"
            >
              {{ t('footer.links.opsAdmin') }}
            </a>
          </div>
        </section>
      </div>

      <div class="space-y-8">
        <section class="space-y-3">
          <h3 class="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">
            <img :src="siteIcon" alt="open platform" class="h-4 w-4 opacity-80" />
            {{ t('footer.sections.feedback') }}
          </h3>
          <div class="flex items-center gap-4">
            <a
                v-for="item in feedbackLinks"
                :key="item.key"
                :href="item.href"
                target="_blank"
                rel="noopener noreferrer"
                :aria-label="t(item.labelKey)"
            >
              <img
                  :src="item.icon"
                  :alt="t(item.labelKey)"
                  class="h-6 w-6 cursor-pointer opacity-90 transition-transform hover:scale-125 hover:opacity-100"
                  :class="item.hoverClass"
              />
            </a>
          </div>
        </section>

        <section class="space-y-3">
          <h3 class="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">
            <img :src="featherIcon" alt="about" class="h-4 w-4 opacity-80" />
            {{ t('footer.sections.about') }}
          </h3>
          <div class="flex gap-2">
            <NuxtLink to="/about" class="footer-link">
              {{ t('sidebar.about') }}
            </NuxtLink>
            <NuxtLink to="/updates" class="footer-link">
              {{ t('navHeader.update') }}
            </NuxtLink>
            <a
                href="https://www.deepfurry.com"
                target="_blank"
                rel="noopener noreferrer"
                class="footer-link"
            >
              DeepFurry
            </a>
          </div>
        </section>
      </div>

      <div class="flex flex-col justify-end gap-3 text-sm text-slate-400">
        <div>{{ currentYear }} gofurry {{ t('footer.rights') }}</div>
        <div>{{ t('footer.license') }}</div>
        <a
            href="https://beian.miit.gov.cn/"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-white"
        >
          {{ t('footer.icp') }}
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import siteIcon from '@/assets/svgs/site.svg'
import apiIcon from '@/assets/svgs/api.svg'
import compassIcon from '@/assets/svgs/compass.svg'
import featherIcon from '@/assets/svgs/feather.svg'
import bilibiliIcon from '@/assets/icons/bilibili.svg'
import weiboIcon from '@/assets/icons/weibo.svg'
import githubIcon from '@/assets/icons/github.svg'
import twitterIcon from '@/assets/icons/twitter.svg'

const { t } = useI18n()

const currentYear = new Date().getFullYear()
const sitemapUrl = '/sitemap.xml'
const openPlatformApiLink = computed(() => (
  import.meta.env.PROD
    ? { external: true, href: 'https://open.go-furry.com' }
    : { external: false, to: '/updates' }
))

const feedbackLinks = [
  {
    key: 'bilibili',
    href: 'https://space.bilibili.com/37124259',
    labelKey: 'footer.links.bilibili',
    icon: bilibiliIcon,
    hoverClass: 'hover:drop-shadow-[0_0_6px_rgb(240,128,128)]'
  },
  {
    key: 'weibo',
    href: 'https://www.weibo.com/u/6233129221',
    labelKey: 'footer.links.weibo',
    icon: weiboIcon,
    hoverClass: 'hover:drop-shadow-[0_0_6px_rgb(255,69,0)]'
  },
  {
    key: 'github',
    href: 'https://github.com/gofurry',
    labelKey: 'footer.links.github',
    icon: githubIcon,
    hoverClass: 'hover:drop-shadow-[0_0_6px_rgb(56,189,248)]'
  },
  {
    key: 'twitter',
    href: 'https://x.com/InLoveWithCharr',
    labelKey: 'footer.links.twitter',
    icon: twitterIcon,
    hoverClass: 'hover:drop-shadow-[0_0_6px_rgb(29,155,240)]'
  },
]
</script>

<style scoped>
.footer-link {
  color: rgb(203 213 225);
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: white;
}
</style>
