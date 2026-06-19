<template>
  <div class="space-y-4">

    <!-- 游戏列表 -->
    <div
        class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3
             lg:grid-cols-4 2xl:grid-cols-5 gap-4"
    >
      <div
          v-for="game in gameList"
          :key="game.id"
          class="p-3 rounded-xl cursor-pointer
         bg-orange-50 hover:bg-orange-100 transition"
          @click="goDetail(game.id)"
      >
        <img
            :src="game.cover"
            class="w-full h-32 object-cover rounded-lg mb-2"
            alt=""
        />

        <div class="flex justify-between items-center gap-1 font-semibold text-sm">
          <div class="flex items-center gap-2 w-[95%]">
            <div class="truncate">
              {{ game.name }}
            </div>
            <div class="rounded-full bg-[#343131] px-2 py-0.5 text-xs text-orange-200 truncate">
              {{ game.primary_tag }}
            </div>
            <div class="overflow-hidden rounded-full truncate bg-orange-200 px-2 py-0.5 text-xs text-[#343131]">
              {{ game.secondary_tag }}
            </div>
          </div>

          <a
              :href="steamPrefix+`${game.appid}`"
              target="_blank"
              rel="noopener noreferrer"
              class="shrink-0"
              @click.stop
          >
            <img
                src="@/assets/icons/steam.svg"
                alt="Steam"
                class="w-4 h-4 opacity-70 hover:opacity-100 transition"
            />
          </a>
        </div>

        <p class="text-xs text-gray-600 mt-1 line-clamp-2 h-[2rem]">
          {{ game.info }}
        </p>

        <div class="mt-2 text-xs text-gray-500 flex justify-between">
          <span class="flex items-center gap-1">
            <img
                src="@/assets/svgs/star.svg"
                alt="评分"
                class="w-3.5 h-3.5"
            />
            <span>
              {{
                game.avg_score > 0
                    ? game.avg_score.toFixed(1)
                    : t("game.panel.none")
              }}
            </span>
          </span>

          <span>{{ game.remark_count }} {{t("game.search.comment")}}</span>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <GamePagination
        :current-page="currentPage"
        :total-pages="totalPages"
        :total="total"
        @page-change="$emit('page-change', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import GamePagination from '@/components/game/search/GamePagination.vue'
import type { SearchPageResponseItem } from '@/types/game.ts'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const router = useRouter()

const steamPrefix = import.meta.env.VITE_STEAM_APP_PREFIX_URL || ''

const goDetail = (id: number | string) => {
  router.push(`/games/${id}`)
}

defineProps<{
  gameList: SearchPageResponseItem[]
  currentPage: number
  totalPages: number
  total: number
}>()

defineEmits<{
  (e: 'page-change', page: number): void
}>()
</script>
