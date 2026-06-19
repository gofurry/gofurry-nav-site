<template>
  <div class="games-page steam-zone-page relative isolate min-h-screen overflow-hidden">
    <GoFurryGridBackground :fixed="false" palette="games" />

    <main class="relative z-10 mx-auto flex w-full max-w-[1880px] flex-1 flex-col px-4 pb-20 pt-8 sm:px-6 lg:pt-12 xl:px-8">
      <section class="game-detail-tabs steam-zone-panel p-6 sm:p-8 lg:p-10">
        <p class="steam-zone-eyebrow text-sm font-semibold uppercase tracking-[0.24em]">
          {{ t('steamZone.eyebrow') }}
        </p>
        <div class="mt-4 grid gap-8 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div>
            <h1 class="steam-zone-title text-4xl font-semibold leading-tight sm:text-5xl">
              {{ t('steamZone.title') }}
            </h1>
            <p class="steam-zone-description mt-4 max-w-3xl text-base leading-7 sm:text-lg">
              {{ t('steamZone.description') }}
            </p>
          </div>

          <NuxtLink class="game-detail-load-more steam-zone-action px-5 py-2 text-sm font-semibold" :to="localePath('/games')">
            {{ t('steamZone.backToGames') }}
          </NuxtLink>
        </div>

        <div class="steam-zone-grid mt-10 grid gap-4 md:grid-cols-3">
          <article v-for="item in cards" :key="item.title" class="steam-zone-card rounded-2xl p-5">
            <h2 class="text-lg font-semibold">{{ item.title }}</h2>
            <p class="mt-2 text-sm leading-6">{{ item.description }}</p>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const { t } = useI18n()
const localePath = useLocalePath()

const cards = computed(() => [
  {
    title: t('steamZone.cards.discovery.title'),
    description: t('steamZone.cards.discovery.description'),
  },
  {
    title: t('steamZone.cards.assets.title'),
    description: t('steamZone.cards.assets.description'),
  },
  {
    title: t('steamZone.cards.community.title'),
    description: t('steamZone.cards.community.description'),
  },
])

useHead(() => ({
  title: t('steamZone.title'),
}))
</script>

<style scoped>
.steam-zone-panel {
  min-height: 420px;
}

.steam-zone-eyebrow {
  color: rgba(251, 146, 60, 0.92);
}

.steam-zone-title {
  color: rgba(255, 250, 244, 0.96);
}

.steam-zone-description,
.steam-zone-card p {
  color: rgba(255, 250, 244, 0.72);
}

.steam-zone-card {
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(15, 23, 42, 0.78);
  color: rgba(255, 250, 244, 0.92);
  transition:
    border-color 500ms ease,
    background-color 500ms ease;
}

.steam-zone-card:hover {
  border-color: rgba(251, 146, 60, 0.5);
  background: rgba(30, 41, 59, 0.88);
}

html:not(.dark) .steam-zone-title {
  color: rgba(67, 36, 18, 0.96);
}

html:not(.dark) .steam-zone-description,
html:not(.dark) .steam-zone-card p {
  color: rgba(67, 36, 18, 0.72);
}

html:not(.dark) .steam-zone-card {
  border-color: rgba(146, 64, 14, 0.16);
  background: rgba(255, 247, 237, 0.94);
  color: rgba(67, 36, 18, 0.94);
}

html:not(.dark) .steam-zone-card:hover {
  border-color: rgba(194, 65, 12, 0.34);
  background: rgba(255, 237, 213, 0.96);
}
</style>
