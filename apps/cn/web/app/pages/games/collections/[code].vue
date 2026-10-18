<template>
  <section class="games-page game-collections-page mx-auto w-full max-w-[1200px] px-6 py-10">
    <header class="game-collection-detail-header mb-6">
      <div class="flex flex-wrap items-baseline gap-x-4 gap-y-2">
        <div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-1">
          <h1 class="game-collections-heading break-words">{{ snapshot?.collection.name || t('game.collections.heading') }}</h1>
          <span v-if="snapshot" class="game-collection-note whitespace-nowrap">{{ t('game.collections.count', { count: snapshot.collection.visible_game_count }) }}</span>
        </div>
        <NuxtLink :to="localePath('/games/collections')" class="game-collection-back ml-auto shrink-0">{{ t('game.collections.back') }}</NuxtLink>
      </div>
      <p v-if="snapshot" class="game-collection-copy mt-2">{{ snapshot.collection.info }}</p>
    </header>
    <div v-if="error" class="game-collections-error mb-6 flex flex-wrap items-center gap-4" role="alert">
      <p>{{ t(snapshot ? 'game.collections.refreshFailed' : 'game.collections.unavailable') }}</p>
      <button class="gf-button gf-button--surface" :disabled="pending" @click="refresh">{{ t('game.collections.retry') }}</button>
    </div>
    <GameCollectionTimeline v-if="snapshot?.items.length" :detail="snapshot" :aria-busy="pending" />
    <p v-else-if="snapshot" class="game-collection-note py-8">{{ t('game.collections.emptyWorks') }}</p>
  </section>
</template>
<script setup lang="ts">
import { setResponseStatus } from 'h3'
import GameCollectionTimeline from '~/components/game/collections/GameCollectionTimeline.vue'
import { getGameCollectionDetail } from '~/services/game'
import { buildGameCollectionSeo } from '~/utils/seo'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { useGameCollectionModeRefresh } from '~/composables/useGameCollectionModeRefresh'

definePageMeta({ key: route => route.path })
const { t, locale } = useI18n()
const localePath = useLocalePath()
const lang = locale.value === 'en' ? 'en' : 'zh'
const code = String(useRoute().params.code)
const initial = await useAsyncData(`game-collection:${lang}:${code}`, () => getGameCollectionDetail(code, lang, 'sfw'))
if (!initial.data.value) {
  // This API has authoritative HTTP 404; no lifecycle/message inference.
  if (authoritativePageStatus(initial.error.value, 'collection') === 404) throw createError({ statusCode: 404, statusMessage: 'Not Found' })
  if (import.meta.server) setResponseStatus(useRequestEvent()!, 503)
}
const { snapshot, pending, error, refresh } = useGameCollectionModeRefresh({
  initial: initial.data.value ?? null, initialError: initial.error.value,
  load: mode => getGameCollectionDetail(code, lang, mode),
  onError: cause => { if (authoritativePageStatus(cause, 'collection') === 404) showError({ statusCode: 404, statusMessage: 'Not Found' }) },
})
const seo = computed(() => buildGameCollectionSeo(locale.value, snapshot.value?.collection.info || t('game.collections.intro'), snapshot.value?.collection.name))
useSeoMeta({ title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description })
</script>
