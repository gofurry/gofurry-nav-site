<template>
  <div class="space-y-6">
    <!-- 快捷功能 -->
    <div>
      <h3 class="sidebar-section-title">
        {{ t("game.action.shortcutFunction") }}
      </h3>

      <div class="grid grid-cols-2 gap-3">
        <button
            class="sidebar-action-button"
            :disabled="loading"
            @click="handleRandomGame"
        >
          {{ t("game.action.dailyGame") }}
        </button>

        <button
            class="sidebar-action-button"
            type="button"
        >
          Steam 专区
        </button>
      </div>
    </div>

    <!-- 相关网站 -->
    <div>
      <h3 class="sidebar-section-title">
        {{ t("game.action.relatedWebsites") }}
      </h3>

      <div class="flex flex-col gap-2 text-sm">
        <!-- 默认网站 -->
        <slot name="default-sites" />

        <!-- 可展开更多网站 -->
        <transition
            enter-active-class="sidebar-sites-expand-enter-active"
            leave-active-class="sidebar-sites-expand-leave-active"
            enter-from-class="sidebar-sites-expand-enter-from"
            enter-to-class="sidebar-sites-expand-enter-to"
            leave-from-class="sidebar-sites-expand-leave-from"
            leave-to-class="sidebar-sites-expand-leave-to"
        >
          <div v-show="showAllSites" class="flex flex-col gap-2 overflow-hidden">
            <slot name="extra-sites" />
          </div>
        </transition>

        <!-- 展开 / 收起按钮 -->
        <button
            class="sidebar-expand-button"
            @click="showAllSites = !showAllSites"
        >
          <span v-if="!showAllSites">{{ t("common.expand") }} ▼</span>
          <span v-else>{{ t("common.collapse") }} ▲</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { getRandomGame } from "@/utils/api/game";
import { i18n } from "@/main";

const { t } = i18n.global;
const router = useRouter();
const localePath = useLocalePath();
const loading = ref(false);

const showAllSites = ref(false);

async function handleRandomGame() {
  if (loading.value) return;
  try {
    loading.value = true;
    const gameId = await getRandomGame();
    if (gameId) router.push(localePath(`/games/${gameId}`));
  } catch (err) {
    console.error("获取随机游戏失败", err);
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.sidebar-section-title {
  margin-bottom: 0.55rem;
  color: rgba(55, 43, 32, 0.82);
  font-size: 0.84rem;
  font-weight: 700;
}

.sidebar-action-button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 2.45rem;
  border: 1px solid rgba(126, 92, 58, 0.17);
  border-radius: 0.75rem;
  background: rgba(255, 250, 242, 0.40);
  box-shadow: 0 4px 12px rgba(91, 62, 28, 0.035);
  color: rgba(124, 45, 18, 0.88);
  font-size: 0.9rem;
  font-weight: 650;
  transition: background-color 180ms ease, border-color 180ms ease, color 180ms ease;
}

.sidebar-action-button:hover {
  border-color: rgba(180, 96, 24, 0.34);
  background: rgba(255, 239, 213, 0.68);
  color: rgba(99, 39, 15, 0.96);
}

.sidebar-action-button:disabled {
  cursor: default;
  opacity: 0.55;
}

.sidebar-expand-button {
  margin: 0.55rem auto 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: rgba(154, 52, 18, 0.78);
  font-size: 0.8rem;
  font-weight: 650;
  transition: color 180ms ease;
}

.sidebar-expand-button:hover {
  color: rgba(124, 45, 18, 0.95);
}

.sidebar-sites-expand-enter-active,
.sidebar-sites-expand-leave-active {
  overflow: hidden;
  transition:
    max-height 460ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 300ms ease,
    transform 460ms cubic-bezier(0.22, 1, 0.36, 1);
}

.sidebar-sites-expand-enter-from,
.sidebar-sites-expand-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-0.35rem);
}

.sidebar-sites-expand-enter-to,
.sidebar-sites-expand-leave-from {
  max-height: 36rem;
  opacity: 1;
  transform: translateY(0);
}

:global(.dark) .sidebar-section-title {
  color: rgba(226, 232, 240, 0.82);
}

:global(.dark) .sidebar-action-button {
  border-color: rgba(226, 232, 240, 0.15);
  background: rgba(226, 232, 240, 0.065);
  box-shadow: none;
  color: rgba(180, 213, 226, 0.70);
}

:global(.dark) .sidebar-action-button:hover {
  border-color: rgba(148, 163, 184, 0.36);
  background: rgba(148, 163, 184, 0.11);
  color: rgba(226, 232, 240, 0.86);
}

:global(.dark) .sidebar-expand-button {
  color: rgba(180, 213, 226, 0.66);
}

:global(.dark) .sidebar-expand-button:hover {
  color: rgba(226, 232, 240, 0.84);
  background: transparent;
}

:global(.games-page--dark) .sidebar-action-button {
  border-color: rgba(226, 232, 240, 0.15);
  background: rgba(226, 232, 240, 0.065);
  color: rgba(180, 213, 226, 0.70) !important;
}

:global(.games-page--dark) .sidebar-action-button:hover {
  border-color: rgba(148, 163, 184, 0.36) !important;
  background: rgba(148, 163, 184, 0.11) !important;
  color: rgba(226, 232, 240, 0.86) !important;
}

:global(.games-page--dark) .sidebar-expand-button {
  color: rgba(180, 213, 226, 0.66) !important;
}

:global(.games-page--dark) .sidebar-expand-button:hover {
  color: rgba(226, 232, 240, 0.84) !important;
  background: transparent !important;
}
</style>
