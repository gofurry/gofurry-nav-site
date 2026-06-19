<template>
  <!-- 遮罩 -->
  <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/30"
  >
    <!-- 提示框 -->
    <div
        class="w-full max-w-md bg-orange-50 rounded-2xl shadow-xl p-5 relative animate-fade-in"
    >
      <!-- 关闭按钮 -->
      <button
          class="absolute right-3 top-3 text-orange-400 hover:text-orange-600"
          @click="close"
      >
        ✕
      </button>

      <!-- 标题 -->
      <h2 class="text-lg font-bold text-orange-700 mb-1">
        {{t("game.detail.makeComment")}}
      </h2>
      <p class="text-sm text-orange-600 mb-4">
        {{ gameName }}
      </p>

      <!-- 表单 -->
      <div class="space-y-3">
        <!-- 昵称 -->
        <div>
          <input
              v-model.trim="form.name"
              type="text"
              :placeholder="t('game.detail.inputName')"
              class="w-full px-3 py-2 rounded-lg border border-orange-200 focus:outline-none focus:ring-1 focus:ring-orange-300"
          />
        </div>

        <!-- 内容 -->
        <div>
          <textarea
              v-model.trim="form.content"
              :placeholder="t('game.detail.inputContent')"
              rows="8"
              class="w-full px-3 py-2 rounded-lg border border-orange-200 resize-none focus:outline-none focus:ring-1 focus:ring-orange-300"
          />
        </div>

        <!-- 评分 -->
        <div>
          <label class="block text-sm text-orange-700 mb-1">
            {{t("game.detail.score")}} ( 0.0 ~ 5.0 )
          </label>
          <input
              v-model.number="form.score"
              class="w-full px-3 py-2 rounded-lg border border-orange-200 focus:outline-none focus:ring-1 focus:ring-orange-300"
          />
        </div>

        <!-- 错误提示 -->
        <p v-if="errorMsg" class="text-sm text-red-500">
          {{ errorMsg }}
        </p>

        <!-- 成功提示 -->
        <p v-if="successMsg" class="text-sm text-green-600">
          {{ successMsg }}
        </p>

        <!-- 提交按钮 -->
        <button
            class="w-full py-2 rounded-lg bg-orange-400 text-white font-medium hover:bg-orange-300 transition disabled:opacity-50"
            :disabled="submitting"
            @click="submit"
        >
          {{ submitting ? t("common.commiting") : t("game.detail.commitComment") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { commitComment } from '@/utils/api/game'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const props = defineProps<{
  visible: boolean
  gameName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const route = useRoute()

const form = reactive({
  name: '',
  content: '',
  score: 0.0,
})

const errorMsg = ref('')
const successMsg = ref('')
const submitting = ref(false)

watch(
    () => props.visible,
    (v) => {
      if (v) {
        errorMsg.value = ''
        successMsg.value = ''
        form.name = ''
        form.content = ''
        form.score = 0.0
      }
    }
)

const validate = (): boolean => {
  errorMsg.value = ''

  if (!form.name || !form.content) {
    errorMsg.value = t("game.detail.contentTip")
    return false
  }

  if (form.score < 0 || form.score > 5.0) {
    errorMsg.value = t("game.detail.scoreTip")
    return false
  }

  return true
}

const submit = async () => {
  if (!validate()) return

  submitting.value = true
  errorMsg.value = ''
  successMsg.value = ''

  try {
    const res = await commitComment({
      id: route.params.id as string,
      name: form.name,
      content: form.content,
      score: form.score,
    })

    if (res.code === 1) {
      successMsg.value = t("game.detail.commitSuccess")
    } else {
      errorMsg.value = res.data
    }
  } catch (e) {
    errorMsg.value = t("game.detail.commitFail")
  } finally {
    submitting.value = false
  }
}

const close = () => {
  emit('close')
}
</script>

<style scoped>
@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 0.2s ease-out;
}
</style>
