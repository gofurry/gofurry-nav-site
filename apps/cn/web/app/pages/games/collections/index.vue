<template>
  <section class="games-page game-collections-page mx-auto w-full max-w-[1440px] px-6 py-10">
    <h1 class="sr-only">{{ t('game.collections.heading') }}</h1>
    <div class="games-search-overlay-scope search-toolbar mb-8 flex w-full items-center gap-3">
      <div class="search-shell relative min-w-0 flex-1">
        <img src="~/assets/svgs/search.svg" alt="" class="game-sidebar-search-icon absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
        <input type="text" role="searchbox" :value="inputValue" :aria-label="t('game.collections.search')" :placeholder="t('game.collections.search')" :aria-busy="pending" class="game-sidebar-search-input w-full py-2 pl-9 pr-3" @input="onInput" @compositionstart="onCompositionStart" @compositionend="onCompositionEnd" />
      </div>
      <button type="button" class="search-filter-button shrink-0" @click="filterOpen = true">{{ t('game.collections.advancedFilter') }}</button>
    </div>
    <Teleport to="body">
      <div v-if="filterOpen" class="games-search-overlay-scope">
        <GameCollectionFilter :criteria="criteria" @close="filterOpen = false" @apply="filter => { filterOpen = false; applyFilter(filter) }" />
      </div>
    </Teleport>
    <div v-if="snapshot" class="game-collections-grid grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3" :aria-busy="pending">
      <GameCollectionCard v-for="collection in snapshot.items" :key="collection.code" :collection="collection" />
    </div>
    <p v-if="snapshot && !snapshot.items.length" class="game-collection-note py-8">{{ t('game.collections.empty') }}</p>
    <div v-if="error" class="game-collections-error mt-6 flex flex-wrap items-center gap-4" role="alert">
      <p>{{ t(snapshot ? 'game.collections.refreshFailed' : 'game.collections.unavailable') }}</p>
      <button class="gf-button gf-button--surface" :disabled="pending" @click="refresh">{{ t('game.collections.retry') }}</button>
    </div>
    <div v-if="snapshot?.has_more && !error" class="game-collections-load mt-8 flex flex-col items-center gap-3">
      <p v-if="loadError" role="alert">{{ t('game.collections.loadFailed') }}</p>
      <button class="gf-button gf-button--surface" :disabled="loadingMore || pending" @click="loadMore">
        {{ t(loadingMore ? 'game.collections.loading' : loadError ? 'game.collections.retry' : 'game.collections.loadMore') }}
      </button>
    </div>
  </section>
</template>
<script setup lang="ts">
import { setResponseStatus } from 'h3'
import GameCollectionCard from '~/components/game/collections/GameCollectionCard.vue'
import { getGameCollections } from '~/services/game'
import { buildGameCollectionSeo } from '~/utils/seo'
import { useGameCollectionDiscovery } from '~/composables/useGameCollectionDiscovery'
import GameCollectionFilter from '~/components/game/collections/GameCollectionFilter.vue'

definePageMeta({ key: route => route.path })
const { t, locale } = useI18n()
const lang = locale.value === 'en' ? 'en' : 'zh'
const initial = await useAsyncData(`game-collections:${lang}`, () => getGameCollections(lang, 'sfw'))
if (import.meta.server && !initial.data.value) setResponseStatus(useRequestEvent()!, 503)
const app = useNuxtApp()
const filterOpen = ref(false)
const { snapshot, criteria, inputValue, pending, error, refresh, loadingMore, loadError,
  loadMore, applyFilter, onInput, onCompositionStart, onCompositionEnd } = useGameCollectionDiscovery({
  initial: initial.data.value ?? null, initialError: initial.error.value,
  load: (mode, page, criteria) => app.runWithContext(() => getGameCollections(lang, mode, page, 24, criteria)),
})
const seo = computed(() => buildGameCollectionSeo(locale.value, t('game.collections.intro')))
useSeoMeta({ title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description })
</script>
