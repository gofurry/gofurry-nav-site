<template>
  <div class="space-y-6">
    <!-- HTTP 记录面板 -->
    <div class="bg-orange-50 rounded-xl overflow-hidden">
      <div class="shadow-sm bg-orange-200 p-5 cursor-pointer flex items-center justify-between" @click="togglePanel('http')">
        <h3 class="font-semibold flex items-center gap-2">
          {{ t('site.siteHttpPanel.httpRecord') }}
        </h3>
      </div>

      <div v-if="panels.http" class="p-5 pt-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          <!-- 访问 URL -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.visitUrl') }}</h4>
            <div class="flex items-center mb-3 group relative">
              <p
                  class="text-sm font-mono hover:text-orange-400 cursor-pointer"
                  @click="copyToClipboard(record.url)"
              >
                {{ record.url }}
              </p>
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
          </div>

          <!-- 状态码 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.statusCode') }}</h4>
            <span
                :class="[
                'px-2 py-0.5 rounded-full text-xs',
                record.statusCode < 400
                  ? 'bg-green-100 text-green-800'
                  : 'bg-red-100 text-red-800'
              ]"
            >
              {{ record.statusCode }}
            </span>
          </div>

          <!-- 响应时间 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500  mb-2">{{ t('site.siteHttpPanel.responseTime') }}</h4>
            <p class="text-sm">{{ record.responseTime }}</p>
          </div>

          <!-- 内容大小 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.contentSize') }}</h4>
            <p class="text-sm">{{ formatBytes(record.contentLength) }}</p>
          </div>

          <!-- 服务器 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.server') }}</h4>
            <p class="text-sm">{{ record.server }}</p>
          </div>

          <!-- 页面标题 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.pageTitle') }}</h4>
            <p class="text-sm">{{ record.title }}</p>
          </div>
        </div>

        <!-- 跳转记录 -->
        <div v-if="record.redirects?.length" class="mb-6">
          <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.redirectPath') }}</h4>
          <div class="flex flex-wrap items-center gap-2 text-sm font-mono">
            <template v-for="(r, idx) in record.redirects" :key="idx">
              <span>{{ r }}</span>
            </template>
          </div>
        </div>

        <!-- 响应头 -->
        <div class="mb-6">
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-bold text-gray-500">{{ t('site.siteHttpPanel.responseHeader') }}</h4>
            <button class="text-xs text-orange-800" @click.stop="toggleSection('httpHeaders')">
              {{ sections.httpHeaders ? t('common.collapse') : t('common.expand') }}
            </button>
          </div>

          <div
              v-if="sections.httpHeaders"
              class="bg-orange-100 rounded-lg p-3 text-sm"
          >
            <div v-for="(vals, key) in record.headers" :key="key" class="mb-1">
              <span class="font-bold">{{ key }}:</span>
              {{ vals.join(', ') }}
            </div>
          </div>
        </div>

        <!-- 元数据 -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-bold text-gray-500">{{ t('site.siteHttpPanel.metadata') }}</h4>
            <button class="text-xs text-orange-800" @click.stop="toggleSection('httpMeta')">
              {{ sections.httpMeta ? t('common.collapse') : t('common.expand') }}
            </button>
          </div>

          <div v-if="sections.httpMeta" class="space-y-3 text-sm bg-orange-100 rounded-lg p-3">
            <div><span class="font-bold">{{ t('site.siteHttpPanel.encoding') }}:</span> {{ record.meta.charset }}</div>
            <div><span class="font-bold">{{ t('site.siteHttpPanel.description') }}:</span> {{ record.meta.description }}</div>
            <div>
              <span class="font-bold">{{ t('site.siteHttpPanel.keywords') }}:</span>
              <div class="flex flex-wrap gap-1 mt-1">
                <span
                    v-for="(kw, idx) in record.meta.keywords.split(/[,\s]+/)"
                    :key="idx"
                    class="px-2 py-0.5 bg-orange-50 rounded-full text-xs"
                >#{{ kw }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TLS 记录面板 -->
    <div class="bg-orange-50 rounded-xl overflow-hidden">
      <div class="shadow-sm bg-orange-200 p-5 cursor-pointer flex items-center justify-between" @click="togglePanel('tls')">
        <h3 class="font-semibold flex items-center gap-2">
          {{ t('site.siteHttpPanel.tlsRecord') }}
        </h3>
      </div>

      <div v-if="panels.tls" class="p-5 pt-4">
        <div class="space-y-4 text-sm">
          <!-- 加密协议 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.encryptionProtocol') }}</h4>
            <span
              :class="[
                'px-2 py-0.5 rounded-full text-xs',
                record.tlsVersion === 'TLS1.3'
                  ? 'bg-green-100 text-green-800'
                  : record.tlsVersion === 'TLS1.2'
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800'
              ]"
            >
              {{ record.tlsVersion }}
            </span>
          </div>

          <!-- 密码套件 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-2">{{ t('site.siteHttpPanel.cipherSuite') }}</h4>
            <p class="text-sm">{{ record.cipherSuite }}</p>
          </div>

          <!-- 证书信息 -->
          <div>
            <h4 class="text-sm font-bold text-gray-500 mb-1">{{ t('site.siteHttpPanel.certInfo') }}</h4>
            <div class="bg-orange-100 rounded-lg p-3 space-y-2">
              <div>
                <span class="font-bold">{{ t('site.siteHttpPanel.validityPeriod') }}:</span> {{ record.certExpiry }}（{{ t('site.siteHttpPanel.remaining') }}
                <span
                  :class="{
                    'text-green-500': certDaysNumber > 90,
                    'text-yellow-500': certDaysNumber <= 90 && certDaysNumber > 30,
                    'text-red-500': certDaysNumber <= 30
                  }"
                >
                  {{ certDaysNumber }}
                </span> {{ t('site.siteHttpPanel.day') }})
              </div>
              <div><span class="font-bold">{{ t('site.siteHttpPanel.issuer') }}:</span> {{ record.certIssuer }}</div>
              <div><span class="font-bold">{{ t('site.siteHttpPanel.organization') }}:</span> {{ record.certIssuerOrg.join(', ') }}</div>
              <div><span class="font-bold">{{ t('site.siteHttpPanel.coveredDomains') }}:</span> {{ record.certDNSNames.join(', ') }}</div>
              <div><span class="font-bold">{{ t('site.siteHttpPanel.isCaCert') }}:</span> {{ record.certIsCA ? t('common.yes') : t('common.no') }}</div>
            </div>
          </div>

          <!-- 算法细节 -->
          <div>
            <button class="text-xs text-orange-800" @click.stop="toggleSection('tlsDetails')">
              {{ sections.tlsDetails ? t('site.siteHttpPanel.collapseAlgDetails') : t('site.siteHttpPanel.expandAlgDetails') }}
            </button>

            <div
                v-if="sections.tlsDetails"
                class="bg-orange-100 rounded-lg p-3 mt-2 space-y-2"
            >
              <div><span class="font-bold">{{ t('site.siteHttpPanel.publicKeyAlg') }}:</span> {{ record.certPubKeyAlg }}</div>
              <div><span class="font-bold">{{ t('site.siteHttpPanel.signatureAlg') }}:</span> {{ record.certSigAlg }}</div>
              <div v-if="record.certEmail"><span class="font-bold">{{ t('site.siteHttpPanel.email') }}:</span> {{ record.certEmail }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed  } from 'vue'
import type { HttpRecord } from '@/types/nav'
import {i18n} from "@/main";

const t = (key: string) => i18n.global.t(key)

const { record } = defineProps<{ record: HttpRecord }>()


const copied = ref(false)
function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

const panels = ref({ http: true, tls: true })
const sections = ref({ httpHeaders: false, httpMeta: false, tlsDetails: false })

const togglePanel = (key: keyof typeof panels.value) => (panels.value[key] = !panels.value[key])
const toggleSection = (key: keyof typeof sections.value) => (sections.value[key] = !sections.value[key])

const formatBytes = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB']
  let i = -1
  do {
    bytes = bytes / 1024
    i++
  } while (bytes >= 1024 && i < units.length - 1)
  return `${bytes.toFixed(1)} ${units[i]}`
}

const certDaysNumber = computed(() => {
  const match = record.certDaysLeft?.match(/\d+/)
  return match ? Number(match[0]) : 0
})
</script>
