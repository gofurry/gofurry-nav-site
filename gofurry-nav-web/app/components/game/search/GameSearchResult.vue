<template>
  <div class="search-results space-y-4">

    <!-- 游戏列表 -->
    <div class="search-result-grid-shell">
      <div class="search-result-grid">
        <div
            v-for="game in gameList"
            :key="game.id"
            class="search-page-card group"
            @click="goDetail(game.id)"
        >
          <div class="relative mb-2">
            <img
                :src="game.cover"
                class="search-page-cover"
                :alt="game.name"
            />

            <button
                class="search-review-button"
                type="button"
                :aria-label="`${game.name} review`"
                @click.stop="openReview(game)"
            >
              <svg
                  aria-hidden="true"
                  viewBox="0 0 24 24"
                  class="h-[1.125rem] w-[1.125rem]"
              >
                <path
                    fill="currentColor"
                    d="M9 22c-.55 0-1-.45-1-1v-3H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v12c0 1.11-.89 2-2 2h-6.1l-3.7 3.71c-.2.19-.45.29-.7.29H9m1-6v3.08L13.08 16H20V4H4v12h6m5.84-7.8l-1.01 1.01l-2.07-2.03l1.01-1.02c.2-.21.54-.22.78 0l1.29 1.25c.21.21.22.55 0 .79M8 11.91l4.17-4.19l2.07 2.08l-4.16 4.2H8v-2.09Z"
                />
              </svg>
            </button>
          </div>

          <div class="search-page-title-row">
            <div class="search-page-title-wrap">
              <div class="search-page-title truncate">
                {{ game.name }}
              </div>
              <div v-if="game.primary_tag" class="search-page-tag search-page-tag--primary">
                {{ game.primary_tag }}
              </div>
              <div v-if="game.secondary_tag" class="search-page-tag search-page-tag--secondary">
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
            <span class="search-page-score">
              <img
                  src="@/assets/svgs/star.svg"
                  alt=""
                  class="w-3.5 h-3.5"
              />
              <span class="search-page-score-value">
                {{
                  game.avg_score > 0
                      ? game.avg_score.toFixed(1)
                      : t("game.panel.none")
                }}
              </span>
            </span>

            <span class="search-page-comment">{{ game.remark_count }} {{t("game.search.comment")}}</span>
          </div>
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

    <GameReviewDialog
        :visible="!!reviewGame"
        :game-id="reviewGame?.id ?? ''"
        :game-name="reviewGame?.name ?? ''"
        @close="reviewGame = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import GamePagination from '@/components/game/search/GamePagination.vue'
import GameReviewDialog from '@/components/game/common/GameReviewDialog.vue'
import type { SearchPageResponseItem } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

const router = useRouter()
const localePath = useLocalePath()

const steamPrefix = import.meta.env.VITE_STEAM_APP_PREFIX_URL || ''
const reviewGame = ref<SearchPageResponseItem | null>(null)

const goDetail = (id: number | string) => {
  router.push(localePath(`/games/${id}`))
}

const openReview = (game: SearchPageResponseItem) => {
  reviewGame.value = game
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
  container: search-card / inline-size;
  cursor: pointer;
}

.search-result-grid-shell {
  container: game-search-results / inline-size;
}

.search-result-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

@container game-search-results (min-width: 42rem) {
  .search-result-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@container game-search-results (min-width: 56rem) {
  .search-result-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@container game-search-results (min-width: 94rem) {
  .search-result-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

.search-page-title {
  min-width: 0;
  max-width: 100%;
  flex: 1 1 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-page-title-row {
  display: flex;
  min-width: 0;
  overflow: hidden;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  font-size: 0.875rem;
}

.search-page-title-wrap {
  display: flex;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  flex: 1 1 auto;
  align-items: center;
  gap: 0.38rem;
}

.search-page-tag {
  display: inline-flex;
  max-width: min(9.25rem, 48%);
  min-width: 0;
  min-height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@container search-card (max-width: 27rem) {
  .search-page-tag--secondary {
    display: none;
  }
}

@container search-card (max-width: 14.25rem) {
  .search-page-tag--primary,
  .search-page-tag--secondary {
    display: none;
  }
}

.search-page-desc {
  margin-top: 0.25rem;
  display: -webkit-box;
  height: 2.2rem;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.search-page-meta {
  margin-top: 0.62rem;
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
}

.search-page-score {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.search-page-score-value {
  min-width: 1.45rem;
}

</style>
