<template>
  <section class="metadata-probe-shell">
    <div class="site-info-tabs-panel">
      <div class="info-tabs-header">
        <h3 class="info-tabs-title">{{ activeInfoTabTitle }}</h3>
        <div class="info-tabs-nav">
          <button
            v-for="tab in infoTabs"
            :key="tab.key"
            type="button"
            class="info-tab-button"
            :class="{ 'is-active': activeInfoTab === tab.key }"
            @click="activeInfoTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <SiteMetadataRows
        v-if="activeInfoTab === 'metadata'"
        :empty-text="label('暂无数据', 'No data')"
        :items="pageInfoItems"
      />
      <SiteChangeEvents
        v-else-if="activeInfoTab === 'changes'"
        :events="changeEvents"
        :loading="changesLoading"
      />
      <SiteObservationHistoryPanel
        v-else
        :histories="observationHistories"
        :loading="observationsLoading"
        :summary-for="observationSummary"
        @set-page="setObservationPage"
      />
    </div>

    <SiteLightProbePanel
      :detail-sections-for="lightProbeDetailSections"
      :entries="lightProbeEntries"
    />
  </section>
</template>

<script setup lang="ts">
import SiteChangeEvents from '@/components/site/SiteChangeEvents.vue'
import SiteLightProbePanel from '@/components/site/SiteLightProbePanel.vue'
import SiteMetadataRows from '@/components/site/SiteMetadataRows.vue'
import SiteObservationHistoryPanel from '@/components/site/SiteObservationHistoryPanel.vue'
import { useSiteMetadataProbePanel, type SiteMetadataProbePanelProps } from '~/composables/useSiteMetadataProbePanel'

const props = defineProps<SiteMetadataProbePanelProps>()
const {
  activeInfoTab,
  activeInfoTabTitle,
  changeEvents,
  changesLoading,
  infoTabs,
  label,
  lightProbeDetailSections,
  lightProbeEntries,
  observationHistories,
  observationsLoading,
  observationSummary,
  pageInfoItems,
  setObservationPage,
} = useSiteMetadataProbePanel(props)
</script>

<style scoped>
.metadata-probe-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: clamp(1.5rem, 3vw, 2.3rem);
}

.site-info-tabs-panel {
  min-width: 0;
}

.info-tabs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.25rem;
  min-height: 2.45rem;
}

.info-tabs-title {
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.35;
}

:global(html.dark .info-tabs-title){
  color: #f8fafc;
}

.info-tabs-nav {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.95rem;
  background: transparent;
}

.info-tab-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  border-radius: 0;
  padding: 0.28rem 0 0.34rem;
  color: #475569;
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.15;
  transition: color 500ms ease;
}

.info-tab-button::after {
  content: "";
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: 2px;
  border-radius: 999px;
  background: rgba(251, 140, 47, 0.72);
  opacity: 0;
  transform: scaleX(0);
  transform-origin: left center;
  transition: opacity 500ms ease, transform 500ms ease;
}

:global(html.dark .info-tab-button){
  color: #cbd5e1;
}

.info-tab-button:hover,
.info-tab-button.is-active {
  color: #8a3c13;
}

:global(html.dark .info-tab-button:hover),
:global(html.dark .info-tab-button.is-active){
  color: #fed7aa;
}

.info-tab-button:hover::after,
.info-tab-button.is-active::after {
  opacity: 1;
  transform: scaleX(1);
}

:global(html.dark .info-tab-button::after){
  background: rgba(251, 146, 60, 0.70);
}

@media (max-width: 640px) {
  .info-tabs-header {
    flex-direction: column;
  }

  .info-tabs-nav {
    width: 100%;
    justify-content: flex-start;
  }

  .info-tab-button {
    flex: 0 0 auto;
  }
}

@media (min-width: 1280px) {
  .metadata-probe-shell {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.05fr);
  }
}
</style>
