<template>
  <div class="mb-8 p-5">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-2xl font-bold text-gray-800">
        {{ group.title }}
      </h3>

      <NuxtLink
        to="/games/search"
        class="cursor-pointer rounded-md p-2 text-md text-orange-900 transition hover:bg-orange-200/50 hover:text-orange-700"
      >
        {{ t('common.showMore') }}
      </NuxtLink>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
      <div
        v-for="item in visibleGames"
        :key="item.id"
        class="cursor-pointer rounded-lg p-2 transition hover:bg-orange-200/50"
      >
        <img
          :src="item.cover"
          class="mb-2 h-32 w-full rounded-md object-cover"
          :alt="item.name"
          @click.stop="goGameDetail(item.id)"
        />

        <p class="line-clamp-1 text-sm font-semibold text-gray-900">
          {{ item.name }}
        </p>

        <p class="mt-1 h-[2rem] overflow-hidden text-xs text-gray-600">
          {{ item.desc }}
        </p>

        <div class="mt-2">
          <RatingStar :score="item.score" :count="item.scoreCount" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import RatingStar from '@/components/common/RatingStar.vue'

const { t } = useI18n()
const router = useRouter()

interface GameItem {
  id: string
  name: string
  cover: string
  desc: string
  score: number
  scoreCount: number
}

interface GameGroup {
  title: string
  games: GameItem[]
}

const props = defineProps<{
  group: GameGroup
}>()

defineEmits<{
  (e: 'more', group: GameGroup): void
}>()

function goGameDetail(id: string) {
  router.push(`/games/${id}`)
}

const screenWidth = ref(import.meta.client ? window.innerWidth : 1280)

function updateWidth() {
  if (import.meta.client) {
    screenWidth.value = window.innerWidth
  }
}

onMounted(() => {
  updateWidth()
  window.addEventListener('resize', updateWidth)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateWidth)
})

const visibleGames = computed(() => {
  if (screenWidth.value >= 1024) {
    return props.group.games.slice(0, 8)
  }

  return props.group.games.slice(0, 6)
})
</script>
