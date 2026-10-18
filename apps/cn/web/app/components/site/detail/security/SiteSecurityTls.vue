<template>
  <div data-site-security-tls class="space-y-6">
    <section data-site-certificate data-site-security-transport-composite class="site-detail-surface">
      <h3 class="site-overview-title">{{ t('siteSecurity.secureTransport') }}</h3>
      <div class="mt-4 grid min-w-0 gap-6 lg:grid-cols-3">
        <div class="min-w-0">
          <h4 class="site-detail-label">{{ t('siteSecurity.transport') }}</h4>
          <p data-site-tls-state :data-state="transport.state" :data-tone="transport.tone" class="site-security-primary-value mt-2">{{ transport.value }}</p>
        </div>
        <div class="min-w-0">
          <h4 class="site-detail-label">{{ t('siteSecurity.verification') }}</h4>
          <p data-site-certificate-verification :data-state="certificate.verification.state" :data-tone="certificate.verification.tone" class="site-security-primary-value mt-2">{{ certificate.verification.label }}</p>
        </div>
        <div class="min-w-0">
          <h4 class="site-detail-label flex items-center gap-2">{{ t('siteSecurity.validity') }}<SiteDetailHelpTooltip :label="t('siteSecurity.validity')" :text="t('siteSecurity.validityHint')" /></h4>
          <p data-site-certificate-expiry :data-expiry="certificate.expiry.state" :data-tone="certificate.expiry.tone" class="site-security-primary-value mt-2">{{ certificate.expiry.value }}</p>
        </div>
      </div>
      <details data-site-transport-details class="site-security-disclosure mt-4">
        <summary>{{ t('siteSecurity.collectorSummary') }}</summary>
        <p v-if="transport.stale" class="site-detail-note mt-2">{{ t('siteSecurity.stale') }}</p>
        <p v-if="transport.truncated" class="site-detail-note mt-2">{{ t('siteSecurity.truncated') }}</p>
        <SiteObservationFacts :items="[...transport.facts, ...transport.errors, ...certificate.errors, ...certificate.validity]" class="mt-3" />
      </details>
    </section>
    <section data-site-certificate-identity class="site-detail-surface site-detail-surface--secondary">
      <h3 class="site-overview-title mb-3">{{ t('siteSecurity.identity') }}</h3>
      <div class="grid min-w-0 gap-5 md:grid-cols-2">
        <div class="min-w-0">
          <h4 class="site-detail-label">{{ t('siteSecurity.subject') }}</h4>
          <SiteObservationFacts :items="certificate.subject" class="mt-2" />
          <details data-site-certificate-san class="site-security-disclosure mt-3">
            <summary>{{ t('siteSecurity.san') }}</summary>
            <ul v-if="certificate.san.length" class="mt-2 space-y-1.5"><li v-for="(name, index) in certificate.san" :key="index" class="site-security-domain-row break-all">{{ name }}</li></ul>
            <p v-else class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
          </details>
        </div>
        <div class="min-w-0">
          <h4 class="site-detail-label">{{ t('siteSecurity.issuer') }}</h4>
          <SiteObservationFacts :items="certificate.issuer" class="mt-2" />
          <details data-site-certificate-chain class="site-security-disclosure mt-3">
            <summary>{{ t('siteSecurity.chain') }}</summary>
            <ol v-if="certificate.chain.length" class="mt-2 space-y-1.5"><li v-for="(name, index) in certificate.chain" :key="index" class="site-security-domain-row min-w-0 break-all"><PhArrowDown v-if="index" class="site-detail-icon mb-2" aria-hidden="true" />{{ name }}</li></ol>
            <p v-else class="site-detail-note mt-2">{{ t('siteSecurity.noEvidence') }}</p>
          </details>
        </div>
      </div>
      <div class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
      <details data-site-certificate-crypto class="site-security-disclosure site-security-crypto">
        <summary>{{ t('siteSecurity.crypto') }}</summary>
        <SiteObservationFacts :items="certificate.crypto" break-all class="mt-2" />
        <p class="site-detail-note mt-2">{{ t('siteSecurity.cryptoHint') }}</p>
      </details>
    </section>
  </div>
</template>
<script setup lang="ts">
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import { PhArrowDown } from '@phosphor-icons/vue'
import SiteDetailHelpTooltip from '../SiteDetailHelpTooltip.vue'
import SiteObservationFacts from '../observation/SiteObservationFacts.vue'
defineProps<{ transport: SiteSecurityPresentation['transport']; certificate: SiteSecurityPresentation['certificate'] }>()
const { t } = useI18n()
</script>
