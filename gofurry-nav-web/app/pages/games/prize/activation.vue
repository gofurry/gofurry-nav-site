<template>
  <div class="lottery-activation-page relative isolate flex min-h-[calc(100svh-3.5rem)] items-center justify-center overflow-hidden px-4 py-8">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="lottery-activation-page__wash absolute inset-0 z-0" aria-hidden="true" />

    <main class="activation-card relative z-10 w-full max-w-xl">
      <div
        class="activation-status"
        :class="isSuccess ? 'activation-status--success' : 'activation-status--fail'"
        aria-hidden="true"
      >
        <svg v-if="isSuccess" viewBox="0 0 24 24">
          <path d="M5 12.5 10 17l9-10" />
        </svg>
        <svg v-else viewBox="0 0 24 24">
          <path d="M7 7 17 17M17 7 7 17" />
        </svg>
      </div>

      <p class="activation-card__eyebrow text-xs font-semibold uppercase tracking-[0.28em]">
        GoFurry Lottery
      </p>
      <h1
        class="activation-card__title mt-4 text-3xl font-semibold tracking-normal sm:text-4xl"
        :class="isSuccess ? 'activation-card__title--success' : 'activation-card__title--fail'"
      >
        {{ title }}
      </h1>

      <p class="activation-card__message mt-4 text-sm leading-7">
        {{ displayMessage }}
      </p>

      <div class="activation-card__countdown mt-7 rounded-lg px-4 py-3 text-sm">
        <span class="font-semibold">{{ countdown }}</span>
        {{ t('game.lottery.activation.autoReturnIn') }}
      </div>

      <div class="mt-7 flex flex-wrap gap-3">
        <RouterLink
          to="/games/prize"
          class="activation-card__link inline-flex min-h-11 items-center justify-center rounded-lg px-4 text-sm font-semibold transition"
        >
          {{ t('game.lottery.activation.returnNow') }}
        </RouterLink>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { i18n } from '@/main'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const { t } = i18n.global

const route = useRoute()
const router = useRouter()

useHead({
  meta: [
    { name: 'robots', content: 'noindex, follow' }
  ]
})

const status = computed(() => String(route.query.status || ''))
const message = computed(() => String(route.query.msg || ''))
const isSuccess = computed(() => status.value === 'success')
const title = computed(() => isSuccess.value
  ? t('game.lottery.activation.success')
  : t('game.lottery.activation.fail')
)
const displayMessage = computed(() => message.value || (isSuccess.value
  ? t('game.lottery.activation.defaultSuccessMessage')
  : t('game.lottery.activation.defaultFailMessage')
))

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
  if (timer) {
    clearInterval(timer)
  }
})
</script>
