<template>
  <div data-site-web class="space-y-5">
    <section data-site-web-metadata class="site-observation-composite">
      <h3 class="site-observation-heading">{{ t('siteObservation.sections.metadata') }}</h3>
      <dl class="mt-4 space-y-3">
        <div v-for="item in presentation.metadata.slice(0, 2)" :key="item.key" :data-site-evidence="item.key">
          <dt class="site-observation-caption">{{ item.label }}</dt><dd class="site-observation-profile-content mt-1 break-words">{{ item.value }}</dd>
        </div>
      </dl>
      <dl class="site-observation-profile-meta mt-4 flex flex-wrap gap-x-6 gap-y-2">
        <div v-for="item in presentation.metadata.slice(2)" :key="item.key" :data-site-evidence="item.key" class="flex min-w-0 flex-wrap gap-x-2">
          <dt>{{ item.label }}</dt><dd class="min-w-0 break-words">{{ item.value }}</dd>
        </div>
      </dl>
    </section>
    <div data-site-web-mosaic class="grid min-w-0 items-start gap-4 lg:grid-cols-2">
      <section v-for="probe in presentation.probes" :key="probe.protocol" :data-site-web-probe="probe.protocol" :data-site-web-state="probe.state" class="site-observation-evidence min-w-0">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h3 class="site-observation-heading flex items-center gap-2"><component :is="probeIcons[probe.protocol]" class="site-detail-icon" aria-hidden="true" />{{ t('siteObservation.probes.' + probe.protocol) }}</h3>
          <p class="site-observation-probe-state" :data-tone="probe.tone">{{ probe.stateLabel }}</p>
        </div>
        <SiteObservationEvidence :items="probe.summaryFacts" />
        <details class="site-observation-disclosure mt-3">
          <summary>{{ t('siteObservation.probeDetails') }}</summary>
          <p class="site-observation-caption mt-2">{{ probe.statusLabel }} · {{ t('siteObservation.fields.duration') }} {{ probe.duration }}</p>
          <section v-for="section in probe.details" :key="section.key" class="mt-3">
            <h4 class="site-observation-caption mb-1">{{ section.title }}</h4><SiteObservationEvidence :items="section.items" />
          </section>
        </details>
        <p class="site-observation-probe-footer mt-4 break-words">{{ t('siteObservation.lastObserved') }} · {{ probe.observed }}</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhRobot, PhFileText, PhImages, PhGlobe } from '@phosphor-icons/vue'
import SiteObservationEvidence from './SiteObservationEvidence.vue'
defineProps<{ presentation: SiteObservationPresentation['web'] }>()
const { t } = useI18n()
const probeIcons: Record<string, Component> = { robots: PhRobot, llms_txt: PhFileText, page_assets: PhImages, rdap: PhGlobe }
</script>
