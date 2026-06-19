<template>
  <a
      :href="href"
      target="_blank"
      class="flex items-center justify-between gap-3
           px-3 py-2 rounded-lg
           bg-orange-50 hover:bg-orange-200/50 transition"
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
import { i18n } from "@/main.ts";

const { t } = i18n.global;

const props = defineProps<{
  href: string;
  icon: string;
  title: string;
  subtitle: string;
}>();

const subtitleText = computed(() => t(`game.site.${props.subtitle}`));
</script>
