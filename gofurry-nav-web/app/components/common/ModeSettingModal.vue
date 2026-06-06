<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-slate-950/64 px-3 py-4 backdrop-blur-md sm:px-4"
    >
      <div class="mode-setting-dialog">
        <div class="mode-setting-header">
          <p class="mode-setting-eyebrow">{{ t("navbar.preferences") }}</p>
          <h2 class="mode-setting-title">
            {{ t("navbar.modeSetting") }}
          </h2>

          <p class="mode-setting-desc">
            {{ t("navbar.modeDesc") }}
          </p>
        </div>

        <div class="mode-setting-body">
          <section class="mode-setting-card">
            <div class="mode-setting-card-copy">
              <label class="mode-setting-label" for="mode-setting-input">
                {{ t("navbar.displayMode") }}
              </label>
              <p class="mode-setting-help">{{ t("navbar.displayModeDesc") }}</p>
            </div>
            <input
                id="mode-setting-input"
                v-model="localMode"
                placeholder="nsfw"
                maxlength="32"
                class="mode-setting-input"
            />
          </section>

          <section class="mode-setting-card mode-setting-card-inline">
            <div class="mode-setting-card-copy">
              <label class="mode-setting-label" for="quick-access-toggle">
                {{ t("navbar.quickAccess") }}
              </label>
              <p class="mode-setting-help">{{ t("navbar.quickAccessDesc") }}</p>
            </div>
            <button
                id="quick-access-toggle"
                type="button"
                class="mode-setting-toggle"
                :class="{ 'mode-setting-toggle-on': showQuickAccessLocal }"
                :aria-pressed="showQuickAccessLocal"
                @click="showQuickAccessLocal = !showQuickAccessLocal"
            >
              <span></span>
              <em>{{ showQuickAccessLocal ? t('navbar.quickAccessOn') : t('navbar.quickAccessOff') }}</em>
            </button>
          </section>

          <section class="mode-setting-card">
            <div class="mode-setting-card-heading">
              <div class="mode-setting-card-copy">
                <label class="mode-setting-label">
                  {{ t("navbar.customNavHeaderBg") }}
                </label>
                <p class="mode-setting-help">
                  {{ customBgFolderNameLocal
                    ? t('navbar.customNavHeaderBgSelected', { name: customBgFolderNameLocal })
                    : t('navbar.customNavHeaderBgEmpty') }}
                </p>
              </div>
              <span class="mode-setting-badge">
                {{ supportsCustomBgPicker
                  ? t('navbar.customNavHeaderBgSupported')
                  : t('navbar.customNavHeaderBgUnsupported') }}
              </span>
            </div>

            <div class="mode-setting-actions">
              <button
                  type="button"
                  class="mode-setting-soft-button"
                  :disabled="!supportsCustomBgPicker"
                  @click="pickCustomBgDirectory"
              >
                {{ t('navbar.customNavHeaderBgPick') }}
              </button>
              <button
                  type="button"
                  class="mode-setting-ghost-button"
                  :disabled="!customBgFolderNameLocal"
                  @click="clearCustomBgDirectory"
              >
                {{ t('navbar.customNavHeaderBgClear') }}
              </button>
            </div>

            <p class="mode-setting-footnote">
              {{ t("navbar.customNavHeaderBgDesc") }}
            </p>
          </section>

        </div>

        <div class="mode-setting-footer">
          <button
              class="mode-setting-cancel"
              @click="emit('cancel')"
          >
            {{ t("common.cancel") }}
          </button>
          <button
              class="mode-setting-save"
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

<style scoped>
.mode-setting-dialog {
  display: flex;
  max-height: min(42rem, calc(100vh - 3rem));
  width: min(100%, 33rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 1rem;
  background:
    linear-gradient(180deg, rgba(22, 30, 45, 0.94), rgba(12, 18, 30, 0.92)),
    rgba(15, 23, 42, 0.88);
  box-shadow: 0 28px 72px rgba(2, 6, 23, 0.42);
  color: #f8fafc;
  backdrop-filter: blur(22px);
}

.mode-setting-header {
  padding: 1.35rem 1.45rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.mode-setting-eyebrow {
  margin: 0 0 0.35rem;
  color: rgba(253, 186, 116, 0.9);
  font-size: 0.72rem;
  font-weight: 700;
}

.mode-setting-title {
  margin: 0;
  font-size: 1.12rem;
  font-weight: 750;
}

.mode-setting-desc,
.mode-setting-help,
.mode-setting-footnote {
  margin: 0;
  color: rgba(203, 213, 225, 0.78);
  font-size: 0.78rem;
  line-height: 1.6;
}

.mode-setting-body {
  display: grid;
  flex: 1;
  gap: 0.75rem;
  overflow-y: auto;
  padding: 1rem 1.45rem;
}

.mode-setting-card {
  display: grid;
  gap: 0.85rem;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.5);
  padding: 1rem;
}

.mode-setting-card-inline {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.mode-setting-card-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.mode-setting-card-copy {
  min-width: 0;
}

.mode-setting-label {
  display: block;
  margin-bottom: 0.25rem;
  color: #f8fafc;
  font-size: 0.86rem;
  font-weight: 700;
}

.mode-setting-input {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.26);
  border-radius: 0.65rem;
  background: rgba(15, 23, 42, 0.72);
  color: #f8fafc;
  font-size: 0.86rem;
  padding: 0.78rem 0.9rem;
  outline: none;
  transition: border-color 180ms ease, background 180ms ease;
}

.mode-setting-input:focus {
  border-color: rgba(251, 146, 60, 0.58);
  background: rgba(15, 23, 42, 0.9);
}

.mode-setting-toggle {
  display: inline-grid;
  min-width: 4.4rem;
  grid-template-columns: 1.45rem 1fr;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgba(148, 163, 184, 0.26);
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.72);
  color: rgba(226, 232, 240, 0.78);
  cursor: pointer;
  padding: 0.28rem 0.58rem 0.28rem 0.28rem;
}

.mode-setting-toggle span {
  display: block;
  width: 1.2rem;
  height: 1.2rem;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.58);
  transition: background 180ms ease;
}

.mode-setting-toggle em {
  font-size: 0.72rem;
  font-style: normal;
  font-weight: 700;
}

.mode-setting-toggle-on {
  border-color: rgba(251, 146, 60, 0.42);
  background: rgba(251, 146, 60, 0.16);
  color: #fed7aa;
}

.mode-setting-toggle-on span {
  background: #fb923c;
}

.mode-setting-badge {
  flex: 0 0 auto;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
  color: rgba(226, 232, 240, 0.84);
  font-size: 0.68rem;
  padding: 0.24rem 0.55rem;
}

.mode-setting-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.mode-setting-soft-button,
.mode-setting-ghost-button,
.mode-setting-cancel,
.mode-setting-save {
  border-radius: 0.65rem;
  cursor: pointer;
  font-size: 0.82rem;
  transition: background 180ms ease, border-color 180ms ease, opacity 180ms ease;
}

.mode-setting-soft-button,
.mode-setting-ghost-button {
  border: 1px solid rgba(148, 163, 184, 0.22);
  padding: 0.55rem 0.75rem;
}

.mode-setting-soft-button {
  background: rgba(255, 255, 255, 0.08);
  color: #f8fafc;
}

.mode-setting-ghost-button,
.mode-setting-cancel {
  background: rgba(255, 255, 255, 0.04);
  color: rgba(226, 232, 240, 0.86);
}

.mode-setting-soft-button:hover,
.mode-setting-ghost-button:hover,
.mode-setting-cancel:hover {
  background: rgba(255, 255, 255, 0.1);
}

.mode-setting-soft-button:disabled,
.mode-setting-ghost-button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.mode-setting-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.65rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1rem 1.45rem;
}

.mode-setting-cancel,
.mode-setting-save {
  border: 1px solid rgba(148, 163, 184, 0.22);
  padding: 0.62rem 0.95rem;
}

.mode-setting-save {
  border-color: rgba(251, 146, 60, 0.4);
  background: #fdba74;
  color: #1e293b;
  font-weight: 750;
}

.mode-setting-save:hover {
  background: #fed7aa;
}

@media (max-width: 640px) {
  .mode-setting-card-inline {
    grid-template-columns: 1fr;
  }
}
</style>
