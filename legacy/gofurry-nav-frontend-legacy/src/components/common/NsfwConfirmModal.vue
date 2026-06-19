<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
    >
      <div class="bg-orange-50 w-[90%] max-w-xs sm:max-w-md rounded-2xl shadow-lg p-6 space-y-4">

        <h2 class="text-lg font-semibold text-gray-700">
          {{t("common.modal.adultConfirmTitle")}}
        </h2>

        <p class="text-sm text-gray-500 leading-relaxed">
          {{t("common.modal.adultConfirmDesc")}}
        </p>

        <div class="bg-orange-100 rounded-xl p-3 text-xs text-orange-900 leading-relaxed">
          <strong>{{t("common.modal.disclaimerTitle")}}</strong><br />
          {{t("common.modal.disclaimerContent")}}
        </div>

        <div class="flex justify-end gap-3 pt-2 text-sm sm:text-md">
          <button
              class="px-4 py-2 text-sm rounded-lg hover:bg-orange-100"
              @click="emit('cancel')"
          >
            <span class="text-gray-500">{{t("common.cancel")}}</span>
          </button>

          <button
              class="px-4 py-2 text-sm rounded-lg bg-orange-400 text-gray-50 hover:bg-orange-300"
              @click="emit('confirm')"
          >
            {{t("common.modal.adultConfirmContinue")}}
          </button>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { i18n } from '@/main.ts'

const { t } = i18n.global

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()
</script>
