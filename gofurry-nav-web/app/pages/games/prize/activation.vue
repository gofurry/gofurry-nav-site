<template>
  <div class="lottery-activation-page relative isolate flex min-h-[calc(100svh-3.5rem)] items-center justify-center overflow-hidden bg-[#11100f] px-4 py-8 text-stone-100">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="absolute inset-0 z-0 bg-[radial-gradient(circle_at_50%_25%,rgba(244,170,96,0.10),transparent_32%),linear-gradient(180deg,rgba(8,14,28,0.52),rgba(8,14,28,0.62))]" aria-hidden="true" />

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

      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-orange-100/65">
        GoFurry Lottery
      </p>
      <h1
        class="mt-4 text-3xl font-semibold tracking-normal sm:text-4xl"
        :class="isSuccess ? 'text-emerald-100' : 'text-orange-100'"
      >
        {{ title }}
      </h1>

      <p class="mt-4 text-sm leading-7 text-stone-300">
        {{ displayMessage }}
      </p>

      <div class="mt-7 rounded-lg border border-white/10 bg-white/[0.055] px-4 py-3 text-sm text-stone-300">
        <span class="font-semibold text-stone-100">{{ countdown }}</span>
        {{ t('game.lottery.activation.autoReturnIn') }}
      </div>

      <div class="mt-7 flex flex-wrap gap-3">
        <RouterLink
          to="/games/prize"
          class="inline-flex min-h-11 items-center justify-center rounded-lg border border-orange-200/24 bg-orange-300/18 px-4 text-sm font-semibold text-orange-50 transition hover:border-orange-200/42 hover:bg-orange-300/26"
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

<style scoped>
.activation-card {
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 1rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.04)),
    rgba(15, 23, 42, 0.62);
  box-shadow: 0 24px 70px rgba(2, 6, 23, 0.34);
  padding: clamp(1.5rem, 4vw, 2.5rem);
  backdrop-filter: blur(18px) saturate(1.12);
}

.activation-status {
  display: grid;
  width: 3rem;
  height: 3rem;
  place-items: center;
  border-radius: 999px;
  margin-bottom: 1.25rem;
}

.activation-status svg {
  width: 1.55rem;
  height: 1.55rem;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 3;
}

.activation-status--success {
  border: 1px solid rgba(74, 222, 128, 0.34);
  background: rgba(34, 197, 94, 0.16);
  color: #86efac;
}

.activation-status--fail {
  border: 1px solid rgba(251, 146, 60, 0.34);
  background: rgba(251, 146, 60, 0.16);
  color: #fdba74;
}
</style>
