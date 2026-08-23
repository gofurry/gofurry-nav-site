<template>
  <a
      :href="href"
      target="_blank"
      class="game-site-item"
  >
    <div class="flex items-center gap-2 min-w-0">
      <img :src="icon" class="w-5 h-5 rounded" :alt="title" />
      <span class="game-site-item__title truncate font-medium">{{ title }}</span>
    </div>
    <span class="game-site-item__subtitle truncate text-xs">{{ subtitleText }}</span>
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
