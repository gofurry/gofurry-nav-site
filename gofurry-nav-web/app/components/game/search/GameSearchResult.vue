<template>
  <div class="space-y-4">

    <!-- 游戏列表 -->
    <div class="grid grid-cols-2 gap-4 md:grid-cols-3 2xl:grid-cols-5">
      <div
          v-for="game in gameList"
          :key="game.id"
          class="search-page-card group"
          @click="goDetail(game.id)"
      >
        <img
            :src="game.cover"
            class="mb-2 aspect-[16/9] w-full rounded-lg object-cover"
            :alt="game.name"
        />

        <div class="flex justify-between items-center gap-1 font-semibold text-sm">
          <div class="flex items-center gap-2 w-[95%]">
            <div class="truncate">
              {{ game.name }}
            </div>
            <div class="search-page-tag search-page-tag--primary">
              {{ game.primary_tag }}
            </div>
            <div class="search-page-tag search-page-tag--secondary">
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

        <p class="search-page-desc">
          {{ game.info }}
        </p>

        <div class="search-page-meta">
          <span class="flex items-center gap-1">
            <img
                src="@/assets/svgs/star.svg"
                alt=""
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
import GamePagination from '@/components/game/search/GamePagination.vue'
import type { SearchPageResponseItem } from '@/types/game'
import { i18n } from '@/main'

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

<style scoped>
.search-page-card {
  cursor: pointer;
  overflow: hidden;
  border: 1px solid rgba(126, 92, 58, 0.12);
  border-radius: 0.92rem;
  background: rgba(255, 250, 242, 0.40);
  padding: 0.72rem;
  transition: background-color 180ms ease, border-color 180ms ease;
}

.search-page-card:hover {
  border-color: rgba(180, 96, 24, 0.32);
  background: rgba(255, 239, 213, 0.68);
}

.search-page-tag {
  overflow: hidden;
  border-radius: 999px;
  padding: 0.12rem 0.48rem;
  font-size: 0.72rem;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-page-tag--primary {
  background: rgba(45, 35, 28, 0.88);
  color: rgba(255, 226, 189, 0.92);
}

.search-page-tag--secondary {
  background: rgba(255, 224, 186, 0.78);
  color: rgba(45, 35, 28, 0.88);
}

.search-page-desc {
  margin-top: 0.25rem;
  display: -webkit-box;
  height: 2.2rem;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgba(87, 83, 78, 0.72);
  font-size: 0.78rem;
  line-height: 1.35;
}

.search-page-meta {
  margin-top: 0.62rem;
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgba(87, 83, 78, 0.66);
  font-size: 0.76rem;
}

:global(.games-search-page.games-page--dark) .search-page-card {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(226, 232, 240, 0.060);
}

:global(.games-search-page.games-page--dark) .search-page-card:hover {
  border-color: rgba(203, 213, 225, 0.36);
  background: rgba(226, 232, 240, 0.12);
}

:global(.games-search-page.games-page--dark) .search-page-tag--primary {
  background: rgba(15, 23, 42, 0.58);
  color: rgba(226, 232, 240, 0.86);
}

:global(.games-search-page.games-page--dark) .search-page-tag--secondary {
  background: rgba(148, 163, 184, 0.14);
  color: rgba(203, 213, 225, 0.78);
}

:global(.games-search-page.games-page--dark) .search-page-desc,
:global(.games-search-page.games-page--dark) .search-page-meta {
  color: rgba(203, 213, 225, 0.66);
}

:global(.games-search-page.games-page--dark) .search-page-card .font-semibold {
  color: rgba(241, 245, 249, 0.90);
}
</style>
