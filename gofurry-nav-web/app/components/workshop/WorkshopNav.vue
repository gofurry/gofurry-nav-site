<template>
  <section class="workshop-top" aria-label="Workshop navigation">
    <div class="workshop-title-row">
      <h1 class="workshop-title">
        {{ t('workshop.title') }}
      </h1>

      <NuxtLink
        class="workshop-search"
        :to="localePath('/games/search')"
      >
        {{ t('workshop.searchGames') }}
      </NuxtLink>
    </div>

    <nav class="workshop-tabs" aria-label="Workshop tabs">
      <NuxtLink
        v-for="tab in tabs"
        :key="tab.key"
        class="workshop-tab"
        :class="{ 'workshop-tab--active': isActive(tab.to) }"
        :to="tab.href"
        :aria-current="isActive(tab.to) ? 'page' : undefined"
      >
        {{ tab.label }}
      </NuxtLink>
    </nav>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

type WorkshopTab = {
  key: 'all' | 'discussion' | 'developer' | 'tools'
  label: string
  to: string
  href: string
}

const route = useRoute()
const localePath = useLocalePath()
const { t } = useI18n()

const normalizedPath = computed(() => normalizeRoutePath(route.path))
const tabs = computed<WorkshopTab[]>(() => [
  { key: 'all', label: t('workshop.tabs.all'), to: '/workshop', href: localePath('/workshop') },
  { key: 'discussion', label: t('workshop.tabs.discussion'), to: '/workshop/discussion', href: localePath('/workshop/discussion') },
  { key: 'developer', label: t('workshop.tabs.developer'), to: '/workshop/developer', href: localePath('/workshop/developer') },
  { key: 'tools', label: t('workshop.tabs.tools'), to: '/workshop/tools', href: localePath('/workshop/tools') },
])

function normalizeRoutePath(path: string) {
  const normalized = path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/'
  return normalized.length > 1 ? normalized.replace(/\/$/, '') : normalized
}

function isActive(tabPath: string) {
  if (tabPath === '/workshop') {
    return normalizedPath.value === '/workshop'
  }

  return normalizedPath.value === tabPath || normalizedPath.value.startsWith(`${tabPath}/`)
}
</script>

<style scoped>
.workshop-top {
  width: 100%;
  color: rgba(241, 245, 249, 0.96);
}

.workshop-title-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  min-height: 36px;
  margin: 0;
  padding: 0;
}

.workshop-title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  color: rgba(226, 232, 240, 0.92);
  font-size: clamp(1.325rem, 1.55vw, 1.45rem);
  font-weight: 600;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workshop-search {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 36px;
  border: 1px solid rgba(125, 211, 252, 0.28);
  border-radius: 2px;
  background: #227ebb;
  color: rgba(248, 250, 252, 0.92);
  font-size: 0.9375rem;
  font-weight: 400;
  padding: 0 16px;
  text-decoration: none;
  transition:
    background-color 180ms ease,
    border-color 180ms ease;
}

.workshop-search:hover {
  border-color: rgba(186, 230, 253, 0.5);
  background: #38a3e2;
}

.workshop-tabs {
  display: flex;
  min-height: 36px;
  align-items: stretch;
  gap: 0;
  overflow-x: auto;
  scrollbar-width: none;
  margin-top: 0.5rem;
}

.workshop-tabs::-webkit-scrollbar {
  display: none;
}

.workshop-tab {
  --workshop-tab-gap: 18px;

  position: relative;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-start;
  border: 0;
  background: transparent;
  color: rgba(203, 213, 225, 0.72);
  cursor: pointer;
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 var(--workshop-tab-gap) 0 0;
  text-align: left;
  text-decoration: none;
  transition:
    color 160ms ease;
  white-space: nowrap;
}

.workshop-tab:hover {
  color: rgba(226, 232, 240, 0.90);
}

.workshop-tab--active {
  color: rgba(226, 232, 240, 0.94);
}

.workshop-tab--active::after {
  position: absolute;
  right: var(--workshop-tab-gap);
  bottom: 0;
  left: 0;
  height: 3px;
  background: #66c0f4;
  content: '';
}

html:not(.dark) .workshop-tab {
  color: rgba(82, 61, 45, 0.72);
}

html:not(.dark) .workshop-title {
  color: rgba(72, 52, 37, 0.90);
}

html:not(.dark) .workshop-search {
  border-color: rgba(191, 103, 65, 0.22);
  background: #d47a54;
  color: rgba(255, 250, 244, 0.96);
}

html:not(.dark) .workshop-search:hover {
  border-color: rgba(177, 86, 51, 0.30);
  background: #c96f4d;
}

html:not(.dark) .workshop-tab:hover,
html:not(.dark) .workshop-tab--active {
  color: rgba(68, 47, 32, 0.88);
}

html:not(.dark) .workshop-tab--active::after {
  background: rgba(90, 65, 45, 0.82);
}

@media (max-width: 639px) {
  .workshop-title-row {
    min-height: 34px;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
    padding: 0;
  }

  .workshop-search {
    min-height: 30px;
    padding: 0 12px;
  }

  .workshop-tab {
    --workshop-tab-gap: 14px;
  }
}
</style>
