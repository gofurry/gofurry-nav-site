<template>
  <div class="site-icon-list flex flex-wrap gap-2">
    <a
        v-for="(item, i) in items"
        :key="item.key + i"
        :href="item.value"
        target="_blank"
        rel="noopener noreferrer"
        class="site-icon-link flex h-8 w-8 items-center justify-center"
        :title="item.key"
    >
      <img
          :src="getIconUrl(item.key)"
          :class="[
            'site-icon-image h-5 w-5 object-contain',
            { 'site-icon-image--official': normalizeIconKey(item.key) === 'official' }
          ]"
          :alt="item.key"
      />
    </a>
  </div>
</template>

<script setup lang="ts">
import platformIcons from '../../data/platform-icons.json'
import type { KvModel } from '@/types/game'

defineProps<{
  items: KvModel[]
}>()

const ICON_MAP: Record<string, string> = platformIcons

const BASE_URL = '/web/platform-icons/'

const normalizeIconKey = (key: string) => key.trim().toLowerCase()

const getIconUrl = (key: string) => {
  const iconKey = normalizeIconKey(key)
  return ICON_MAP[iconKey] ? BASE_URL + ICON_MAP[iconKey] : '/defaultLogo.svg'
}
</script>
