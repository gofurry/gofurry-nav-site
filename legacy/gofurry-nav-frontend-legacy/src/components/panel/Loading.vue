<template>
  <div class="loading fixed inset-0 bg-[rgba(0,0,0,0.7)] z-50 flex items-center justify-center">
    <div class="loadbox text-center">
      <img src="https://qcdn.go-furry.com/nav/stat-bg/panel/loading.gif" :alt='t("common.loading")' class="w-16 h-16 mx-auto mb-3">
      <span class="text-white text-lg">{{t("common.loading")}}...</span>
    </div>
  </div>
</template>

<script setup>
import {i18n} from "@/main.ts";
const { t } = i18n.global
</script>

<style scoped></style>