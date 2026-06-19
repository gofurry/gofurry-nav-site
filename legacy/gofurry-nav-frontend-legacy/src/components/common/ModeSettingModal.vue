<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 px-3 py-4 backdrop-blur-sm sm:px-4"
    >
      <div class="flex max-h-[calc(100vh-4rem)] w-full max-w-md flex-col overflow-hidden rounded-lg border border-white/15 bg-[rgba(18,24,37,0.78)] p-4 text-gray-100 shadow-[0_24px_60px_rgba(15,23,42,0.34)] ring-1 ring-white/10 backdrop-blur-xl sm:max-h-[36rem] sm:p-6">
        <div class="shrink-0 space-y-2">
          <h2 class="text-lg font-semibold text-white">
            {{ t("navbar.modeSetting") }}
          </h2>

          <p class="text-sm leading-6 text-gray-300">
            {{ t("navbar.modeDesc") }}
          </p>
        </div>

        <div class="mt-5 flex-1 space-y-4 overflow-y-auto pr-1">
          <div>
            <input
                v-model="localMode"
                placeholder="nsfw"
                maxlength="32"
                class="w-full rounded-lg duration-300 border border-white/10 bg-white/8 px-4 py-3 text-sm text-gray-100 placeholder:text-gray-400 focus:outline-none"
            />
          </div>

          <label for="show-bubble" class="flex cursor-pointer items-center justify-between gap-3 rounded-lg border border-white/10 bg-white/5 px-4 py-3 transition hover:bg-white/10">
            <span class="text-sm text-gray-200">
              {{ t("navbar.showBubble") }}
            </span>
            <input
                id="show-bubble"
                v-model="showBubbleLocal"
                type="checkbox"
                class="h-4 w-4 rounded border-white/20 bg-white/10 text-orange-300"
            />
          </label>

          <label for="show-custom-sites" class="flex cursor-pointer items-center justify-between gap-3 rounded-lg border border-white/10 bg-white/5 px-4 py-3 transition hover:bg-white/10">
            <span class="text-sm text-gray-200">
              {{ t("navbar.showCustomSites") }}
            </span>
            <input
                id="show-custom-sites"
                v-model="showCustomSitesLocal"
                type="checkbox"
                class="h-4 w-4 rounded border-white/20 bg-white/10 text-orange-300"
            />
          </label>

          <div class="rounded-lg border border-white/10 bg-white/5 px-4 py-4">
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <div class="space-y-1">
                  <label class="block text-sm text-gray-200">
                    {{ t("navbar.customNavHeaderBg") }}
                  </label>
                  <p class="text-xs leading-5 text-gray-400">
                    {{ customBgFolderNameLocal
                      ? t('navbar.customNavHeaderBgSelected', { name: customBgFolderNameLocal })
                      : t('navbar.customNavHeaderBgEmpty') }}
                  </p>
                </div>
                <span
                    class="rounded-full border border-white/10 bg-white/8 px-2 py-1 text-[11px] text-gray-300"
                >
                  {{ supportsCustomBgPicker
                    ? t('navbar.customNavHeaderBgSupported')
                    : t('navbar.customNavHeaderBgUnsupported') }}
                </span>
              </div>

              <div class="flex flex-wrap gap-2 pt-1">
                <button
                    type="button"
                    class="rounded-lg border border-white/10 bg-white/8 px-3 py-2 text-sm text-gray-100 transition hover:bg-white/12 disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="!supportsCustomBgPicker"
                    @click="pickCustomBgDirectory"
                >
                  {{ t('navbar.customNavHeaderBgPick') }}
                </button>
                <button
                    type="button"
                    class="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-gray-200 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="!customBgFolderNameLocal"
                    @click="clearCustomBgDirectory"
                >
                  {{ t('navbar.customNavHeaderBgClear') }}
                </button>
              </div>

              <p class="text-xs leading-5 text-gray-400">
                {{ t("navbar.customNavHeaderBgDesc") }}
              </p>
            </div>
          </div>

          <div class="space-y-2">
            <label for="custom-panel-height" class="block text-sm text-gray-200">
              {{ t("navbar.customPanelHeight") }}
            </label>
            <input
                id="custom-panel-height"
                v-model="customPanelHeightLocal"
                class="w-full rounded-lg border border-white/10 bg-white/8 px-4 py-3 text-sm text-gray-100 placeholder:text-gray-400 focus:outline-none"
                :placeholder="t('navbar.customPanelHeightPlaceholder')"
            />
            <p class="text-xs leading-5 text-gray-400">
              {{ t("navbar.customPanelHeightDesc") }}
            </p>
          </div>

          <div class="space-y-2">
            <label for="custom-panel-code" class="block text-sm text-gray-200">
              {{ t("navbar.customPanelCode") }}
            </label>
            <textarea
                id="custom-panel-code"
                v-model="customPanelCodeLocal"
                rows="3"
                class="w-full rounded-lg border border-white/10 bg-white/8 px-4 py-3 text-sm text-gray-100 placeholder:text-gray-400 focus:outline-none"
                :placeholder="t('navbar.customPanelCodePlaceholder')"
            ></textarea>
            <p class="text-xs leading-5 text-gray-400">
              {{ t("navbar.customPanelCodeDesc") }}
            </p>
          </div>
        </div>

        <div class="mt-5 flex shrink-0 justify-end gap-3 border-t border-white/10 pt-4">
          <button
              class="rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-sm text-gray-200 transition hover:bg-white/10"
              @click="emit('cancel')"
          >
            {{ t("common.cancel") }}
          </button>
          <button
              class="rounded-lg bg-orange-300 px-4 py-2 text-sm font-medium text-slate-900 transition hover:bg-orange-200"
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
import { i18n } from '@/main.ts'
import {
  loadCustomPanelCode,
  loadCustomPanelHeight,
  saveCustomPanelCode,
  saveCustomPanelHeight,
} from '@/utils/customPanel'
import {
  clearCustomNavHeaderBackgroundDirectory,
  type CustomNavHeaderBackgroundSelection,
  loadCustomNavHeaderBackgroundMeta,
  pickCustomNavHeaderBackgroundDirectory,
  saveCustomNavHeaderBackgroundDirectory,
  supportsCustomNavHeaderBackground,
} from '@/utils/customNavHeaderBackground'

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
const showBubbleLocal = ref(true)
const showCustomSitesLocal = ref(true)
const customPanelCodeLocal = ref('')
const customPanelHeightLocal = ref('320')
const supportsCustomBgPicker = supportsCustomNavHeaderBackground()
const customBgFolderNameLocal = ref('')

let pendingCustomBgSelection: CustomNavHeaderBackgroundSelection | null = null
let shouldClearCustomBg = false

function syncCustomBgState() {
  const meta = loadCustomNavHeaderBackgroundMeta()
  customBgFolderNameLocal.value = meta.folderName
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
        customPanelCodeLocal.value = loadCustomPanelCode()
        customPanelHeightLocal.value = String(loadCustomPanelHeight())
        syncCustomBgState()
      }
    }
)

onMounted(() => {
  const savedBubble = localStorage.getItem('showBubble')
  const savedCustomSites = localStorage.getItem('showCustomSites')
  showBubbleLocal.value = savedBubble !== 'false'
  showCustomSitesLocal.value = savedCustomSites !== 'false'
  customPanelCodeLocal.value = loadCustomPanelCode()
  customPanelHeightLocal.value = String(loadCustomPanelHeight())
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
  emit('save', localMode.value)
  localStorage.setItem('showBubble', showBubbleLocal.value ? 'true' : 'false')
  localStorage.setItem('showCustomSites', showCustomSitesLocal.value ? 'true' : 'false')
  window.dispatchEvent(
    new CustomEvent('show-bubble-change', {
      detail: { visible: showBubbleLocal.value },
    })
  )
  saveCustomPanelCode(customPanelCodeLocal.value)
  saveCustomPanelHeight(customPanelHeightLocal.value)

  if (shouldClearCustomBg) {
    await clearCustomNavHeaderBackgroundDirectory()
  } else if (pendingCustomBgSelection) {
    await saveCustomNavHeaderBackgroundDirectory(pendingCustomBgSelection)
  }

  window.dispatchEvent(new Event('custom-sites-visibility-change'))
}
</script>
