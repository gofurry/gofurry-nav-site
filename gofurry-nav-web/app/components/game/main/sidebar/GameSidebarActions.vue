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
