<template>
  <a
      :href="href"
      target="_blank"
      class="game-site-item"
  >
    <div class="flex items-center gap-2 min-w-0">
      <img :src="icon" class="w-5 h-5 rounded" :alt="title" />
      <span class="font-medium text-gray-800 truncate">{{ title }}</span>
    </div>
    <span class="text-xs text-gray-500 truncate">{{ subtitleText }}</span>
  </a>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { i18n } from "@/main";

const { t } = i18n.global;

const props = defineProps<{
  href: string;
  icon: string;
  title: string;
  subtitle: string;
}>();

const subtitleText = computed(() => t(`game.site.${props.subtitle}`));
</script>

<style scoped>
.game-site-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid rgba(126, 92, 58, 0.16);
  border-radius: 0.72rem;
  background: rgba(255, 250, 242, 0.40);
  box-shadow: 0 4px 12px rgba(91, 62, 28, 0.03);
  padding: 0.48rem 0.62rem;
  transition: background-color 180ms ease, border-color 180ms ease;
}

.game-site-item:hover {
  border-color: rgba(180, 96, 24, 0.34);
  background: rgba(255, 239, 213, 0.68);
}

:global(.dark) .game-site-item {
  border-color: rgba(226, 232, 240, 0.15);
  background: rgba(226, 232, 240, 0.065);
  box-shadow: none;
}

:global(.dark) .game-site-item:hover {
  border-color: rgba(148, 163, 184, 0.36);
  background: rgba(148, 163, 184, 0.11);
}

:global(.dark) .game-site-item .text-gray-800 {
  color: rgba(241, 245, 249, 0.86);
}

:global(.dark) .game-site-item .text-gray-500 {
  color: rgba(203, 213, 225, 0.66);
}

:global(.games-page--dark) .game-site-item:hover {
  border-color: rgba(148, 163, 184, 0.36) !important;
  background: rgba(148, 163, 184, 0.11) !important;
}
</style>
