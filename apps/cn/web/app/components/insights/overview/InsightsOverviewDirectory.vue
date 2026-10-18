<template>
  <section class="overview-explore" aria-labelledby="overview-explore-title" data-overview-explore>
    <h2 id="overview-explore-title" class="overview-heading overview-explore__title mt-0 mb-6">{{ $t('insights.overviewDirectory.title') }}</h2>
    <div class="grid grid-cols-1 gap-8 min-[960px]:grid-cols-2 min-[1200px]:gap-12">
      <section v-for="(items, domain) in overviewDirectoryGroups" :key="domain" class="min-w-0" :aria-labelledby="`overview-directory-${domain}`" :data-directory-domain="domain">
        <h3 :id="`overview-directory-${domain}`" class="overview-explore__domain flex items-center gap-2"><component :is="domain === 'game' ? PhGameController : PhGlobe" class="shrink-0" aria-hidden="true" />{{ $t(`insights.overviewDirectory.${domain}`) }}</h3>
        <ul class="m-0 list-none p-0">
          <li v-for="item in items" :key="item.path">
            <NuxtLink :to="localePath(item.path)" class="overview-explore__link flex items-center justify-between gap-4 min-[768px]:gap-6" data-directory-link>
              <span class="min-w-0"><strong class="block">{{ $t(`insights.overviewDirectory.links.${item.key}`) }}</strong><span class="overview-explore__description mt-1 block">{{ $t(`insights.editorial.links.${item.key}.description`) }}</span></span><span class="overview-explore__arrow shrink-0" aria-hidden="true">↗</span>
            </NuxtLink>
          </li>
        </ul>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { PhGameController, PhGlobe } from '@phosphor-icons/vue'
import { overviewDirectoryGroups } from '@/utils/insightOverviewPresentation'

const localePath = useLocalePath()
</script>
