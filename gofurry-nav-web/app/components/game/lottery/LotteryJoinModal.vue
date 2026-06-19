<template>
  <div
      class="lottery-modal fixed inset-0 z-50 flex items-center justify-center px-4 py-6 backdrop-blur-md"
  >
    <div
        class="lottery-modal__dialog relative max-h-[calc(100vh-3rem)] w-full max-w-2xl overflow-y-auto rounded-xl p-5 backdrop-blur-xl sm:p-6"
    >
      <div class="lottery-modal__top-line absolute inset-x-6 top-0 h-px" aria-hidden="true" />

      <div class="mb-4 flex items-start justify-between gap-4">
        <h3 class="lottery-modal__title text-xl font-semibold leading-7">
          {{ lottery.lottery.title }}
        </h3>
        <button
            type="button"
            class="lottery-modal__close grid size-8 shrink-0 place-items-center rounded-lg transition"
            :aria-label="t('common.cancel')"
            @click="emit('close')"
        >
          ×
        </button>
      </div>

      <p class="lottery-modal__desc mb-5 text-sm leading-6">
        {{ lottery.lottery.desc }}
      </p>

      <div class="lottery-modal__summary mb-5 grid gap-3 rounded-lg p-4 text-sm">
        <div class="flex items-center justify-between gap-4">
          <div class="lottery-modal__label">{{ t('game.lottery.home.prize') }}</div>
          <div>
            {{ lottery.lottery.prize.title }}
            ({{ lottery.lottery.prize.platform }})
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="lottery-modal__stat rounded-lg px-3 py-2">
            <div class="lottery-modal__label text-[11px]">{{ t('game.lottery.home.prizeQuantity') }}</div>
            <div>
              {{ lottery.lottery.prize.count }}
            </div>
          </div>

          <div class="lottery-modal__stat rounded-lg px-3 py-2">
            <div class="lottery-modal__label text-[11px]">{{ t('game.lottery.home.participants') }}</div>
            <div>
              {{ lottery.count }}
            </div>
          </div>
        </div>
      </div>

      <div class="mb-4">
        <div class="lottery-modal__section-title mb-2 text-sm font-medium">
          {{ t('game.lottery.submitModal.currentParticipants') }}
        </div>

        <div
            v-if="!lottery.member.length"
            class="lottery-modal__empty text-sm"
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
              class="lottery-modal__chip rounded-full px-3 py-1 text-xs"
          >
            {{ m.name }} - {{ m.email }}
          </span>
        </div>

        <div v-if="!allLoaded" class="mt-3 flex items-center justify-start">
          <button
              type="button"
              @click="loadMore"
              class="lottery-modal__load-more rounded-lg px-2.5 py-1 text-xs transition"
          >
            {{ t('common.loadMore') }}
          </button>
        </div>

        <div
            v-else-if="lottery.member.length > 5"
            class="lottery-modal__empty mt-2 text-xs"
        >
          {{ t('game.lottery.submitModal.allLoaded') }}
        </div>
      </div>

      <div class="space-y-3">
        <input
            v-model="keyInput"
            :placeholder="t('game.lottery.submitModal.enterLotteryKey')"
            class="lottery-modal__input w-full rounded-lg px-3 py-2.5 text-sm outline-none transition"
        />

        <input
            v-model="nameInput"
            :placeholder="t('game.lottery.submitModal.enterName')"
            class="lottery-modal__input w-full rounded-lg px-3 py-2.5 text-sm outline-none transition"
        />

        <input
            v-model="emailInput"
            :placeholder="t('game.lottery.submitModal.enterEmail')"
            class="lottery-modal__input w-full rounded-lg px-3 py-2.5 text-sm outline-none transition"
        />

        <div v-if="emailError" class="lottery-modal__message lottery-modal__message--error text-xs">
          {{ emailError }}
        </div>

        <div v-if="submitError" class="lottery-modal__message lottery-modal__message--error text-xs">
          {{ submitError }}
        </div>

        <div v-if="successMsg" class="lottery-modal__message lottery-modal__message--success text-xs">
          {{ successMsg }}
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <button
            @click="emit('close')"
            class="lottery-modal__button lottery-modal__button--secondary rounded-lg px-4 py-2 text-sm transition"
        >
          {{ t('common.cancel') }}
        </button>

        <button
            @click="submit"
            :disabled="loading"
            class="lottery-modal__button lottery-modal__button--primary rounded-lg px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50"
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
