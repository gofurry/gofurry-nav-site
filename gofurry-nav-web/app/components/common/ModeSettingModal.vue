<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="gf-modal-backdrop fixed inset-0 z-[120] flex items-center justify-center px-3 py-4 sm:px-4"
    >
      <div class="gf-modal">
        <div class="gf-modal__header">
          <p class="gf-modal__eyebrow">{{ t("navbar.preferences") }}</p>
          <h2 class="gf-modal__title">
            {{ t("navbar.modeSetting") }}
          </h2>

          <p class="gf-modal__desc">
            {{ t("navbar.modeDesc") }}
          </p>
        </div>

        <div class="gf-modal__body">
          <section class="gf-card gf-card--flat gf-modal__section">
            <div class="gf-modal__copy">
              <label class="gf-modal__label" for="mode-setting-input">
                {{ t("navbar.displayMode") }}
              </label>
              <p class="gf-modal__help">{{ t("navbar.displayModeDesc") }}</p>
            </div>
            <input
                id="mode-setting-input"
                v-model="localMode"
                placeholder="nsfw"
                maxlength="32"
                class="gf-input"
            />
          </section>

          <section class="gf-card gf-card--flat gf-modal__section gf-modal__section--inline">
            <div class="gf-modal__copy">
              <label class="gf-modal__label" for="quick-access-toggle">
                {{ t("navbar.quickAccess") }}
              </label>
              <p class="gf-modal__help">{{ t("navbar.quickAccessDesc") }}</p>
            </div>
            <button
                id="quick-access-toggle"
                type="button"
                class="gf-modal__toggle"
                :class="{ 'gf-modal__toggle--on': showQuickAccessLocal }"
                :aria-pressed="showQuickAccessLocal"
                @click="showQuickAccessLocal = !showQuickAccessLocal"
            >
              <span></span>
              <em>{{ showQuickAccessLocal ? t('navbar.quickAccessOn') : t('navbar.quickAccessOff') }}</em>
            </button>
          </section>

          <section class="gf-card gf-card--flat gf-modal__section">
            <div class="gf-modal__section-heading">
              <div class="gf-modal__copy">
                <label class="gf-modal__label">
                  {{ t("navbar.customNavHeaderBg") }}
                </label>
                <p class="gf-modal__help">
                  {{ customBgFolderNameLocal
                    ? t('navbar.customNavHeaderBgSelected', { name: customBgFolderNameLocal })
                    : t('navbar.customNavHeaderBgEmpty') }}
                </p>
              </div>
              <span class="gf-chip gf-chip--muted">
                {{ supportsCustomBgPicker
                  ? t('navbar.customNavHeaderBgSupported')
                  : t('navbar.customNavHeaderBgUnsupported') }}
              </span>
            </div>

            <div class="gf-modal__actions">
              <button
                  type="button"
                  class="gf-button gf-button--surface"
                  :disabled="!supportsCustomBgPicker"
                  @click="pickCustomBgDirectory"
              >
                {{ t('navbar.customNavHeaderBgPick') }}
              </button>
              <button
                  type="button"
                  class="gf-button gf-button--ghost"
                  :disabled="!customBgFolderNameLocal"
                  @click="clearCustomBgDirectory"
              >
                {{ t('navbar.customNavHeaderBgClear') }}
              </button>
            </div>

            <p class="gf-modal__footnote">
              {{ t("navbar.customNavHeaderBgDesc") }}
            </p>
          </section>

        </div>

        <div class="gf-modal__footer">
          <button
              class="gf-button gf-button--ghost"
              @click="emit('cancel')"
          >
            {{ t("common.cancel") }}
          </button>
          <button
              class="gf-button gf-button--primary"
              @click="save"
          >
            {{ t("common.save") }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { i18n } from '@/main'
import {
  clearCustomNavHeaderBackgroundDirectory,
  type CustomNavHeaderBackgroundSelection,
  loadCustomNavHeaderBackgroundMeta,
  pickCustomNavHeaderBackgroundDirectory,
  saveCustomNavHeaderBackgroundDirectory,
  supportsCustomNavHeaderBackground,
} from '@/utils/customNavHeaderBackground'
import {
  readShowQuickAccess,
  writeShowQuickAccess,
} from '@/utils/navHeaderSettings'

const { t } = i18n.global

const props = defineProps<{
  show: boolean
  mode: string
}>()

const emit = defineEmits<{
  (e: 'save', value: string): void
  (e: 'cancel'): void
}>()

const localMode = ref('')
const showQuickAccessLocal = ref(true)
const supportsCustomBgPicker = supportsCustomNavHeaderBackground()
const customBgFolderNameLocal = ref('')

let pendingCustomBgSelection: CustomNavHeaderBackgroundSelection | null = null
let shouldClearCustomBg = false

function syncCustomBgState() {
  const meta = loadCustomNavHeaderBackgroundMeta()
  customBgFolderNameLocal.value = meta.folderName
  showQuickAccessLocal.value = readShowQuickAccess()
  pendingCustomBgSelection = null
  shouldClearCustomBg = false
}

watch(
    () => props.mode,
    value => {
      localMode.value = value
    },
    { immediate: true }
)

watch(
    () => props.show,
    visible => {
      if (visible) {
        syncCustomBgState()
      }
    }
)

onMounted(() => {
  syncCustomBgState()
})

async function pickCustomBgDirectory() {
  try {
    const selection = await pickCustomNavHeaderBackgroundDirectory()
    if (!selection) {
      return
    }

    pendingCustomBgSelection = selection
    customBgFolderNameLocal.value = selection.folderName
    shouldClearCustomBg = false
  } catch (error) {
    console.error('Pick custom nav header background directory err:', error)
  }
}

function clearCustomBgDirectory() {
  pendingCustomBgSelection = null
  customBgFolderNameLocal.value = ''
  shouldClearCustomBg = true
}

const save = async () => {
  localMode.value = localMode.value.trim().slice(0, 32)
  writeShowQuickAccess(showQuickAccessLocal.value)
  emit('save', localMode.value)

  if (shouldClearCustomBg) {
    await clearCustomNavHeaderBackgroundDirectory()
  } else if (pendingCustomBgSelection) {
    await saveCustomNavHeaderBackgroundDirectory(pendingCustomBgSelection)
  }
}
</script>
