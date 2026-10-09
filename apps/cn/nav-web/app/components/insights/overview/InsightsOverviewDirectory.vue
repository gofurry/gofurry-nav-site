<template>
  <section class="overview-explore" aria-labelledby="overview-explore-title" data-overview-explore>
    <h2 id="overview-explore-title" class="overview-heading mb-6">{{ $t('insights.overviewDirectory.title') }}</h2>
    <div class="grid grid-cols-1 gap-8 min-[960px]:grid-cols-2 min-[1200px]:gap-12">
      <section v-for="(items, domain) in overviewDirectoryGroups" :key="domain" class="min-w-0" :aria-labelledby="`overview-directory-${domain}`" :data-directory-domain="domain">
        <h3 :id="`overview-directory-${domain}`">{{ $t(`insights.overviewDirectory.${domain}`) }}</h3>
        <ul class="m-0 list-none p-0">
          <li v-for="item in items" :key="item.path">
            <NuxtLink :to="localePath(item.path)" class="overview-explore__link flex items-center justify-between gap-6" data-directory-link>
              <span class="min-w-0"><strong class="block">{{ $t(`insights.overviewDirectory.links.${item.key}`) }}</strong><span class="block">{{ $t(`insights.editorial.links.${item.key}.description`) }}</span></span><span class="shrink-0" aria-hidden="true">↗</span>
            </NuxtLink>
          </li>
        </ul>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { overviewDirectoryGroups } from '@/utils/insightOverviewPresentation'

const localePath = useLocalePath()
</script>
