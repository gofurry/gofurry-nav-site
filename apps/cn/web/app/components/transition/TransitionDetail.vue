<template>
  <article class="mx-auto w-full max-w-4xl flex-1 px-4 py-8 sm:px-8" data-transition-detail>
    <div class="gf-card p-6 sm:p-8">
      <div class="flex flex-col items-start gap-6 sm:flex-row">
        <ManagedAssetImage v-if="kind === 'site'" :object-key="detail?.image || undefined" :alt="detail?.name || ''" class="h-24 w-24 shrink-0 object-contain" />
        <SteamAssetImage v-else-if="detail?.image" :src="detail.image" :alt="detail.name" class="w-48 shrink-0 object-contain" />
        <div class="min-w-0 space-y-4">
          <h1 class="site-detail-hero__name break-words">{{ detail?.name }}</h1>
          <p class="whitespace-pre-line break-words">{{ detail?.info }}</p>
          <a v-if="detail?.url" :href="detail.url" target="_blank" rel="noopener noreferrer" class="gf-button gf-button--primary gf-button--stationary">
            {{ kind === 'site' ? t('siteDetail.visit') : t('game.detail.officialWebsite') }}
          </a>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import ManagedAssetImage from '~/components/common/ManagedAssetImage.vue'
import SteamAssetImage from '~/components/common/SteamAssetImage.vue'
import { getNavSiteDirectory } from '~/services/nav'
import { getGameBaseInfo } from '~/services/game'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { buildGameDetailSeo, buildSiteDetailSeo } from '~/utils/seo'

const props = defineProps<{ kind: 'site' | 'game' }>()
const route = useRoute()
const { t, locale } = useI18n()
const id = computed(() => String(route.params.id))
const lang = computed(() => locale.value === 'en' ? 'en' : 'zh')
const { data: detail, error } = await useAsyncData(
  () => `transition-detail:${props.kind}:${id.value}:${lang.value}`,
  async () => {
    if (props.kind === 'site') {
      // The existing directory contains base profiles only; the normal Detail
      // endpoint also loads monitoring. Keep that entire runtime unmounted.
      const sites = await getNavSiteDirectory(lang.value)
      if (!sites.length) throw createError({ statusCode: 503 })
      const site = sites.find(item => String(item.id) === id.value)
      if (!site) throw createError({ statusCode: 404 })
      return { name: site.name, info: site.info, image: site.icon, url: externalUrl(site.domain) }
    }
    const game = await getGameBaseInfo(id.value, lang.value)
    // The existing base API bundles optional fields. Only base-profile fields
    // enter Nuxt's SSR payload; no news/review/recommendation/Insights component mounts.
    return { name: game.name, info: game.info || game.short_description, image: game.cover,
      url: externalUrl(game.website) || (game.appid ? `https://store.steampowered.com/app/${game.appid}/` : '') }
  },
)
if (error.value) {
  const statusCode = authoritativePageStatus(error.value, props.kind)
  throw createError({ statusCode, statusMessage: statusCode === 404 ? 'Not found' : 'Service temporarily unavailable' })
}
const seo = computed(() => (props.kind === 'site' ? buildSiteDetailSeo : buildGameDetailSeo)({
  name: detail.value?.name, description: detail.value?.info, locale: locale.value,
}))
useSeoMeta({ title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description })

function externalUrl(value: string | null | undefined) {
  if (!value) return ''
  try {
    const url = new URL(value.includes('://') ? value : `https://${value}`)
    return ['http:', 'https:'].includes(url.protocol) ? url.href : ''
  } catch { return '' }
}
</script>
