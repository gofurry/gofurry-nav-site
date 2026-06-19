<template>
  <div
      class="fixed inset-0 bg-black/40 backdrop-blur-sm flex items-center justify-center z-50"
  >
    <div
        class="bg-orange-50 w-[60%] md:w-[80%] lg:w-[95%] max-w-2xl rounded-2xl shadow-xl p-6 relative"
    >
      <!-- 标题 -->
      <div class="flex justify-start items-center mb-4">
        <h3 class="text-lg font-bold text-orange-800">
          {{ lottery.lottery.title }}
        </h3>
      </div>

      <p class="text-gray-600 mb-4">
        {{ lottery.lottery.desc }}
      </p>

      <!-- 抽奖信息 -->
      <div class="text-sm text-gray-500 space-y-1 mb-1">

        <div class="flex">
          <div class="font-bold text-gray-600">{{ t('game.lottery.home.prize') }}:&nbsp;</div>
          <div>
            {{ lottery.lottery.prize.title }}
            ({{ lottery.lottery.prize.platform }})
          </div>
        </div>

        <div class="flex justify-between items-center">
          <div class="flex">
            <div class="font-bold text-gray-600">{{ t('game.lottery.home.prizeQuantity') }}:&nbsp;</div>
            <div>
              {{ lottery.lottery.prize.count }}
            </div>
          </div>

          <div class="flex">
            <div class="font-bold text-gray-600">{{ t('game.lottery.home.participants') }}:&nbsp;</div>
            <div>
              {{ lottery.count }}
            </div>
          </div>
        </div>
      </div>

      <!-- 参与者列表 -->
      <div class="mb-4">
        <div class="text-sm font-semibold mb-2 text-gray-600">
          {{ t('game.lottery.submitModal.currentParticipants') }}
        </div>

        <div
            v-if="!lottery.member.length"
            class="text-gray-400 text-sm"
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
              class="bg-orange-100 text-orange-700 text-xs px-3 py-1 rounded-full"
          >
            {{ m.name }} - {{ m.email }}
          </span>
        </div>

        <div v-if="!allLoaded" class="mt-3 flex items-center justify-start">
          <div
              @click="loadMore"
              class="bg-orange-200 py-1 px-2 rounded-lg text-xs text-orange-600 cursor-pointer"
          >
            {{ t('common.loadMore') }}
          </div>
        </div>

        <div
            v-else-if="lottery.member.length > 5"
            class="text-xs text-gray-400 mt-2"
        >
          {{ t('game.lottery.submitModal.allLoaded') }}
        </div>
      </div>

      <!-- 表单 -->
      <div class="space-y-3">
        <input
            v-model="keyInput"
            :placeholder="t('game.lottery.submitModal.enterLotteryKey')"
            class="w-full px-3 py-2 rounded-lg border-2 border-orange-200 focus:outline-none
             focus:border-orange-400 transition duration-300"
        />

        <input
            v-model="nameInput"
            :placeholder="t('game.lottery.submitModal.enterName')"
            class="w-full px-3 py-2 rounded-lg border-2 border-orange-200 focus:outline-none
             focus:border-orange-400 transition duration-300"
        />

        <input
            v-model="emailInput"
            :placeholder="t('game.lottery.submitModal.enterEmail')"
            class="w-full px-3 py-2 rounded-lg border-2 border-orange-200 focus:outline-none
             focus:border-orange-400 transition duration-300"
        />

        <div v-if="emailError" class="text-xs text-red-500">
          {{ emailError }}
        </div>

        <div v-if="submitError" class="text-xs text-red-500">
          {{ submitError }}
        </div>

        <div v-if="successMsg" class="text-xs text-green-600">
          {{ successMsg }}
        </div>
      </div>

      <!-- 按钮 -->
      <div class="flex justify-end gap-3 mt-6">
        <button
            @click="emit('close')"
            class="px-4 py-2 text-sm rounded-lg hover:bg-orange-100 text-gray-600"
        >
          {{ t('common.cancel') }}
        </button>

        <button
            @click="submit"
            :disabled="loading"
            class="px-4 py-2 text-sm rounded-lg bg-orange-400 text-white
                 hover:bg-orange-300 disabled:opacity-50"
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
import { i18n } from '@/main.ts'

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