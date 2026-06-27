<template>
  <div class="games-page steam-workshop-page relative isolate min-h-screen overflow-hidden">
    <GoFurryGridBackground :fixed="false" palette="games" />

    <main class="relative z-10 mx-auto flex w-full max-w-[1800px] flex-1 flex-col px-4 pb-20 pt-8 sm:px-6 lg:pt-10 xl:px-8">
      <section class="steam-workshop-top" aria-label="Workshop navigation">
        <div class="steam-workshop-title-row">
          <h1 class="steam-workshop-title">
            {{ t('steamZone.title') }}
          </h1>

          <NuxtLink
            class="steam-workshop-search"
            :to="localePath('/games/search')"
          >
            {{ t('steamZone.searchGames') }}
          </NuxtLink>
        </div>

        <nav class="steam-workshop-tabs" aria-label="Workshop tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="steam-workshop-tab"
            :class="{ 'steam-workshop-tab--active': activeTab === tab.key }"
            type="button"
            :aria-selected="activeTab === tab.key"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </nav>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

type WorkshopTabKey = 'all' | 'discussions' | 'screenshots' | 'artwork' | 'broadcasts' | 'videos' | 'workshop' | 'news' | 'guides' | 'reviews'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const activeTab = ref<WorkshopTabKey>('workshop')
const isEnglish = computed(() => locale.value === 'en')

const tabs = computed<Array<{ key: WorkshopTabKey, label: string }>>(() => {
  if (isEnglish.value) {
    return [
      { key: 'all', label: 'All' },
      { key: 'discussions', label: 'Discussions' },
      { key: 'screenshots', label: 'Screenshots' },
      { key: 'artwork', label: 'Artwork' },
      { key: 'broadcasts', label: 'Broadcasts' },
      { key: 'videos', label: 'Videos' },
      { key: 'workshop', label: 'Workshop' },
      { key: 'news', label: 'News' },
      { key: 'guides', label: 'Guides' },
      { key: 'reviews', label: 'Reviews' },
    ]
  }

  return [
    { key: 'all', label: '全部' },
    { key: 'discussions', label: '讨论' },
    { key: 'screenshots', label: '截图' },
    { key: 'artwork', label: '艺术作品' },
    { key: 'broadcasts', label: '直播' },
    { key: 'videos', label: '视频' },
    { key: 'workshop', label: '工坊' },
    { key: 'news', label: '新闻' },
    { key: 'guides', label: '指南' },
    { key: 'reviews', label: '评测' },
  ]
})

useHead(() => ({
  title: t('steamZone.title'),
}))
</script>

<style scoped>
.steam-workshop-top {
  width: 100%;
  color: rgba(241, 245, 249, 0.96);
}

.steam-workshop-title-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  min-height: 36px;
  margin: 0;
  padding: 0;
}

.steam-workshop-title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  color: rgba(241, 245, 249, 0.98);
  font-size: clamp(1rem, 1.2vw, 1.125rem);
  font-weight: 600;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.steam-workshop-search {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 32px;
  border: 1px solid rgba(125, 211, 252, 0.28);
  border-radius: 2px;
  background: #2f8fce;
  color: #fff;
  font-size: 0.875rem;
  font-weight: 400;
  padding: 0 16px;
  text-decoration: none;
  transition:
    background-color 180ms ease,
    border-color 180ms ease;
}

.steam-workshop-search:hover {
  border-color: rgba(186, 230, 253, 0.5);
  background: #38a3e2;
}

.steam-workshop-tabs {
  display: flex;
  min-height: 36px;
  align-items: stretch;
  gap: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.steam-workshop-tabs::-webkit-scrollbar {
  display: none;
}

.steam-workshop-tab {
  --steam-workshop-tab-gap: 18px;

  position: relative;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-start;
  border: 0;
  background: transparent;
  color: rgba(226, 232, 240, 0.72);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 var(--steam-workshop-tab-gap) 0 0;
  text-align: left;
  transition:
    color 160ms ease;
  white-space: nowrap;
}

.steam-workshop-tab:hover {
  color: rgba(248, 250, 252, 0.94);
}

.steam-workshop-tab--active {
  color: #fff;
}

.steam-workshop-tab--active::after {
  position: absolute;
  right: var(--steam-workshop-tab-gap);
  bottom: 0;
  left: 0;
  height: 3px;
  background: #66c0f4;
  content: '';
}

html:not(.dark) .steam-workshop-tab {
  color: rgba(50, 36, 24, 0.72);
}

html:not(.dark) .steam-workshop-tab:hover,
html:not(.dark) .steam-workshop-tab--active {
  color: rgba(50, 36, 24, 0.96);
}

html:not(.dark) .steam-workshop-tab--active::after {
  background: rgba(50, 36, 24, 0.88);
}

@media (max-width: 639px) {
  .steam-workshop-title-row {
    min-height: 34px;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
    padding: 0;
  }

  .steam-workshop-search {
    min-height: 30px;
    padding: 0 12px;
  }

  .steam-workshop-tab {
    --steam-workshop-tab-gap: 14px;
  }
}
</style>
