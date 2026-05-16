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

<style scoped>
.quick-access-grid {
  display: grid;
  width: max-content;
  max-width: 100%;
  gap: 1rem;
}

.quick-access-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  width: max-content;
  padding: 0.55rem 0.65rem 0.65rem;
  border-radius: 1rem;
}

.quick-access-title {
  display: inline-flex;
  align-items: center;
  width: max-content;
  max-width: 100%;
  padding: 0.28rem 0.56rem;
  border-radius: 999px;
  background: rgba(8, 12, 18, 0.24);
  backdrop-filter: blur(10px);
  color: rgba(255, 247, 237, 0.92);
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.quick-access-tiles {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.4rem;
  justify-items: start;
}

.quick-site-tile {
  display: flex;
  width: 2.5rem;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 0;
  border-radius: 0.75rem;
  background: rgba(0, 0, 0, 0.18);
  color: rgba(255, 248, 241, 0.94);
  backdrop-filter: blur(10px);
  transition: background 180ms ease, opacity 180ms ease;
}

.quick-site-tile:hover:not(:disabled) {
  background: rgba(241, 245, 249, 0.3);
}

.quick-site-tile:disabled {
  cursor: default;
}

.quick-site-tile-empty {
  opacity: 0.12;
}

.quick-site-icon,
.quick-site-fallback {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 0.375rem;
}

.quick-site-icon {
  object-fit: cover;
}

.quick-site-fallback {
  display: grid;
  place-items: center;
  background: rgba(17, 24, 39, 0.84);
  font-size: 0.72rem;
  font-weight: 700;
}

.quick-site-manage {
  background: rgba(0, 0, 0, 0.16);
}

.quick-site-manage:hover:not(:disabled) {
  background: rgba(241, 245, 249, 0.3);
}

.quick-site-manage-icon {
  display: grid;
  place-items: center;
  width: 1.5rem;
  height: 1.5rem;
  font-size: 1.05rem;
  line-height: 1;
  color: rgba(255, 248, 241, 0.94);
}

@media (min-width: 768px) {
  .quick-access-grid {
    grid-template-columns: repeat(2, max-content);
    gap: 1.1rem;
  }
}
</style>
