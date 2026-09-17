<template>
  <Teleport to="body">
    <div
        v-if="show"
        class="gf-modal-backdrop fixed inset-0 z-[120] flex items-center justify-center px-3 py-4 sm:px-4"
    >
      <div class="gf-modal gf-preferences-modal">
        <div class="gf-modal__header">
          <h2 class="gf-modal__title">
            {{ t("navbar.preferences") }}
          </h2>

          <div class="gf-modal__header-actions">
            <button
                type="button"
                class="gf-button gf-button--ghost"
                @click="emit('cancel')"
            >
              {{ t("common.cancel") }}
            </button>
            <button
                type="button"
                class="gf-button gf-button--primary"
                @click="save"
            >
              {{ t("common.save") }}
            </button>
          </div>
        </div>

        <div class="preferences-tabs" role="tablist" :aria-label="t('navbar.preferences')">
          <button v-for="(tab, index) in tabs" :id="`${panelId}-tab-${index}`" :key="tab" type="button" role="tab"
              :aria-selected="activeTab === index" :aria-controls="`${panelId}-panel-${index}`" :tabindex="activeTab === index ? 0 : -1"
              @click="selectTab(index)" @keydown="tabKeydown($event, index)">{{ tab }}</button>
        </div>
        <div ref="pages" class="preferences-pages" @scroll="syncTab">
          <div :id="`${panelId}-panel-0`" role="tabpanel" :aria-labelledby="`${panelId}-tab-0`" :inert="activeTab !== 0" class="preferences-page">
          <section class="gf-modal__section">
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

          <section class="gf-modal__section gf-modal__section--inline">
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
                :aria-label="t('navbar.quickAccess')"
                @click="showQuickAccessLocal = !showQuickAccessLocal"
            >
              <span></span>
            </button>
          </section>

          <section class="gf-modal__section">
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
          <div :id="`${panelId}-panel-1`" role="tabpanel" :aria-labelledby="`${panelId}-tab-1`" :inert="activeTab !== 1" class="preferences-page">
            <BackgroundPreferencesEditor ref="backgroundEditor" />
          </div>
          <div :id="`${panelId}-panel-2`" role="tabpanel" :aria-labelledby="`${panelId}-tab-2`" :inert="activeTab !== 2" class="preferences-page">
            <ResourceRoutingPreferencesEditor ref="routingEditor" />
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, useId, watch } from 'vue'
import BackgroundPreferencesEditor from './BackgroundPreferencesEditor.vue'
import ResourceRoutingPreferencesEditor from './ResourceRoutingPreferencesEditor.vue'
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

const panelId = useId()
const activeTab = ref(0)
const pages = ref<HTMLElement | null>(null)
const tabs = computed(() => [t('navbar.homePreferences'), t('navbar.pageBackground'), t('resourceRouting.title')])
function selectTab(index: number) {
  activeTab.value = index
  const element = pages.value
  if (element) element.scrollTo({ left: index * element.clientWidth, behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth' })
}
let tabScrollTimer: ReturnType<typeof setTimeout> | undefined
function syncTab() {
  clearTimeout(tabScrollTimer)
  tabScrollTimer = setTimeout(() => {
    if (pages.value?.clientWidth) activeTab.value = Math.round(pages.value.scrollLeft / pages.value.clientWidth)
  }, 100)
}
onUnmounted(() => clearTimeout(tabScrollTimer))
function tabKeydown(event: KeyboardEvent, index: number) {
  const count = tabs.value.length
  const next = event.key === 'Home' ? 0 : event.key === 'End' ? count - 1
    : event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count : undefined
  if (next === undefined) return
  event.preventDefault()
  selectTab(next)
  document.getElementById(`${panelId}-tab-${next}`)?.focus()
}
const localMode = ref('')
const backgroundEditor = ref<InstanceType<typeof BackgroundPreferencesEditor> | null>(null)
const routingEditor = ref<InstanceType<typeof ResourceRoutingPreferencesEditor> | null>(null)
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
        activeTab.value = 0
        localMode.value = props.mode
        nextTick(() => { pages.value?.scrollTo({ left: 0, behavior: 'instant' }) })
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
  if (backgroundEditor.value && !(await backgroundEditor.value.save())) { selectTab(1); return }
  localMode.value = localMode.value.trim().slice(0, 32)

  if (shouldClearCustomBg) {
    await clearCustomNavHeaderBackgroundDirectory()
  } else if (pendingCustomBgSelection) {
    await saveCustomNavHeaderBackgroundDirectory(pendingCustomBgSelection)
  }

  writeShowQuickAccess(showQuickAccessLocal.value)
  routingEditor.value?.save()
  emit('save', localMode.value)
}
</script>

<style scoped>
.gf-preferences-modal .gf-modal__header { border-bottom: 0; flex-shrink: 0; }
.preferences-tabs { display: flex; flex-shrink: 0; margin: 0 1.3rem; border-bottom: 1px solid var(--gf-border-strong); overflow-x: auto; }
.preferences-tabs button { position: relative; flex: 1; padding: .85rem 1rem; border: 0; background: transparent; color: var(--gf-text-muted); font-size: .88rem; font-weight: 600; white-space: nowrap; cursor: pointer; transition: color 160ms, background 160ms; }
.preferences-tabs button::after { content: ''; position: absolute; inset: auto 0 0; height: 2px; background: var(--gf-accent); transform: scaleX(0); transition: transform 160ms; }
.preferences-tabs button[aria-selected='true'] { color: var(--gf-accent); }
.preferences-tabs button[aria-selected='true']::after { transform: scaleX(1); }
.preferences-tabs button:hover { background: var(--gf-accent-soft); color: var(--gf-text-main); }
.preferences-tabs button:focus-visible { outline: 2px solid var(--gf-accent); outline-offset: -3px; }
.preferences-pages { display: flex; height: min(34rem, calc(100dvh - 11rem)); min-height: 0; overflow-x: auto; overflow-y: hidden; scroll-snap-type: x mandatory; scrollbar-width: none; overscroll-behavior-x: contain; }
.preferences-pages::-webkit-scrollbar { display: none; }
.preferences-page { flex: 0 0 100%; min-width: 0; padding: .8rem 1.3rem 1.2rem; overflow-y: auto; scroll-snap-align: start; overscroll-behavior-y: contain; }
@media (prefers-reduced-motion: reduce) { .preferences-tabs button, .preferences-tabs button::after { transition: none; } }
</style>
