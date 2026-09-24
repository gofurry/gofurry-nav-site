<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="gf-modal-backdrop fixed inset-0 z-50 flex items-center justify-center px-4 py-5"
    >
      <div ref="panel" class="gf-modal gf-modal--compact" role="dialog" aria-modal="true" :aria-labelledby="titleId" tabindex="-1">
        <div class="gf-modal__body">

          <h2 :id="titleId" class="gf-modal__title">
            {{t("common.modal.adultConfirmTitle")}}
          </h2>

          <p class="gf-modal__desc">
            {{t("common.modal.adultConfirmDesc")}}
          </p>

          <div class="gf-modal__notice">
            <strong class="gf-modal__notice-title">{{t("common.modal.disclaimerTitle")}}</strong><br />
            {{t("common.modal.disclaimerContent")}}
          </div>

          <div class="gf-modal__footer gf-modal__footer--flush">
            <button
                class="gf-button gf-button--ghost"
                @click="emit('cancel')"
            >
              {{t("common.cancel")}}
            </button>

            <button
                class="gf-button gf-button--primary"
                @click="emit('confirm')"
            >
              {{t("common.modal.adultConfirmContinue")}}
            </button>
          </div>

        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { i18n } from '@/main'
import { ref, useId } from 'vue'
import { useGameDetailDialog } from '@/composables/useGameDetailDialog'

const { t } = i18n.global

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()
const panel = ref<HTMLElement | null>(null)
const titleId = useId()
useGameDetailDialog(panel, () => emit('cancel'))
</script>
