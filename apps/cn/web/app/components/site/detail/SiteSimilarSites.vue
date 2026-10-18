<template>
  <section id="site-similar-workspace" data-site-similar :data-variant="variant" class="site-detail-similar min-w-0"
    :role="variant === 'panel' ? 'tabpanel' : undefined" :aria-labelledby="variant === 'panel' ? 'site-tab-similar' : 'site-similar-title'" :tabindex="variant === 'panel' ? 0 : undefined">
    <h2 id="site-similar-title" class="site-overview-title mb-3">{{ t('siteDetail.similarSites') }}</h2>
    <ul v-if="items.length" class="site-detail-similar-list grid min-w-0 gap-1.5">
      <li v-for="site in items" :key="site.id" class="min-w-0">
        <NuxtLink :to="localePath(siteEntityPath(site.id))" :data-site-similar-id="site.id" class="site-detail-similar-item grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-start gap-x-3">
          <ManagedAssetImage :object-key="site.icon || undefined" alt="" class="site-detail-similar-icon row-span-2 object-contain" />
          <div class="flex min-w-0 flex-wrap items-baseline justify-between gap-x-2 gap-y-1">
            <span class="site-detail-similar-name min-w-0 break-words">{{ site.name }}</span>
            <span data-site-similar-views class="site-detail-similar-views inline-flex shrink-0 items-center gap-1" :aria-label="t('siteDetail.views') + ' ' + views(site.view_count)"><PhEye class="site-detail-icon" aria-hidden="true" />{{ views(site.view_count) }}</span>
          </div>
          <p v-if="site.info" class="site-detail-similar-info col-start-2 min-w-0 line-clamp-1">{{ site.info }}</p>
        </NuxtLink>
      </li>
    </ul>
    <p v-else class="site-detail-note" data-site-similar-empty>{{ t('siteDetail.noVisibleSimilarSites') }}</p>
  </section>
</template>
<script setup lang="ts">
import { PhEye } from '@phosphor-icons/vue'
import ManagedAssetImage from '~/components/common/ManagedAssetImage.vue'
import type { Site } from '~/types/nav'
import { siteEntityPath } from '~/utils/siteRoutes'
defineProps<{ items: Site[]; variant: 'aside' | 'panel' }>()
const { t, locale } = useI18n(), localePath = useLocalePath()
const views = (count: number) => new Intl.NumberFormat(locale.value).format(count)
</script>
