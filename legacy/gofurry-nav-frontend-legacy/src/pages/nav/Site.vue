<template>
  <div ref="pageRoot" class="flex flex-col w-full min-h-full overflow-x-hidden bg-orange-50 text-gray-800">


    <div v-if="loading" class="flex-1 flex items-center justify-center text-gray-500">{{t("common.loading")}}</div>

    <div v-else-if="errorMsg" class="flex-1 flex items-center justify-center text-red-500">
      {{ errorMsg }}
    </div>

    <div class="mx-10 my-8">
      <SiteOverview
          v-if="siteInfo"
          :site="{
          name: siteInfo.name || '',
          icon: siteInfo.icon || undefined,
          domain: domain || '',
          welfare: siteInfo.welfare === '1',
          nsfw: siteInfo.nsfw === '1',
          description: siteInfo.info || ''
        }"
      />
    </div>

    <div class="mx-10 mb-8">
      <SitePerformance
          v-if="sitePingRecord && siteHttpRecord"
          :pingRecord="sitePingRecord"
          :httpRecord="siteHttpRecord"
      />
    </div>

    <div class="mx-10 mb-8">
      <SiteHttpPanel
          v-if="siteHttpRecord"
          :record="siteHttpRecord"
      />
    </div>

    <div class="mx-10 mb-8">
      <SiteDnsPanel
          v-if="siteDnsRecord"
          :record="siteDnsRecord"
      />
    </div>

    <div class="mb-8 mr-4 flex flex-wrap gap-3 justify-center items-center text-orange-800">
      <button
          class="px-4 py-2 bg-orange-300 hover:bg-orange-200 rounded-lg text-sm flex justify-center items-center gap-2 transition-colors"
          @click="generateReport"
      >
        {{t("common.save")}}
      </button>

      <button
          class="px-4 py-2 hover:bg-orange-100 rounded-lg text-sm flex justify-center items-center gap-2 transition-colors"
          @click="loadData"
      >
        {{t("common.refresh")}}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useLangStore } from '@/store/langStore.ts'
import { getSiteDetail, getSitePingRecord, getSiteHttpRecord, getSiteDnsRecord } from '@/utils/api/nav.ts'
import SiteOverview from '@/components/site/SiteOverview.vue'
import SitePerformance from '@/components/site/SitePerformance.vue'
import SiteHttpPanel from '@/components/site/SiteHttpPanel.vue'
import SiteDnsPanel from '@/components/site/SiteDnsPanel.vue'
import { i18n } from '@/main.ts'
import type { SiteInfo, PingRecord, HttpRecord, DnsRecord, DnsItem } from '@/types/nav.ts'
import { safeJsonParse } from '@/utils/util.ts'

const t = (key: string) => i18n.global.t(key)
const route = useRoute()
const langStore = useLangStore()

const siteId = route.params.id as string
const domain = route.query.domain as string

const loading = ref(true)
const errorMsg = ref('')
const siteInfo = ref<SiteInfo | null>(null)
const sitePingRecord = ref<PingRecord | null>(null)
const siteHttpRecord = ref<HttpRecord | null>(null)
const siteDnsRecord = ref<DnsRecord | null>(null)

const pageRoot = ref<HTMLElement | null>(null)

async function loadData() {
  loading.value = true
  errorMsg.value = ''
  try {
    const lang = langStore.lang
    const [info, http, dns, ping] = await Promise.all([
      getSiteDetail(siteId, lang),
      getSiteHttpRecord(domain),
      getSiteDnsRecord(domain),
      getSitePingRecord(domain)
    ])

    siteInfo.value = info
    siteHttpRecord.value = safeJsonParse<HttpRecord>(http as any)

    const parsedDns = { ...dns }
    for (const key in parsedDns) {
      if (typeof parsedDns[key as keyof DnsRecord] === 'string') {
        parsedDns[key as keyof DnsRecord] = safeJsonParse<DnsItem[]>(parsedDns[key as keyof DnsRecord]) as any
      }
    }
    siteDnsRecord.value = parsedDns
    sitePingRecord.value = ping
  } catch (err: any) {
    console.error('鍔犺浇璇︽儏椤垫暟鎹け璐?', err)
    errorMsg.value = '鍔犺浇绔欑偣淇℃伅澶辫触锛岃绋嶅悗鍐嶈瘯'
  } finally {
    loading.value = false
  }
}

async function loadSiteInfoOnly() {
  try {
    const lang = langStore.lang
    siteInfo.value = await getSiteDetail(siteId, lang)
  } catch (e) {
    console.error('鍔犺浇绔欑偣鍩虹淇℃伅澶辫触:', e)
  }
}

onMounted(() => loadData())
watch(() => langStore.lang, () => loadSiteInfoOnly())

const generateReport = () => {
}
</script>

<style scoped>
</style>
