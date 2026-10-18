<template>
  <section class="gf-modal__section" data-hero-preferences>
    <div class="gf-modal__copy"><h3 class="gf-modal__label">{{ label('首屏背景', 'Hero background') }}</h3><p class="gf-modal__help">{{ label('选择首页首屏使用的背景来源。', 'Choose the background source for your homepage Hero.') }}</p></div>
    <div class="preferences-sources" role="group" :aria-label="label('首屏背景来源', 'Hero background source')">
      <button v-for="mode in (['random', 'fixed', 'local'] as const)" :key="mode" type="button" :aria-pressed="draft.mode === mode"
        :aria-label="mode === 'fixed' ? label('固定云端背景', 'Fixed cloud background') : undefined"
        @click="draft.mode = mode; error = ''">{{ mode === 'random' ? label('云端随机', 'Random cloud') : mode === 'fixed' ? label('固定云端', 'Fixed cloud') : label('本地文件夹', 'Local folder') }}</button>
    </div>
    <p v-if="draft.mode === 'random'" class="gf-modal__help">{{ label('每次打开首页独立随机选择桌面与手机背景。', 'Each visit independently selects a random desktop and mobile background.') }}</p>
    <template v-if="draft.mode === 'fixed'">
      <h4 class="gf-modal__label">{{ label('固定云端背景', 'Fixed cloud background') }}</h4>
      <div class="preferences-tabs preferences-tabs--compact" role="tablist" :aria-label="label('背景设备', 'Hero viewport')">
        <button v-for="item in viewports" :id="`${fieldId}-tab-${item}`" :key="item" type="button" role="tab"
          :aria-selected="viewport === item" :aria-controls="`${fieldId}-panel-${item}`" :tabindex="viewport === item ? 0 : -1"
          @click="viewport = item" @keydown="viewportKeydown($event, item)">{{ item === 'desktop' ? label('桌面', 'Desktop') : label('手机', 'Mobile') }}</button>
      </div>
      <div v-for="item in viewports" :id="`${fieldId}-panel-${item}`" :key="item" v-show="viewport === item" role="tabpanel" :aria-labelledby="`${fieldId}-tab-${item}`">
        <HeroCatalogPicker :variant="item" :selected-id="item === 'desktop' ? draft.desktopId : draft.mobileId" :active="active && viewport === item" @select="select(item, $event)" />
      </div>
    </template>
    <template v-if="draft.mode === 'local'">
      <p class="gf-modal__help">{{ folderName ? t('navbar.customNavHeaderBgSelected', { name: folderName }) : t('navbar.customNavHeaderBgEmpty') }}</p>
      <div class="gf-modal__actions">
        <button type="button" class="gf-button gf-button--surface" :disabled="!supportsPicker" @click="pick">{{ t('navbar.customNavHeaderBgPick') }}</button>
        <button type="button" class="gf-button gf-button--ghost" :disabled="!folderName" @click="clear">{{ t('navbar.customNavHeaderBgClear') }}</button>
      </div>
      <p class="gf-modal__footnote">{{ t('navbar.customNavHeaderBgDesc') }}</p>
      <p v-if="!supportsPicker" class="gf-modal__help">{{ t('navbar.customNavHeaderBgUnsupported') }}</p>
    </template>
    <p v-if="error" role="alert" class="gf-modal__help">{{ error }}</p>
  </section>
</template>

<script setup lang="ts">
import { ref, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import HeroCatalogPicker from './HeroCatalogPicker.vue'
import { clearCustomNavHeaderBackgroundDirectory, loadCustomNavHeaderBackgroundMeta, pickCustomNavHeaderBackgroundDirectory, saveCustomNavHeaderBackgroundDirectory, supportsCustomNavHeaderBackground, type CustomNavHeaderBackgroundSelection } from '~/utils/customNavHeaderBackground'
defineProps<{ active: boolean }>()
const { locale, t } = useI18n()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const settings = useHeroPreferences()
const draft = ref({ ...settings.preference.value })
const fieldId = useId()
const viewports = ['desktop', 'mobile'] as const
const viewport = ref<'desktop' | 'mobile'>('desktop')
const error = ref(''), folderName = ref(loadCustomNavHeaderBackgroundMeta().folderName)
const supportsPicker = supportsCustomNavHeaderBackground()
let pending: CustomNavHeaderBackgroundSelection | null = null, remove = false
function select(item: 'desktop' | 'mobile', id: string | null) { if (item === 'desktop') draft.value.desktopId = id; else draft.value.mobileId = id }
function viewportKeydown(event: KeyboardEvent, item: 'desktop' | 'mobile') {
  const next = event.key === 'Home' ? 'desktop' : event.key === 'End' ? 'mobile'
    : ['ArrowLeft', 'ArrowRight'].includes(event.key) ? item === 'desktop' ? 'mobile' : 'desktop' : null
  if (!next) return
  event.preventDefault()
  event.stopPropagation()
  viewport.value = next
  document.getElementById(`${fieldId}-tab-${next}`)?.focus()
}
async function pick() {
  try {
    const selection = await pickCustomNavHeaderBackgroundDirectory()
    if (selection) { pending = selection; folderName.value = selection.folderName; remove = false; error.value = '' }
  } catch { error.value = label('无法读取文件夹，请重新选择。', 'Cannot read the folder. Please choose it again.') }
}
function clear() { pending = null; remove = true; folderName.value = ''; draft.value.mode = 'random' }
function validate() {
  if (draft.value.mode === 'local' && !folderName.value) { error.value = label('请先选择包含图片的本地文件夹。', 'Choose a local folder containing images first.'); return false }
  return true
}
async function save() {
  if (!validate()) return false
  try {
    if (remove) await clearCustomNavHeaderBackgroundDirectory()
    else if (pending) await saveCustomNavHeaderBackgroundDirectory(pending)
    settings.save(draft.value, Boolean(pending || remove))
    return true
  } catch { error.value = label('无法保存首屏设置，请重试。', 'Cannot save Hero preferences. Please retry.'); return false }
}
defineExpose({ save, validate })
</script>
