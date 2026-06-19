<template>
  <div class="relative">

    <!-- 内容区域 -->
    <div
        class="max-h-[960px] overflow-y-auto transition"
        :class="enable ? 'blur-xl select-none' : ''"
        :style="enable ? 'pointer-events: none;' : ''"
    >
      <slot />
    </div>

    <!-- 遮罩 -->
    <div
        v-if="enable"
        class="absolute inset-0 flex flex-col items-center justify-center gap-4"
    >
      <div class="bg-black/60 text-gray-200 text-sm px-4 py-2 rounded-xl backdrop-blur">
        {{ tip }}
      </div>

      <!-- 点击解锁 -->
      <button
          class="flex items-center justify-center gap-1 px-5 py-2 text-sm rounded-full bg-orange-900 text-gray-200 hover:bg-orange-800 transition pointer-events-auto"
          @click.stop="emit('unlock')"
      >
        <img :src="key" class="w-4 h-4" alt="flag" />
        <span>{{t("common.unlock")}}</span>
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import key from '@/assets/svgs/key.svg'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const emit = defineEmits<{
  (e: 'unlock'): void
}>()

defineProps<{
  enable: boolean
  tip: string
}>()
</script>
