<template>
  <Teleport to="body">
    <Transition
      name="review-overlay"
    >
      <div
        v-if="visible"
        class="review-dialog-backdrop fixed inset-0 z-50 flex items-center justify-center px-4"
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
            <h2 class="review-dialog__title">
              {{ t('game.detail.makeComment') }}
            </h2>
            <p class="review-dialog__game-name mt-1 truncate">
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
              <span class="review-dialog__score-label mb-1 block">
                {{ t('game.detail.score') }} (0.0 ~ 5.0)
              </span>
              <input
                v-model.trim="form.score"
                type="text"
                inputmode="decimal"
                autocomplete="off"
                placeholder="0.0"
                class="review-field review-score-field"
                @blur="normalizeScore"
              />
            </label>

            <p v-if="errorMsg" class="review-dialog__feedback review-dialog__feedback--error">
              {{ errorMsg }}
            </p>
            <p v-if="successMsg" class="review-dialog__feedback review-dialog__feedback--success">
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
  score: '0',
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
    form.score = '0'
  }
)

function parseScore() {
  const raw = form.score.trim().replace(',', '.')
  if (!/^\d+(?:\.\d+)?$/.test(raw)) {
    return null
  }

  const score = Number(raw)
  if (!Number.isFinite(score)) {
    return null
  }

  return score
}

function normalizeScore() {
  const score = parseScore()
  if (score === null || score < 0 || score > 5) {
    return
  }

  form.score = String(Math.round(score * 10) / 10)
}

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

  const score = parseScore()
  if (score === null || score < 0 || score > 5) {
    errorMsg.value = t('game.detail.scoreTip')
    return false
  }

  return true
}

async function submit() {
  if (!validate()) {
    return
  }

  const score = parseScore()
  if (score === null) {
    errorMsg.value = t('game.detail.scoreTip')
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
      score,
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
  padding: 1.2rem;
}

.review-dialog__close {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  display: grid;
  width: 1.9rem;
  height: 1.9rem;
  place-items: center;
}

.review-field {
  width: 100%;
  padding: 0.65rem 0.75rem;
}

.review-submit {
  width: 100%;
  padding: 0.72rem 1rem;
}
</style>
