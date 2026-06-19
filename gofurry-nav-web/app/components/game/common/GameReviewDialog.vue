<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="visible"
        class="fixed inset-0 z-50 flex items-center justify-center bg-stone-950/30 px-4 backdrop-blur-[2px]"
        @click.self="close"
      >
        <div class="review-dialog">
          <button
            class="review-dialog__close"
            type="button"
            :aria-label="t('common.close')"
            @click="close"
          >
            x
          </button>

          <div class="mb-4">
            <h2 class="text-lg font-bold text-orange-950">
              {{ t('game.detail.makeComment') }}
            </h2>
            <p class="mt-1 truncate text-sm text-stone-500">
              {{ gameName }}
            </p>
          </div>

          <div class="space-y-3">
            <input
              v-model.trim="form.name"
              type="text"
              :placeholder="t('game.detail.inputName')"
              class="review-field"
            />

            <textarea
              v-model.trim="form.content"
              :placeholder="t('game.detail.inputContent')"
              rows="7"
              class="review-field resize-none"
            />

            <label class="block">
              <span class="mb-1 block text-sm font-medium text-stone-600">
                {{ t('game.detail.score') }} (0.0 ~ 5.0)
              </span>
              <input
                v-model.number="form.score"
                type="number"
                min="0"
                max="5"
                step="0.1"
                class="review-field"
              />
            </label>

            <p v-if="errorMsg" class="text-sm text-red-500">
              {{ errorMsg }}
            </p>
            <p v-if="successMsg" class="text-sm text-green-600">
              {{ successMsg }}
            </p>

            <button
              class="review-submit"
              type="button"
              :disabled="submitting"
              @click="submit"
            >
              {{ submitting ? t('common.commiting') : t('game.detail.commitComment') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { commitComment } from '@/utils/api/game'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  visible: boolean
  gameId: string
  gameName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submitted'): void
}>()

const form = reactive({
  name: '',
  content: '',
  score: 0,
})

const errorMsg = ref('')
const successMsg = ref('')
const submitting = ref(false)

watch(
  () => props.visible,
  (value) => {
    if (!value) {
      return
    }

    errorMsg.value = ''
    successMsg.value = ''
    form.name = ''
    form.content = ''
    form.score = 0
  }
)

function validate() {
  errorMsg.value = ''

  if (!props.gameId) {
    errorMsg.value = t('game.detail.commitFail')
    return false
  }

  if (!form.name || !form.content) {
    errorMsg.value = t('game.detail.contentTip')
    return false
  }

  if (form.score < 0 || form.score > 5) {
    errorMsg.value = t('game.detail.scoreTip')
    return false
  }

  return true
}

async function submit() {
  if (!validate()) {
    return
  }

  submitting.value = true
  errorMsg.value = ''
  successMsg.value = ''

  try {
    const res = await commitComment({
      id: props.gameId,
      name: form.name,
      content: form.content,
      score: form.score,
    })

    if (res.code === 1) {
      successMsg.value = t('game.detail.commitSuccess')
      emit('submitted')
      return
    }

    errorMsg.value = res.data
  } catch {
    errorMsg.value = t('game.detail.commitFail')
  } finally {
    submitting.value = false
  }
}

function close() {
  emit('close')
}
</script>

<style scoped>
.review-dialog {
  position: relative;
  width: min(30rem, 100%);
  border: 1px solid rgba(126, 92, 58, 0.18);
  border-radius: 1.05rem;
  background: rgba(255, 250, 242, 0.92);
  padding: 1.2rem;
  box-shadow: 0 22px 70px rgba(45, 28, 12, 0.20);
  backdrop-filter: blur(12px);
  animation: review-dialog-in 180ms cubic-bezier(0.22, 1, 0.36, 1);
}

.review-dialog__close {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  display: grid;
  width: 1.9rem;
  height: 1.9rem;
  place-items: center;
  border-radius: 999px;
  color: rgba(120, 83, 53, 0.72);
  transition: background-color 160ms ease, color 160ms ease;
}

.review-dialog__close:hover {
  background: rgba(251, 146, 60, 0.12);
  color: rgba(124, 45, 18, 0.94);
}

.review-field {
  width: 100%;
  border: 1px solid rgba(126, 92, 58, 0.18);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.46);
  padding: 0.65rem 0.75rem;
  color: rgba(41, 37, 36, 0.92);
  outline: none;
  transition: border-color 160ms ease, background-color 160ms ease;
}

.review-field:focus {
  border-color: rgba(249, 115, 22, 0.42);
  background: rgba(255, 255, 255, 0.68);
}

.review-submit {
  width: 100%;
  border-radius: 0.75rem;
  background: rgba(180, 83, 9, 0.86);
  padding: 0.72rem 1rem;
  color: white;
  font-weight: 700;
  transition: background-color 160ms ease, opacity 160ms ease;
}

.review-submit:hover {
  background: rgba(154, 52, 18, 0.92);
}

.review-submit:disabled {
  opacity: 0.56;
}

@keyframes review-dialog-in {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.985);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
