<template>
  <div
      class="w-full min-h-full flex-1 overflow-hidden flex flex-col"
      :style="bgImage ? `background-image: url(${bgImage}); background-size: cover; background-position: center;` : 'background-color: #fef3c7'"
  >
    <div class="px-6 pt-4 shrink-0">
      <div class="inline-flex rounded-xl bg-white/30 backdrop-blur-md p-1">
        <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            class="px-4 py-2 text-sm font-semibold rounded-lg transition"
            :class="activeTab === tab.key
            ? 'bg-white/60 shadow-lg'
            : 'hover:bg-white/20'"
        >
          <span
              :class="activeTab === tab.key
            ? 'text-orange-800 font-bold'
            : 'text-gray-800'"
          >
            {{ tab.label }}
          </span>
        </button>
      </div>
    </div>

    <div class="flex-1 overflow-hidden p-4">
      <div v-if="activeTab === 'global'" class="h-full overflow-auto scrollbar-none">
        <GlobalMetrics />
      </div>

      <div v-else class="flex items-center justify-center h-full w-full overflow-hidden">
        <div
            class="h-full w-full origin-top-left
                 scale-[0.85]
                 translate-x-[7.5%]"
        >
          <Panel />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getImageUrl } from '@/utils/api/stat.ts'
import Panel from '@/components/panel/index.vue'
import GlobalMetrics from "@/components/metrics/GlobalMetrics.vue";
import { i18n } from '@/main.ts'

const { t } = i18n.global
type TabKey = 'global' | 'screen'

const tabs = [
  { key: 'global', label: t("metrics.global") },
  { key: 'screen', label: t("metrics.screen") },
] as const

const activeTab = ref<TabKey>('global')
const bgImage = ref<string | null>(null)

onMounted(async () => {
  try {
    bgImage.value = await getImageUrl()
  } catch (err) {
    console.error('Get background image URL err:', err)
  }
})
</script>

<style>
.scrollbar-none {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.scrollbar-none::-webkit-scrollbar {
  display: none;
}
</style>
