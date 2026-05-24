<template>
  <div ref="pageRoot" class="flex min-h-full w-full flex-col overflow-x-hidden bg-orange-50 text-gray-800">
    <div v-if="pending" class="flex flex-1 items-center justify-center text-gray-500">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="error" class="flex flex-1 items-center justify-center text-red-500">
      {{ loadFailedText }}
    </div>

    <div v-else class="contents">
      <div class="mx-10 my-8">
        <SiteOverview
          v-if="sitePageData.siteInfo"
          :site="{
            name: sitePageData.siteInfo.name || '',
            icon: sitePageData.siteInfo.icon || undefined,
            domain: sitePageData.domain || '',
            welfare: sitePageData.siteInfo.welfare === '1',
            nsfw: sitePageData.siteInfo.nsfw === '1',
            description: sitePageData.siteInfo.info || '',
            viewCount: sitePageData.siteInfo.view_count ?? 0,
          }"
        />
      </div>

      <div v-if="healthSummaryEnabled && (sitePageData.siteHealthSummary || sitePageData.targetHealthSummary)" class="mx-10 mb-8">
        <SiteHealthSummaryPanel
          :site-summary="sitePageData.siteHealthSummary"
          :target-summary="sitePageData.targetHealthSummary"
        />
      </div>

      <div class="mx-10 mb-8">
        <SitePerformance
          v-if="sitePageData.sitePingRecord && sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :http-record="sitePageData.siteHttpRecord"
        />
      </div>

      <div class="mx-10 mb-8">
        <SiteHttpPanel
          v-if="sitePageData.siteHttpRecord"
          :record="sitePageData.siteHttpRecord"
        />
      </div>

      <div class="mx-10 mb-8">
        <SiteDnsPanel
          v-if="sitePageData.siteDnsRecord"
          :record="sitePageData.siteDnsRecord"
        />
      </div>

      <div class="mb-8 mr-4 flex flex-wrap items-center justify-center gap-3 text-orange-800">
        <button
          class="flex items-center justify-center gap-2 rounded-lg bg-orange-300 px-4 py-2 text-sm transition-colors hover:bg-orange-200"
          @click="generateReport"
        >
          {{ t('common.save') }}
        </button>

        <button
          class="flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm transition-colors hover:bg-orange-100"
          @click="() => refresh()"
        >
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import SiteDnsPanel from '@/components/site/SiteDnsPanel.vue'
import SiteHealthSummaryPanel from '@/components/site/SiteHealthSummaryPanel.vue'
import SiteHttpPanel from '@/components/site/SiteHttpPanel.vue'
import SiteOverview from '@/components/site/SiteOverview.vue'
import SitePerformance from '@/components/site/SitePerformance.vue'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'

const { t } = useI18n()
const { data, pending, error, refresh, healthSummaryEnabled } = await useSiteDetailPage()
const sitePageData = computed(() => data.value!)
const pageRoot = ref<HTMLElement | null>(null)
const loadFailedText = computed(() => (t('common.loading') === 'Loading...' ? 'Failed to load site data.' : '站点数据加载失败。'))

const seoTitle = computed(() => {
  const name = sitePageData.value.siteInfo?.name?.trim()
  return name ? `${name} - gofurry` : 'gofurry Sites'
})
const seoDescription = computed(() => {
  const description = sitePageData.value.siteInfo?.info?.trim() ?? ''
  return description.slice(0, 160)
})

useSeoMeta({
  title: () => seoTitle.value,
  description: () => seoDescription.value,
  ogTitle: () => seoTitle.value,
  ogDescription: () => seoDescription.value,
})

const generateReport = () => {}
</script>
