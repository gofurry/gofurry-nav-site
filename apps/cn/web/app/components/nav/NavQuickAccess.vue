<template>
  <div class="quick-access-grid">
    <section class="quick-access-section">
      <p class="quick-access-title">{{ t('customSites.recentTitle') }}</p>
      <div class="quick-access-tiles">
        <button
          v-for="item in recentSlots"
          :key="`recent-${item?.id ?? 'empty'}`"
          class="quick-site-tile"
          :class="{ 'quick-site-tile-empty': !item }"
          type="button"
          :disabled="!item"
          :title="item?.name || ''"
          @click="item && emit('visit-recent', item)"
        >
          <template v-if="item">
            <img
              v-if="!failedRecentIcons[item.id]"
              :src="`https://favicon.im/${toExternalUrl(item.url)}?larger=true`"
              :alt="item.name"
              class="quick-site-icon"
              loading="lazy"
              @error="markRecentIconFailed(item.id)"
            />
            <div v-else class="quick-site-fallback">{{ item.name.slice(0, 1).toUpperCase() }}</div>
          </template>
        </button>
      </div>
    </section>

    <section class="quick-access-section">
      <p class="quick-access-title">{{ t('customSites.customTitle') }}</p>
      <div class="quick-access-tiles">
        <button
          v-for="item in customEntries"
          :key="item.key"
          class="quick-site-tile"
          :class="{
            'quick-site-tile-empty': item.kind === 'empty',
            'quick-site-manage': item.kind === 'manage',
          }"
          type="button"
          :disabled="item.kind === 'empty'"
          :title="item.kind === 'site' ? item.site.name : item.kind === 'manage' ? t('customSites.manageTitle') : ''"
          @click="handleCustomEntryClick(item)"
        >
          <template v-if="item.kind === 'site'">
            <img
              v-if="!failedCustomIcons[item.site.id]"
              :src="`https://favicon.im/${toExternalUrl(item.site.url)}?larger=true`"
              :alt="item.site.name"
              class="quick-site-icon"
              loading="lazy"
              @error="markCustomIconFailed(item.site.id)"
            />
            <div v-else class="quick-site-fallback">{{ item.site.name.slice(0, 1).toUpperCase() }}</div>
          </template>
          <template v-else-if="item.kind === 'manage'">
            <span class="quick-site-manage-icon" aria-hidden="true">+</span>
          </template>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { toExternalUrl, type RecentSiteItem } from '@/utils/recentSites'
import type { CustomSiteItem } from '@/utils/customSites'

const props = defineProps<{
  recentSites: RecentSiteItem[]
  customSites: CustomSiteItem[]
}>()

const emit = defineEmits<{
  (e: 'visit-recent', item: RecentSiteItem): void
  (e: 'visit-custom', item: CustomSiteItem): void
  (e: 'manage'): void
}>()

type CustomEntry =
  | { key: string; kind: 'site'; site: CustomSiteItem }
  | { key: string; kind: 'manage' }
  | { key: string; kind: 'empty' }

const { t } = useI18n()
const failedRecentIcons = ref<Record<string, boolean>>({})
const failedCustomIcons = ref<Record<string, boolean>>({})

const recentSlots = computed(() => {
  const filled = props.recentSites.slice(0, 8)
  return [...filled, ...Array.from({ length: Math.max(0, 8 - filled.length) }, () => null)]
})

const customEntries = computed<CustomEntry[]>(() => {
  const entries: CustomEntry[] = props.customSites.slice(0, 7).map(site => ({
    key: `site-${site.id}`,
    kind: 'site',
    site,
  }))

  entries.push({ key: 'manage', kind: 'manage' })

  while (entries.length < 8) {
    entries.push({ key: `empty-${entries.length}`, kind: 'empty' })
  }

  return entries
})

function handleCustomEntryClick(item: CustomEntry) {
  if (item.kind === 'site') {
    emit('visit-custom', item.site)
    return
  }

  if (item.kind === 'manage') {
    emit('manage')
  }
}

function markRecentIconFailed(id: string) {
  failedRecentIcons.value = {
    ...failedRecentIcons.value,
    [id]: true,
  }
}

function markCustomIconFailed(id: string) {
  failedCustomIcons.value = {
    ...failedCustomIcons.value,
    [id]: true,
  }
}
</script>
