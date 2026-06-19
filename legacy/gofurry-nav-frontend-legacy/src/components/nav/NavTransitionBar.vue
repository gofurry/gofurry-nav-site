<template>
  <div class="relative z-10">
    <div class="mx-auto flex items-center gap-4 border-t-4 border-black/30 bg-[rgba(18,24,37,0.55)] px-4 py-2 text-sm text-gray-100 shadow-lg ring-1 ring-white/10 md:gap-6 md:px-6">
      <div class="flex items-center font-semibold justify-between text-sm text-gray-300">
        {{ formattedDateTime }}
      </div>

      <div class="flex flex-col">
        <div class="w-full h-1 textgray"></div>
        <iframe
            class="w-[230px] h-8"
            allowtransparency="true"
            src="https://i.tianqi.com/index.php?c=code&id=73&icon=1&num=3&color=d1d5dc"
        ></iframe>
      </div>

      <div v-if="saying" class="hidden min-w-0 flex-1 sm:block">
        <div class="group relative flex justify-end">
          <div class="absolute bottom-full mb-3 opacity-0 transition duration-200 group-hover:translate-y-0 group-hover:opacity-100">
            <div class="rounded-lg border border-white/15 bg-[rgba(18,24,37,0.7)] px-3 py-2 text-xs text-gray-200 ring-1 ring-white/10">
              {{ quoteAuthor }}
            </div>
          </div>
          <div>
            {{ quoteDisplay }}
          </div>
        </div>
      </div>
      <div v-else>
        你的恩情狼不会忘记
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSaying } from '@/utils/api/nav'
import type { SayingModel } from '@/types/nav.ts'

const { locale } = useI18n()

const formattedDateTime = ref('')
const saying = ref<SayingModel | null>(null)

const quoteDisplay = computed(() => `“${saying.value?.content ?? ''}”`)
const quoteAuthor = computed(() => {
  const author = saying.value?.author?.trim()
  if (author) {
    return author
  }
  return locale.value === 'zh' ? '佚名' : 'Unknown'
})

const updateTime = () => {
  const formatLocale = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  formattedDateTime.value = new Intl.DateTimeFormat(formatLocale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date())
}

let timeTimer: number | null = null

watch(locale, () => {
  updateTime()
})

onMounted(async () => {
  updateTime()
  timeTimer = window.setInterval(updateTime, 60 * 1000)
  saying.value = await getSaying()
})

onUnmounted(() => {
  if (timeTimer) {
    window.clearInterval(timeTimer)
  }
})
</script>

<style scoped>
</style>
