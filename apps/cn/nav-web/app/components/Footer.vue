<template>
  <footer class="gf-footer">
    <div class="mx-auto grid w-full max-w-[1700px] gap-10 px-4 py-10 sm:px-6 md:grid-cols-3">
      <div class="space-y-8">
        <section class="space-y-3">
          <h3 class="gf-footer__section-title flex items-center gap-2">
            <PhCompass :size="16" weight="regular" aria-hidden="true" />
            {{ t('footer.sections.sitemap') }}
          </h3>
          <div class="flex flex-wrap gap-2">
            <a
                :href="sitemapUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.sitemapXml') }}
            </a>
            <a
                :href="llmsUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.llmsTxt') }}
            </a>
            <a
                :href="securityTxtUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.securityTxt') }}
            </a>
          </div>
        </section>

        <section class="space-y-3">
          <h3 class="gf-footer__section-title flex items-center gap-2">
            <PhBracketsCurly :size="16" weight="regular" aria-hidden="true" />
            {{ t('footer.sections.openPlatform') }}
          </h3>
          <div class="flex flex-wrap gap-2">
            <a
                :href="uptimeUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.uptime') }}
            </a>
            <a
                :href="navMonitorUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.api') }}
            </a>
            <a
                href="https://op.go-furry.com"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-footer__link"
            >
              {{ t('footer.links.opsAdmin') }}
            </a>
          </div>
        </section>
      </div>

      <div class="space-y-8">
        <section class="space-y-3">
          <h3 class="gf-footer__section-title flex items-center gap-2">
            <PhChatCircleDots :size="16" weight="regular" aria-hidden="true" />
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
                  class="gf-footer__social-icon h-6 w-6"
                  :data-brand="item.key"
              />
            </a>
          </div>
        </section>

        <section class="space-y-3">
          <h3 class="gf-footer__section-title flex items-center gap-2">
            <PhInfo :size="16" weight="regular" aria-hidden="true" />
            {{ t('footer.sections.about') }}
          </h3>
          <div class="flex gap-2">
            <NuxtLink :to="localePath('/about')" class="gf-footer__link">
              {{ t('sidebar.about') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/updates')" class="gf-footer__link">
              {{ t('navHeader.update') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/terms')" class="gf-footer__link">
              {{ t('footer.links.terms') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/privacy')" class="gf-footer__link">
              {{ t('footer.links.privacy') }}
            </NuxtLink>
          </div>
        </section>
      </div>

      <div class="gf-footer__meta flex flex-col justify-end gap-3">
        <div>{{ currentYear }} GoFurry {{ t('footer.rights') }}</div>
        <div>{{ t('footer.license') }}</div>
        <a
            href="https://beian.miit.gov.cn/"
            target="_blank"
            rel="noopener noreferrer"
            class="gf-footer__meta-link"
        >
          {{ t('footer.icp') }}
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PhBracketsCurly, PhChatCircleDots, PhCompass, PhInfo } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import bilibiliIcon from '@/assets/icons/bilibili.svg'
import weiboIcon from '@/assets/icons/weibo.svg'
import githubIcon from '@/assets/icons/github.svg'
import twitterIcon from '@/assets/icons/twitter.svg'

const { t } = useI18n()
const localePath = useLocalePath()

const currentYear = new Date().getFullYear()
const sitemapUrl = '/sitemap.xml'
const llmsUrl = '/llms.txt'
const securityTxtUrl = '/.well-known/security.txt'
const config = useRuntimeConfig()
const uptimeUrl = computed(() => config.public.uptimeUrl)
const navMonitorUrl = computed(() => config.public.navMonitorUrl)

interface FeedbackLink {
  key: string
  href: string
  labelKey: string
  icon: string
}

const feedbackLinks: FeedbackLink[] = [
  {
    key: 'bilibili',
    href: 'https://space.bilibili.com/37124259',
    labelKey: 'footer.links.bilibili',
    icon: bilibiliIcon,
  },
  {
    key: 'weibo',
    href: 'https://www.weibo.com/u/6233129221',
    labelKey: 'footer.links.weibo',
    icon: weiboIcon,
  },
  {
    key: 'github',
    href: 'https://github.com/gofurry',
    labelKey: 'footer.links.github',
    icon: githubIcon,
  },
  {
    key: 'twitter',
    href: 'https://x.com/InLoveWithCharr',
    labelKey: 'footer.links.twitter',
    icon: twitterIcon,
  },
]
</script>
