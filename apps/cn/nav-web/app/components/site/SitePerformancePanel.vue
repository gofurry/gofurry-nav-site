<template>
  <section class="performance-panel">
    <SitePerformance
      :domain="domain"
      :http-record="httpRecord"
      :ping-record="pingRecord"
      :site-id="siteId"
      :target-latest-core="targetLatestCore"
    />
  </section>
</template>

<script setup lang="ts">
import SitePerformance from '@/components/site/SitePerformance.vue'
import type { HttpRecord, PingRecord, TargetLatestResponse } from '@/types/nav'

defineProps<{
  domain?: string
  httpRecord: HttpRecord
  pingRecord: PingRecord | null
  siteId?: string | number
  targetLatestCore?: TargetLatestResponse | null
}>()
</script>

<style scoped>
.performance-panel {
  margin-top: 1.5rem;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.performance-panel :deep(section > .grid:first-child) {
  display: none;
}

.performance-panel :deep(section > .rounded-2xl) {
  position: relative;
  overflow: visible;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:global(html.dark .performance-panel section > .rounded-2xl){
  background: transparent;
  box-shadow: none;
}

.performance-panel :deep(section > .rounded-2xl::before) {
  content: none;
}

:global(html.dark .performance-panel section > .rounded-2xl::before){
  content: none;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child) {
  position: relative;
  z-index: 1;
  margin-bottom: 1rem;
  align-items: center;
  gap: 0.8rem;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:first-child) {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:first-child > div) {
  display: none;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:first-child h3) {
  color: #0f172a;
  font-size: clamp(1.08rem, 1.4vw, 1.28rem);
}

:global(html.dark .performance-panel section > .rounded-2xl > .grid:first-child > div:first-child h3){
  color: #f8fafc;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2)) {
  align-self: center;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2) > div) {
  gap: 0.2rem;
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.42);
  box-shadow: none;
}

:global(html.dark .performance-panel section > .rounded-2xl > .grid:first-child > div:nth-child(2) > div) {
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.10);
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child button) {
  border-radius: 7px;
  padding: 0.62rem 1rem;
  font-size: 0.88rem;
  transition-duration: 500ms;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child button.bg-orange-200) {
  background: rgba(251, 146, 60, 0.22);
  color: #7c2d12;
}

:global(html.dark .performance-panel section > .rounded-2xl > .grid:first-child button.bg-orange-200){
  background: rgba(251, 146, 60, 0.22);
  color: #fed7aa;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
  gap: 0.55rem;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child > div) {
  min-width: 4.8rem;
  border-radius: 7px;
  background: rgba(255, 250, 242, 0.42);
  box-shadow: none;
}

:global(html.dark .performance-panel section > .rounded-2xl > .grid:first-child > div:last-child > div){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.10);
}

.performance-panel :deep(section > .rounded-2xl > .rounded-xl) {
  position: relative;
  z-index: 1;
  overflow: visible;
  border-radius: 0.55rem;
  background:
    linear-gradient(rgba(255, 250, 242, 0.44), rgba(255, 250, 242, 0.44)),
    repeating-linear-gradient(0deg, transparent 0 39px, rgba(148, 163, 184, 0.038) 40px),
    repeating-linear-gradient(90deg, transparent 0 71px, rgba(79, 111, 237, 0.020) 72px);
  box-shadow: none;
}

:global(html.dark .performance-panel section > .rounded-2xl > .rounded-xl){
  background:
    linear-gradient(rgba(15, 23, 42, 0.48), rgba(15, 23, 42, 0.48)),
    repeating-linear-gradient(0deg, transparent 0 39px, rgba(148, 163, 184, 0.040) 40px),
    repeating-linear-gradient(90deg, transparent 0 71px, rgba(147, 197, 253, 0.024) 72px);
}

.performance-panel :deep(section > .rounded-2xl.bg-orange-100\/45) {
  padding: clamp(1rem, 1.6vw, 1.25rem);
}

.performance-panel :deep(section > .rounded-2xl > .rounded-xl.bg-orange-50\/70) {
  margin-top: 0.2rem;
  padding: 0.5rem 0.2rem 0.9rem;
  background-color: rgba(255, 250, 242, 0.44);
}

:global(html.dark .performance-panel section > .rounded-2xl > .rounded-xl.bg-orange-50\/70) {
  background-color: rgba(15, 23, 42, 0.48);
}

.performance-panel :deep(section > .rounded-2xl > .rounded-xl > div) {
  height: 320px;
}

@media (max-width: 1023px) {
  .performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
    justify-self: center;
  }

  .performance-panel :deep(section > .rounded-2xl > .rounded-xl > div) {
    height: 300px;
  }
}

@media (max-width: 640px) {
  .performance-panel :deep(section > .rounded-2xl) {
    padding: 1rem;
  }

  .performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:first-child) {
    text-align: center;
  }

  .performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    justify-self: stretch;
  }

  .performance-panel :deep(section > .rounded-2xl > .rounded-xl) {
    padding: 0.35rem 0.35rem 1.15rem;
  }

  .performance-panel :deep(section > .rounded-2xl > .rounded-xl.bg-orange-50\/70) {
    padding: 0.35rem 0.35rem 1.15rem;
  }

  .performance-panel :deep(section > .rounded-2xl > .rounded-xl > div) {
    height: 300px;
  }
}
</style>
