<template>
  <Teleport to="body">
  <div
      class="lottery-modal fixed inset-0 z-[120] flex items-center justify-center px-4 py-6"
  >
    <div
        ref="panel"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-busy="loading"
        tabindex="-1"
        class="lottery-modal__dialog relative max-h-[calc(100vh-3rem)] w-full max-w-2xl overflow-y-auto p-5 sm:p-6"
    >
      <div class="lottery-modal__top-line absolute inset-x-6 top-0 h-px" aria-hidden="true" />

      <div class="mb-4 flex items-start justify-between gap-4">
        <h3 :id="titleId" class="lottery-modal__title">
          {{ lottery.lottery.title }}
        </h3>
        <button
            type="button"
            class="lottery-modal__close grid size-8 shrink-0 place-items-center"
            :aria-label="t('common.cancel')"
            @click="emit('close')"
        >
          ×
        </button>
      </div>

      <p class="lottery-modal__desc mb-5">
        {{ lottery.lottery.desc }}
      </p>

      <div class="lottery-modal__summary mb-5 grid gap-3 p-4">
        <div class="flex items-center justify-between gap-4">
          <div class="lottery-modal__label">{{ t('game.lottery.home.prize') }}</div>
          <div>
            {{ lottery.lottery.prize.title }}
            ({{ lottery.lottery.prize.platform }})
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="lottery-modal__stat px-3 py-2">
            <div class="lottery-modal__label">{{ t('game.lottery.home.prizeQuantity') }}</div>
            <div>
              {{ lottery.lottery.prize.count }}
            </div>
          </div>

          <div class="lottery-modal__stat px-3 py-2">
            <div class="lottery-modal__label">{{ t('game.lottery.home.participants') }}</div>
            <div>
              {{ lottery.count }}
            </div>
          </div>
        </div>
      </div>

      <div class="mb-4">
        <div class="lottery-modal__section-title mb-2">
          {{ t('game.lottery.submitModal.currentParticipants') }}
        </div>

        <div
            v-if="!lottery.member.length"
            class="lottery-modal__empty"
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
              class="lottery-modal__chip px-3 py-1"
          >
            {{ m.name }} - {{ m.email }}
          </span>
        </div>

        <div v-if="!allLoaded" class="mt-3 flex items-center justify-start">
          <button
              type="button"
              @click="loadMore"
              class="lottery-modal__load-more px-2.5 py-1"
          >
            {{ t('common.loadMore') }}
          </button>
        </div>

        <div
            v-else-if="lottery.member.length > 5"
            class="lottery-modal__empty lottery-modal__empty--complete mt-2"
        >
          {{ t('game.lottery.submitModal.allLoaded') }}
        </div>
      </div>

      <div class="space-y-3">
        <input
            ref="keyField"
            v-model="keyInput"
            :aria-label="t('game.lottery.submitModal.enterLotteryKey')"
            :placeholder="t('game.lottery.submitModal.enterLotteryKey')"
            class="lottery-modal__input w-full px-3 py-2.5"
        />

        <input
            v-model="nameInput"
            :aria-label="t('game.lottery.submitModal.enterName')"
            :placeholder="t('game.lottery.submitModal.enterName')"
            class="lottery-modal__input w-full px-3 py-2.5"
        />

        <input
            v-model="emailInput"
            :aria-label="t('game.lottery.submitModal.enterEmail')"
            :placeholder="t('game.lottery.submitModal.enterEmail')"
            class="lottery-modal__input w-full px-3 py-2.5"
        />

        <div v-if="emailError" class="lottery-modal__message lottery-modal__message--error">
          {{ emailError }}
        </div>

        <div v-if="submitError" class="lottery-modal__message lottery-modal__message--error">
          {{ submitError }}
        </div>

        <div v-if="successMsg" class="lottery-modal__message lottery-modal__message--success">
          {{ successMsg }}
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <button
            @click="emit('close')"
            class="lottery-modal__button lottery-modal__button--secondary px-4 py-2"
        >
          {{ t('common.cancel') }}
        </button>

        <button
            @click="submit"
            :disabled="loading"
            class="lottery-modal__button lottery-modal__button--primary px-4 py-2 disabled:cursor-not-allowed"
        >
          {{ loading ? t("common.commiting") : t("common.commit") }}
        </button>
      </div>
    </div>
  </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from "vue"
import { getLotteryParticipation } from "@/utils/api/game"
import type { LotteryActiveModel } from "@/types/game"
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  lottery: LotteryActiveModel
}>()

const emit = defineEmits(["close"])
const titleId = useId()
const panel = ref<HTMLElement | null>(null)
const keyField = ref<HTMLInputElement | null>(null)
let disposed = false
let cleanupDialog: (() => void) | undefined

// Lottery owns this body-mounted dialog lifecycle; Search/Review remain independent.
onMounted(async () => {
  const element = panel.value
  if (!element) return
  const trigger = document.activeElement instanceof HTMLElement ? document.activeElement : null
  const root = document.querySelector<HTMLElement>('#__nuxt')
  const wasInert = root?.inert ?? false
  const html = document.documentElement
  const overflow = html.style.overflow
  const gutter = html.style.scrollbarGutter
  const scroll = { left: window.scrollX, top: window.scrollY }
  html.style.scrollbarGutter = 'stable'
  html.style.overflow = 'hidden'
  const focusable = () => [...element.querySelectorAll<HTMLElement>('button, input, [tabindex]')]
    .filter(node => node.tabIndex >= 0 && !node.matches(':disabled') && node.getClientRects().length
      && getComputedStyle(node).visibility !== 'hidden' && !node.closest('[inert]'))
  const focusInside = () => (keyField.value ?? focusable()[0] ?? element).focus({ preventScroll: true })
  const onKey = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault(); event.stopImmediatePropagation(); emit('close')
    } else if (event.key === 'Tab') {
      const nodes = focusable(), first = nodes[0], last = nodes.at(-1)
      if (!first || !last) { event.preventDefault(); element.focus(); return }
      if (event.shiftKey && (document.activeElement === first || !element.contains(document.activeElement))) {
        event.preventDefault(); last.focus()
      } else if (!event.shiftKey && (document.activeElement === last || !element.contains(document.activeElement))) {
        event.preventDefault(); first.focus()
      }
    }
  }
  const onFocus = (event: FocusEvent) => {
    if (!element.contains(event.target as Node)) focusInside()
  }
  cleanupDialog = () => {
    document.removeEventListener('keydown', onKey, true)
    document.removeEventListener('focusin', onFocus)
    if (root) root.inert = wasInert
    html.style.overflow = overflow; html.style.scrollbarGutter = gutter
    window.scrollTo({ ...scroll, behavior: 'instant' })
    void nextTick(() => {
      if (trigger?.isConnected && !trigger.closest('[inert]')) trigger.focus({ preventScroll: true })
    })
  }
  await nextTick()
  if (disposed) return
  focusInside()
  if (root) root.inert = true
  document.addEventListener('keydown', onKey, true)
  document.addEventListener('focusin', onFocus)
})

onBeforeUnmount(() => {
  disposed = true
  cleanupDialog?.()
})

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
  if (loading.value) return
  submitError.value = ""
  emailError.value = ""
  successMsg.value = ""

  const req = {
    id: Number(props.lottery.lottery.id),
    name: nameInput.value.trim(),
    email: emailInput.value.trim(),
    key: keyInput.value.trim()
  }

  if (!req.key || !req.name || !req.email) {
    submitError.value = t('game.lottery.submitModal.fillAllInfo')
    return
  }

  if (!validateEmail(req.email)) {
    emailError.value = t('game.lottery.submitModal.invalidEmail')
    return
  }

  try {
    loading.value = true

    const res = await getLotteryParticipation(req)
    if (disposed) return

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

  } catch {
    if (!disposed) submitError.value = t('game.lottery.submitModal.networkError')
  } finally {
    loading.value = false
  }
}
</script>
