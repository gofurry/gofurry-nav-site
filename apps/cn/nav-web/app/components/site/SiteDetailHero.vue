<template>
  <header data-site-hero class="site-detail-hero grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-start gap-x-4 gap-y-3 sm:gap-x-5 lg:grid-cols-[auto_minmax(0,1fr)_auto]">
    <ManagedAssetImage :object-key="site?.icon || undefined" :alt="name" class="site-detail-hero__icon object-contain" />
    <div class="min-w-0">
      <h1 class="site-detail-hero__name break-words">{{ name }}</h1>
      <p class="site-detail-hero__domain mt-1 break-all">{{ domain }}</p>
      <div class="site-detail-hero__meta mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
        <span v-if="site?.country" :aria-label="t('siteDetail.country')">{{ site.country }}</span>
        <span>{{ site?.nsfw === '1' ? 'NSFW' : 'SFW' }}</span>
        <span v-if="site?.welfare === '1'">{{ t('siteDetail.welfare') }}</span>
        <span>{{ t('siteDetail.views') }} {{ viewCount }}</span>
      </div>
    </div>
    <p v-if="site?.info" class="site-detail-hero__description col-span-2 min-w-0 line-clamp-3 break-words sm:col-start-2 sm:col-span-1 sm:line-clamp-none">{{ site.info }}</p>
    <a v-if="visitUrl" :href="visitUrl" target="_blank" rel="noopener noreferrer" class="gf-button gf-button--primary col-span-2 justify-self-start sm:col-start-2 sm:col-span-1 lg:col-start-3 lg:row-start-1">
      {{ t('siteDetail.visit') }} <span aria-hidden="true">↗</span>
    </a>
  </header>
</template>

<script setup lang="ts">
import ManagedAssetImage from '@/components/common/ManagedAssetImage.vue'
import type { SiteInfo } from '~/types/nav'
defineProps<{ site: SiteInfo | null; name: string; domain: string; viewCount: number; visitUrl: string }>()
const { t } = useI18n()
</script>
