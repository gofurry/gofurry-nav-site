<template>
  <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4 py-6 backdrop-blur-md"
  >
    <div
        class="relative max-h-[calc(100vh-3rem)] w-full max-w-2xl overflow-y-auto rounded-xl border border-white/10 bg-[rgba(21,20,18,0.92)] p-5 text-stone-100 shadow-[0_24px_90px_rgba(0,0,0,0.45)] ring-1 ring-white/5 backdrop-blur-xl sm:p-6"
    >
      <div class="absolute inset-x-6 top-0 h-px bg-gradient-to-r from-transparent via-orange-200/60 to-transparent" aria-hidden="true" />

      <div class="mb-4 flex items-start justify-between gap-4">
        <h3 class="text-xl font-semibold leading-7 text-white">
          {{ lottery.lottery.title }}
        </h3>
        <button
            type="button"
            class="grid size-8 shrink-0 place-items-center rounded-lg border border-white/10 text-stone-400 transition hover:border-white/20 hover:bg-white/[0.08] hover:text-white"
            :aria-label="t('common.cancel')"
            @click="emit('close')"
        >
          ×
        </button>
      </div>

      <p class="mb-5 text-sm leading-6 text-stone-300">
        {{ lottery.lottery.desc }}
      </p>

      <div class="mb-5 grid gap-3 rounded-lg border border-white/10 bg-white/[0.04] p-4 text-sm text-stone-300">
        <div class="flex items-center justify-between gap-4">
          <div class="text-stone-500">{{ t('game.lottery.home.prize') }}</div>
          <div>
            {{ lottery.lottery.prize.title }}
            ({{ lottery.lottery.prize.platform }})
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="rounded-lg bg-black/20 px-3 py-2">
            <div class="text-[11px] text-stone-500">{{ t('game.lottery.home.prizeQuantity') }}</div>
            <div>
              {{ lottery.lottery.prize.count }}
            </div>
          </div>

          <div class="rounded-lg bg-black/20 px-3 py-2">
            <div class="text-[11px] text-stone-500">{{ t('game.lottery.home.participants') }}</div>
            <div>
              {{ lottery.count }}
            </div>
          </div>
        </div>
      </div>

      <div class="mb-4">
        <div class="mb-2 text-sm font-medium text-stone-300">
          {{ t('game.lottery.submitModal.currentParticipants') }}
        </div>

        <div
            v-if="!lottery.member.length"
            class="text-sm text-stone-500"
        >
          {{ t('game.lottery.submitModal.noParticipants') }}
        </div>

        <div
            v-else
            class="flex flex-wrap gap-2"
        >
          <span
              v-for="m in visibleMembers"
              :key="m.email"
              class="rounded-full border border-orange-200/16 bg-orange-200/10 px-3 py-1 text-xs text-orange-100/90"
          >
            {{ m.name }} - {{ m.email }}
          </span>
        </div>

        <div v-if="!allLoaded" class="mt-3 flex items-center justify-start">
          <button
              type="button"
              @click="loadMore"
              class="rounded-lg border border-white/10 bg-white/[0.06] px-2.5 py-1 text-xs text-stone-300 transition hover:bg-white/[0.1]"
          >
            {{ t('common.loadMore') }}
          </button>
        </div>

        <div
            v-else-if="lottery.member.length > 5"
            class="mt-2 text-xs text-stone-500"
        >
          {{ t('game.lottery.submitModal.allLoaded') }}
        </div>
      </div>

      <div class="space-y-3">
        <input
            v-model="keyInput"
            :placeholder="t('game.lottery.submitModal.enterLotteryKey')"
            class="w-full rounded-lg border border-white/10 bg-black/[0.24] px-3 py-2.5 text-sm text-stone-100 outline-none transition placeholder:text-stone-600 focus:border-orange-200/50"
        />

        <input
            v-model="nameInput"
            :placeholder="t('game.lottery.submitModal.enterName')"
            class="w-full rounded-lg border border-white/10 bg-black/[0.24] px-3 py-2.5 text-sm text-stone-100 outline-none transition placeholder:text-stone-600 focus:border-orange-200/50"
        />

        <input
            v-model="emailInput"
            :placeholder="t('game.lottery.submitModal.enterEmail')"
            class="w-full rounded-lg border border-white/10 bg-black/[0.24] px-3 py-2.5 text-sm text-stone-100 outline-none transition placeholder:text-stone-600 focus:border-orange-200/50"
        />

        <div v-if="emailError" class="text-xs text-red-300">
          {{ emailError }}
        </div>

        <div v-if="submitError" class="text-xs text-red-300">
          {{ submitError }}
        </div>

        <div v-if="successMsg" class="text-xs text-emerald-300">
          {{ successMsg }}
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <button
            @click="emit('close')"
            class="rounded-lg border border-white/10 px-4 py-2 text-sm text-stone-300 transition hover:bg-white/[0.08] hover:text-white"
        >
          {{ t('common.cancel') }}
        </button>

        <button
            @click="submit"
            :disabled="loading"
            class="rounded-lg bg-orange-200 px-4 py-2 text-sm font-medium text-stone-950 transition hover:bg-orange-100 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {{ loading ? t("common.commiting") : t("common.commit") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue"
import { getLotteryParticipation } from "@/utils/api/game"
import type { LotteryActiveModel } from "@/types/game"
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  lottery: LotteryActiveModel
}>()

const emit = defineEmits(["close"])

const visibleCount = ref(5)
const loading = ref(false)

const keyInput = ref("")
const nameInput = ref("")
const emailInput = ref("")

const emailError = ref("")
const submitError = ref("")
const successMsg = ref("")

const totalMembers = computed(() => props.lottery.member.length)

const visibleMembers = computed(() =>
    props.lottery.member.slice(0, visibleCount.value)
)

function loadMore() {
  visibleCount.value += 5
}

const allLoaded = computed(() =>
    visibleCount.value >= totalMembers.value
)

function validateEmail(email: string) {
  const reg =
      /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/
  return reg.test(email)
}

async function submit() {
  submitError.value = ""
  emailError.value = ""
  successMsg.value = ""

  if (!keyInput.value || !nameInput.value || !emailInput.value) {
    submitError.value = t('game.lottery.submitModal.fillAllInfo')
    return
  }

  if (!validateEmail(emailInput.value)) {
    emailError.value = t('game.lottery.submitModal.invalidEmail')
    return
  }

  try {
    loading.value = true

    const req = {
      id: Number(props.lottery.lottery.id),
      name: nameInput.value.trim(),
      email: emailInput.value.trim(),
      key: keyInput.value.trim()
    }

    const res = await getLotteryParticipation(req)

    // 根据 code 判断
    if (res.code === 1) {
      successMsg.value = t('game.lottery.submitModal.successEmailSent')

      keyInput.value = ""
      nameInput.value = ""
      emailInput.value = ""

    } else {
      // 失败时 data 是错误信息
      submitError.value = res.data || t('game.lottery.submitModal.submitFail')
    }

  } catch (e) {
    submitError.value = t('game.lottery.submitModal.networkError')
  } finally {
    loading.value = false
  }
}
</script>
