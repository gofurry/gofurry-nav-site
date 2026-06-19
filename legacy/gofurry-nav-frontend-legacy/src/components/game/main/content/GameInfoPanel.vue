<template>
  <div class="bg-white/50 backdrop-blur-md rounded-2xl shadow p-5 mb-8">
    <GameInfoGroup
        v-if="firstGroup"
        :group="firstGroup"
        :key="firstGroup.title"
        @more="handleMore"
    />

    <GameStatsPanels
        v-if="panelData"
        :topPriceList="panelData.top_price_vo"
        :discountList="panelData.top_discount_vo"
        :topCountList="panelData.top_count"
        :bottomPriceList="panelData.bottom_price"
    />

    <GameInfoGroup
        v-for="g in middleGroups"
        :key="g.title"
        :group="g"
        @more="handleMore"
    />

    <GameUpdateNews />

    <GameInfoGroup
        v-if="lastGroup"
        :group="lastGroup"
        :key="lastGroup.title"
        @more="handleMore"
    />
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, watch, computed} from "vue";
import GameInfoGroup from "@/components/game/main/content/GameInfoGroup.vue";
import GameStatsPanels from "@/components/game/main/content/GameStatsPanels.vue";
import { useLangStore } from '@/store/langStore.ts'
import { getGameMainInfo, getGameMainPanel } from "@/utils/api/game.ts"
import type { GameGroupRecord, BaseGameInfoRecord, GamePanelRecord } from "@/types/game.ts"
import GameUpdateNews from "@/components/game/main/content/GameUpdateNews.vue";

const langStore = useLangStore()
const lang = computed(() => langStore.lang)

const firstGroup = computed(() => groups.value[0] || null)
const middleGroups = computed(() => groups.value.slice(1, groups.value.length - 1))
const lastGroup = computed(() => groups.value.length > 1 ? groups.value[groups.value.length - 1] : null)

// 后端原始数据
const rawData = ref<GameGroupRecord | null>(null);

// 渲染用的数据
const groups = ref<
    {
      title: string;
      games: {
        id: string;
        name: string;
        cover: string;
        desc: string;
        score: number;
        scoreCount: number;
      }[];
    }[]
>([]);

function mapGames(list: BaseGameInfoRecord[], lang: string) {
  return list.map(g => ({
    id: g.game_id,
    name: lang === "en" ? g.name_en : g.name,
    cover: g.header,
    desc: lang === "en" ? g.info_en : g.info,
    score: g.avg_score,
    scoreCount: g.comment_count
  }))
}

function updateGroups() {
  if (!rawData.value) return;

  const r = rawData.value;
  const currentLang = lang.value;

  groups.value = [
    {
      title: currentLang === "en" ? "Latest Release" : "最近发售",
      games: mapGames(r.latest, currentLang)
    },
    {
      title: currentLang === "en" ? "Recently Added" : "最近收录",
      games: mapGames(r.recent, currentLang)
    },
    {
      title: currentLang === "en" ? "Free to Play" : "免费专区",
      games: mapGames(r.free, currentLang)
    },
    {
      title: currentLang === "en" ? "Hot Ranking" : "热门排行",
      games: mapGames(r.hot, currentLang)
    }
  ];
}

const panelData = ref<GamePanelRecord | null>(null);

onMounted(async () => {
  const groupRes = await getGameMainInfo();
  rawData.value = groupRes;
  updateGroups();

  panelData.value = await getGameMainPanel();
});

// 监听语言变化
watch(() => langStore.lang, () => {
  updateGroups();
});

const handleMore = (group: any) => {
  console.log("查看更多:", group.title);
};
</script>

