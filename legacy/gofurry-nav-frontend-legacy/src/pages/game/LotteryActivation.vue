<template>
  <div class="flex min-h-[60vh] items-center justify-center bg-orange-50">
    <div class="bg-orange-100 rounded-lg py-20 w-[80%] text-center space-y-4">

      <div v-if="status === 'success'" class="text-green-600 font-bold text-xl">
        馃惒 {{ t('game.lottery.activation.success') }}
      </div>

      <div v-else class="text-red-600 font-bold text-xl">
        馃惀 {{ t('game.lottery.activation.fail') }}
      </div>

      <p class="text-sm text-gray-600">
        {{ message }}
      </p>

      <p class="text-xs text-gray-500 mt-2">
        {{ countdown }} {{ t('game.lottery.activation.autoReturnIn') }}
      </p>

      <router-link to="/games/prize">
        <div class="inline-block mt-4 px-4 py-2 bg-orange-300
             text-orange-800 rounded-lg hover:bg-orange-400">
          {{ t('game.lottery.activation.returnNow') }}
        </div>
      </router-link>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const route = useRoute()
const router = useRouter()

const status = route.query.status as string
const message = (route.query.msg as string) || ''

const countdown = ref(15)
let timer: number | null = null

onMounted(() => {
  timer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      router.push('/games/prize')
    }
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
