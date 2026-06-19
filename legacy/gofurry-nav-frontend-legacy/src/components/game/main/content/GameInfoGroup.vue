<template>
  <div class="p-5 mb-8">

    <!-- 标题 -->
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-2xl font-bold text-gray-800">
        {{ group.title }}
      </h3>

      <router-link
          to="/games/search"
          class="text-md text-orange-900 hover:text-orange-700 transition cursor-pointer
         hover:bg-orange-200/50 p-2 rounded-md"
      >
        {{ t("common.showMore") }}
      </router-link>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div
          v-for="item in visibleGames"
          :key="item.id"
          class="cursor-pointer p-2 rounded-lg hover:bg-orange-200/50 transition"
      >
        <!-- 封面 -->
        <img
            :src="item.cover"
            class="w-full h-32 object-cover rounded-md mb-2"
            alt="封面图加载失败"
            @click.stop="goGameDetail(item.id)"
        />

        <!-- 标题 -->
        <p class="text-sm font-semibold text-gray-900 line-clamp-1">
          {{ item.name }}
        </p>

        <!-- 简介 -->
        <p class="text-xs text-gray-600 mt-1 overflow-hidden h-[2rem]">
          {{ item.desc }}
        </p>

        <!-- 评分 -->
        <div class="mt-2">
          <RatingStar :score="item.score" :count="item.scoreCount" />
        </div>
      </div>
    </div>


  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import RatingStar from "@/components/common/RatingStar.vue";
import { i18n } from '@/main.ts'

const { t } = i18n.global
const router = useRouter();

interface GameItem {
  id: string;
  name: string;
  cover: string;
  desc: string;
  score: number;
  scoreCount: number;
}

interface GameGroup {
  title: string;
  games: GameItem[];
}

const props = defineProps<{
  group: GameGroup;
}>();

defineEmits<{
  (e: "more", group: GameGroup): void;
}>();

function goGameDetail(id: string) {
  router.push(`/games/${id}`);
}

const screenWidth = ref(window.innerWidth);

const updateWidth = () => {
  screenWidth.value = window.innerWidth;
};

onMounted(() => {
  window.addEventListener("resize", updateWidth);
});

const visibleGames = computed(() => {
  if (screenWidth.value >= 1024) {
    return props.group.games.slice(0, 8);
  } else {
    return props.group.games.slice(0, 6);
  }
});
</script>
