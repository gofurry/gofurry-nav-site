<template>
  <div class="space-y-6">
    <!-- 快捷功能 -->
    <div>
      <h3 class="text-sm font-semibold text-gray-700 mb-2">
        {{ t("game.action.shortcutFunction") }}
      </h3>

      <div class="grid grid-cols-2 gap-3">
        <button
            class="flex items-center justify-center gap-1
                 bg-orange-100 hover:bg-orange-200/50
                 text-orange-900
                 py-2 rounded-lg
                 transition disabled:opacity-60"
            :disabled="loading"
            @click="handleRandomGame"
        >
          {{ t("game.action.dailyGame") }}
        </button>

        <button
            class="flex items-center justify-center gap-1
                 bg-orange-100 hover:bg-orange-200/50
                 text-orange-900
                 py-2 rounded-lg
                 transition"
            @click="openInstallModal"
        >
          {{ t("game.action.oneClickAddToLibrary") }}
        </button>

        <button
            class="flex items-center justify-center gap-1
                 bg-orange-100 hover:bg-orange-200/50
                 text-orange-900
                 py-2 rounded-lg
                 transition"
            @click="router.push('/games/creator')"
        >
          {{ t("game.action.authorList") }}
        </button>

        <button
            class="flex items-center justify-center gap-1
                 bg-orange-100 hover:bg-orange-200/50
                 text-orange-900
                 py-2 rounded-lg
                 transition"
            @click="router.push('/games/news/more')"
        >
          {{ t("game.action.moreNews") }}
        </button>

      </div>
    </div>

    <!-- 一键入库弹窗 -->
    <div
        v-if="showInstallModal"
        class="fixed inset-0 z-50 flex items-center justify-center pointer-events-none"
    >
      <div
          class="relative pointer-events-auto
               w-[360px]
               bg-orange-50 backdrop-blur-md
               rounded-xl shadow-lg p-5 space-y-4"
      >
        <button
            class="absolute top-3 right-3
                 text-gray-500 hover:text-gray-800 transition"
            @click="closeInstallModal"
        >
          ✕
        </button>

        <h3 class="text-lg font-semibold text-gray-800 mb-1">
          {{ t("game.action.steamOneClickAddRun") }}
        </h3>

        <!-- Install -->
        <div class="space-y-2">
          <p class="text-sm text-gray-600">{{ t("game.action.inputAppIDInstallFreeGame") }}</p>
          <input
              v-model="steamAppId"
              type="text"
              :placeholder="t('game.action.installTip')"
              class="w-full px-3 py-2 rounded-lg border border-orange-200
                   focus:outline-none focus:ring-2 focus:ring-orange-300 text-sm"
          />
          <button
              class="w-full py-2 rounded-lg text-sm font-medium
                   transition bg-orange-200 text-orange-900
                   hover:bg-orange-300 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="!steamAppId"
              @click="handleInstall"
          >
            {{ t("game.action.startInstall") }}
          </button>
        </div>

        <!-- 热门免费游戏 -->
        <div class="flex flex-wrap gap-2 py-2 border-y border-orange-200">
          <span
              v-for="game in hotGames"
              :key="game.id"
              class="px-2 py-1 bg-orange-100 rounded-md text-sm text-orange-900 cursor-pointer hover:bg-orange-200"
              @click="steamAppId = game.id"
          >
            {{ game.name }}
          </span>
        </div>

        <!-- Run / Connect -->
        <div class="space-y-2">
          <p class="text-sm text-gray-600">{{ t("game.action.inputContentRunConnectServer") }}</p>
          <input
              v-model="steamRunId"
              type="text"
              :placeholder="t('game.action.runTip')"
              class="w-full px-3 py-2 rounded-lg border border-orange-200
                   focus:outline-none focus:ring-2 focus:ring-orange-300 text-sm"
          />
          <div class="grid grid-cols-2 gap-2">
            <button
                class="py-2 rounded-lg text-sm font-medium
                     transition bg-orange-200 text-orange-900
                     hover:bg-orange-300 disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="!steamRunId"
                @click="handleRun"
            >
              {{ t("game.action.runGame") }}
            </button>
            <button
                class="py-2 rounded-lg text-sm font-medium
                     transition bg-orange-200 text-orange-900
                     hover:bg-orange-300 disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="!steamRunId"
                @click="handleConnect"
            >
              {{ t("game.action.connectServer") }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 相关网站 -->
    <div>
      <h3 class="text-sm font-semibold text-gray-700 mb-2">
        {{ t("game.action.relatedWebsites") }}
      </h3>

      <div class="flex flex-col gap-2 text-sm">
        <!-- 默认网站 -->
        <slot name="default-sites" />

        <!-- 可展开更多网站 -->
        <transition
            enter-active-class="transition-all duration-300 ease-out"
            leave-active-class="transition-all duration-200 ease-in"
            enter-from-class="opacity-0 max-h-0"
            enter-to-class="opacity-100 max-h-[2000px]"
            leave-from-class="opacity-100 max-h-[2000px]"
            leave-to-class="opacity-0 max-h-0"
        >
          <div v-show="showAllSites" class="flex flex-col gap-2 overflow-hidden">
            <slot name="extra-sites" />
          </div>
        </transition>

        <!-- 展开 / 收起按钮 -->
        <button
            class="mx-auto mt-2 p-1 rounded-md
                 text-xs text-orange-800
                 hover:bg-orange-200 hover:text-orange-700
                 transition"
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
import { useRouter } from "vue-router";
import { getRandomGame } from "@/utils/api/game.ts";
import { i18n } from "@/main.ts";

const { t } = i18n.global;
const router = useRouter();
const loading = ref(false);

const showInstallModal = ref(false);
const steamAppId = ref("");
const steamRunId = ref("");
const showAllSites = ref(false);

const hotGames = ref([
  { name: "家有大猫", id: "570840" },
  { name: "FarmD", id: "1814630" },
  { name: "Arctic Wolves", id: "3326280" },
  { name: "外兽祭", id: "1590570" },
  { name: "椰城蓝调", id: "2670920" },
  { name: "Cumdy", id: "2697060" },
  { name: "The Lar", id: "1084570" },
  { name: "绿洲计划", id: "1168840" },
  { name: "Neglected", id: "1300360" },
  { name: "Illusion", id: "1875610" },
]);

async function handleRandomGame() {
  if (loading.value) return;
  try {
    loading.value = true;
    const gameId = await getRandomGame();
    if (gameId) router.push(`/games/${gameId}`);
  } catch (err) {
    console.error("获取随机游戏失败", err);
  } finally {
    loading.value = false;
  }
}

function openInstallModal() {
  showInstallModal.value = true;
}

function closeInstallModal() {
  showInstallModal.value = false;
  steamAppId.value = "";
  steamRunId.value = "";
}

function handleInstall() {
  if (!steamAppId.value) return;
  const appId = steamAppId.value.trim();
  window.location.href = `steam://install/${appId}`;
}

function handleRun() {
  if (!steamRunId.value) return;
  const id = steamRunId.value.trim();
  window.location.href = `steam://run/${id}`;
}

function handleConnect() {
  if (!steamRunId.value) return;
  const ip = steamRunId.value.trim();
  window.location.href = `steam://connect/${ip}`;
}
</script>
