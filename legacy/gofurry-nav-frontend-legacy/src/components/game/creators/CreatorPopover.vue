<template>
  <div
      v-if="creator"
      class="creator-popover z-50 w-80
           bg-orange-100 rounded-xl shadow-2xl p-4
           transition-opacity duration-200"
      :style="style"
      @mouseenter="$emit('mouseenter')"
      @mouseleave="$emit('mouseleave')"
  >
    <!-- 名称 -->
    <p class="text-sm font-semibold mb-1 text-gray-800">
      {{ creator.name }}
    </p>

    <!-- Info -->
    <p class="text-xs text-gray-600 leading-relaxed mb-3 overflow-hidden"
       style="-webkit-line-clamp: 10; display: -webkit-box; -webkit-box-orient: vertical;">
      {{ creator.info || 'No description' }}
    </p>

    <!-- Links -->
    <div v-if="creator.links?.length" class="mb-2 max-h-24 overflow-hidden">
      <p class="text-xs font-semibold text-gray-700 mb-1">
        {{t("game.creator.link")}}
      </p>
      <SiteIconList :items="creator.links" />
    </div>

    <!-- Contact -->
    <div v-if="creator.contact?.length">
      <p class="text-xs font-semibold text-gray-700 mb-1">
        {{ t("game.creator.contact") }}
      </p>

      <ul class="grid grid-cols-1 gap-x-2 gap-y-1">
        <li
            v-for="c in creator.contact.slice(0, 3)"
            :key="c.key"
            class="text-xs text-gray-600 truncate"
            title="{{ c.key }}：{{ c.value }}"
        >
          {{ c.key }}：{{ c.value }}
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CreatorResponse } from '@/types/game'
import SiteIconList from "@/components/common/SiteIconList.vue";
import {i18n} from "@/main.ts";

const { t } = i18n.global

defineProps<{
  creator: CreatorResponse | null
  style: Record<string, string>
}>()

defineEmits<{
  (e: 'mouseenter'): void
  (e: 'mouseleave'): void
}>()
</script>
