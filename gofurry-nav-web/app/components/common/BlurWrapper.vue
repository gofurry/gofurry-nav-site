<template>
  <div class="blur-wrapper relative">

    <!-- 内容区域 -->
    <div
        class="blur-wrapper__content transition"
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
      <div class="blur-wrapper__notice text-sm px-4 py-2">
        {{ tip }}
      </div>

      <!-- 点击解锁 -->
      <button
          class="blur-wrapper__unlock flex items-center justify-center gap-1 px-5 py-2 text-sm transition pointer-events-auto"
          @click.stop="emit('unlock')"
      >
        <img :src="key" class="w-4 h-4" alt="" />
        <span>{{t("common.unlock")}}</span>
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import key from '@/assets/svgs/key.svg'
import { i18n } from '@/main'

const { t } = i18n.global

const emit = defineEmits<{
  (e: 'unlock'): void
}>()

defineProps<{
  enable: boolean
  tip: string
}>()
</script>
