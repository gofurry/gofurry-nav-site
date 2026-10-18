<script setup lang="ts">
import { PhArrowRight } from '@phosphor-icons/vue'
import UpdateMetadata from '~/components/updates/UpdateMetadata.vue'
import { useUpdateIndex } from '~/composables/useUpdateIndex'
import { groupUpdatesByMonth, updateCopy } from '~/utils/updatePresentation'
import { formatUpdatesDate } from '~/utils/updatesDate'

const { locale } = useI18n()
const localePath = useLocalePath()
const lang = computed(() => locale.value === 'en' ? 'en' : 'zh')
const copy = computed(() => updateCopy[lang.value])
const { items, total, pending, state, refresh, hasMore, loadingMore, moreError, loadMore } = await useUpdateIndex(lang)
const latest = computed(() => items.value[0])
const groups = computed(() => groupUpdatesByMonth(items.value.slice(1)))


useSeoMeta({
  title: () => copy.value.seoTitle, description: () => copy.value.description,
  ogTitle: () => copy.value.seoTitle, ogDescription: () => copy.value.description,
  twitterTitle: () => copy.value.seoTitle, twitterDescription: () => copy.value.description,
})
</script>

<template>
  <div class="updates-page" data-updates-index :data-updates-state="state">
    <div class="updates-main mx-auto w-full max-w-[1080px] px-5 py-8 md:px-10 md:py-10">
      <header class="updates-header mb-6 md:mb-8">
        <h1 class="updates-title">{{ copy.heading }}</h1>
        <div v-if="state === 'ready' && latest" class="updates-index-meta mt-3 flex flex-wrap gap-x-4 gap-y-1">
          <span>{{ total }} {{ copy.count }}</span>
          <span>{{ copy.latestDate }} <time :datetime="latest.published_at">{{ formatUpdatesDate(latest.published_at, lang === 'en' ? 'en-US' : 'zh-CN') }}</time></span>
        </div>
      </header>
      <section :aria-busy="pending" :aria-label="copy.heading">
        <p v-if="state === 'loading'" class="updates-state" role="status">{{ copy.loading }}</p>
        <div v-else-if="state === 'error'" class="updates-state" role="alert">
          <p>{{ copy.error }}</p>
          <button type="button" class="gf-button gf-button--ghost mt-4" @click="refresh()">{{ copy.retry }}</button>
        </div>
        <p v-else-if="state === 'empty'" class="updates-state" role="status">{{ copy.empty }}</p>
        <template v-else-if="latest">
          <article class="updates-latest p-5 md:p-6" data-update-latest>
            <div class="flex flex-wrap items-center gap-4">
              <p class="updates-eyebrow">{{ copy.latest }}</p>
              <span v-if="latest.version" class="updates-version" data-update-version>{{ latest.version }}</span>
            </div>
            <h2 class="updates-latest-title mt-2"><NuxtLink :to="localePath(`/updates/${latest.id}`)">{{ latest.title }}</NuxtLink></h2>
            <p v-if="latest.summary" class="updates-summary mt-2" data-update-summary>{{ latest.summary }}</p>
            <UpdateMetadata class="mt-3" :published-at="latest.published_at" :commit-sha="latest.commit_sha" />
            <NuxtLink :to="localePath(`/updates/${latest.id}`)" class="updates-action mt-4 inline-flex items-center gap-2" data-update-detail-link>
              {{ copy.read }}<PhArrowRight aria-hidden="true" />
            </NuxtLink>
          </article>
          <section v-if="groups.length" class="mt-7 md:mt-9" :aria-label="copy.history">
            <section v-for="group in groups" :key="group.month" class="updates-month mt-5 grid gap-2 md:grid-cols-[110px_minmax(0,1fr)] md:gap-5" :data-updates-month="group.month">
              <h2 class="updates-month-title">{{ group.month }}</h2>
              <ol class="grid min-w-0">
                <li v-for="item in group.items" :key="item.id">
                  <article class="updates-history-row px-2 py-3 md:px-3" :data-update-history="item.id">
                    <h3 class="updates-history-title"><NuxtLink :to="localePath(`/updates/${item.id}`)" data-update-detail-link>{{ item.title }}</NuxtLink></h3>
                    <p v-if="item.summary" class="updates-history-summary mt-2" data-update-summary>{{ item.summary }}</p>
                    <UpdateMetadata class="mt-2" :published-at="item.published_at" :commit-sha="item.commit_sha" />
                  </article>
                </li>
              </ol>
            </section>
          </section>
          <div v-if="hasMore" class="mt-6 flex flex-col items-center gap-2" :aria-busy="loadingMore">
            <p v-if="moreError" class="updates-note" role="alert">{{ copy.moreError }}</p>
            <button type="button" class="updates-more" data-updates-more :disabled="loadingMore" @click="loadMore">{{ loadingMore ? copy.loading : moreError ? copy.retry : copy.more }}</button>
          </div>
        </template>
      </section>
    </div>
  </div>
</template>
