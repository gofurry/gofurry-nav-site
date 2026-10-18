<template>
  <div data-site-security :data-site-security-target="presentation.target" class="site-security min-w-0">
    <header data-site-workspace-subnav-header class="flex min-w-0 items-center justify-between gap-3">
      <h2 class="site-detail-workspace-title shrink-0">{{ t('siteDetail.tabs.security') }}</h2>
      <SiteSecurityNav :active="view" class="min-w-0 flex-1 sm:flex-initial" @select="emit('select', $event)" />
    </header>
    <section id="site-security-panel" :data-site-security-view="view" role="tabpanel" :aria-labelledby="'security-tab-' + view" tabindex="0" class="site-security-panel mt-5 min-w-0">
      <SiteSecurityOverview v-if="view === 'overview'" :presentation="presentation.overview" />
      <SiteSecurityTls v-else-if="view === 'tls'" :transport="presentation.transport" :certificate="presentation.certificate" />
      <SiteSecurityWeb v-else-if="view === 'web'" :headers="presentation.headers" :security-txt="presentation.securityTxt" :raw-headers-to="rawHeadersTo" />
      <SiteSecurityExposure v-else :ports="presentation.portCheck" :waf="presentation.wafCanary" />
    </section>
  </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { SiteSecurityView } from '~/utils/siteDetailRouteState'
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import SiteSecurityNav from './SiteSecurityNav.vue'
import SiteSecurityOverview from './SiteSecurityOverview.vue'
import SiteSecurityTls from './SiteSecurityTls.vue'
import SiteSecurityWeb from './SiteSecurityWeb.vue'
import SiteSecurityExposure from './SiteSecurityExposure.vue'
defineProps<{ view: SiteSecurityView; presentation: SiteSecurityPresentation; rawHeadersTo: RouteLocationRaw }>()
const emit = defineEmits<{ select: [view: SiteSecurityView] }>()
const { t } = useI18n()
</script>
