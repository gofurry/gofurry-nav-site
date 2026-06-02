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
  border-radius: 8px;
  background:
    linear-gradient(118deg, rgba(255, 237, 213, 0.55), rgba(255, 250, 242, 0.78) 48%, rgba(239, 246, 255, 0.32)),
    rgba(255, 247, 235, 0.70);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.performance-panel :deep(section > .rounded-2xl::before) {
  content: "";
  pointer-events: none;
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, transparent, rgba(79, 111, 237, 0.08), transparent),
    repeating-linear-gradient(90deg, rgba(79, 111, 237, 0.055) 0 1px, transparent 1px 76px);
  mask-image: linear-gradient(to bottom, transparent 0, #000 18%, #000 82%, transparent 100%);
  opacity: 0.5;
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

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2)) {
  align-self: center;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2) > div) {
  gap: 0.2rem;
  border-radius: 8px;
  background: rgba(255, 232, 196, 0.34);
  box-shadow: none;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child button) {
  border-radius: 7px;
  padding: 0.62rem 1rem;
  font-size: 0.88rem;
  transition-duration: 500ms;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child button.bg-orange-200) {
  background: #fdba74;
  color: #111827;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
  gap: 0.55rem;
}

.performance-panel :deep(section > .rounded-2xl > .grid:first-child > div:last-child > div) {
  min-width: 4.8rem;
  border-radius: 7px;
  background: rgba(255, 232, 196, 0.38);
  box-shadow: none;
}

.performance-panel :deep(section > .rounded-2xl > .rounded-xl) {
  position: relative;
  z-index: 1;
  overflow: visible;
  border-radius: 0;
  background:
    linear-gradient(rgba(255, 250, 242, 0.30), rgba(255, 250, 242, 0.30)),
    repeating-linear-gradient(0deg, transparent 0 39px, rgba(148, 163, 184, 0.13) 40px),
    repeating-linear-gradient(90deg, transparent 0 71px, rgba(79, 111, 237, 0.08) 72px);
  box-shadow: none;
}

.performance-panel :deep(section > .rounded-2xl.bg-orange-100\/45) {
  padding: clamp(1rem, 1.6vw, 1.25rem);
}

.performance-panel :deep(section > .rounded-2xl > .rounded-xl.bg-orange-50\/70) {
  margin-top: 0.2rem;
  padding: 0.5rem 0.2rem 0.9rem;
  background-color: transparent;
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
