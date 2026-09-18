<template>
  <section class="gf-modal__section hero-preferences" data-hero-preferences>
    <div class="gf-modal__copy"><h3 class="gf-modal__label">{{ label('首屏背景来源', 'Hero background source') }}</h3><p class="gf-modal__help">{{ label('预览后点击“保存”应用，取消不会改变当前首屏。', 'Preview, then Save to apply. Cancel keeps the current Hero.') }}</p></div>
    <div class="hero-preferences__sources" role="group" :aria-label="label('首屏背景来源', 'Hero background source')">
      <button v-for="mode in (['random', 'fixed', 'local'] as const)" :key="mode" type="button" :aria-pressed="draft.mode === mode" @click="draft.mode = mode; error = ''">{{ mode === 'random' ? label('云端随机', 'Random cloud') : mode === 'fixed' ? label('固定云端背景', 'Fixed cloud') : label('本地文件夹', 'Local folder') }}</button>
    </div>
    <p v-if="draft.mode === 'random'" class="gf-modal__help">{{ label('每次打开首页独立随机选择桌面与手机背景。', 'Each visit independently selects a random desktop and mobile background.') }}</p>
    <template v-if="draft.mode === 'fixed'">
      <div class="hero-preferences__sources" role="group" :aria-label="label('背景设备', 'Hero viewport')">
        <button v-for="item in (['desktop', 'mobile'] as const)" :key="item" type="button" :aria-pressed="viewport === item" @click="viewport = item">{{ item === 'desktop' ? label('桌面背景', 'Desktop') : label('手机背景', 'Mobile') }}</button>
      </div>
      <HeroCatalogPicker :key="viewport" :variant="viewport" :selected-id="viewport === 'desktop' ? draft.desktopId : draft.mobileId" :active="active" @select="select" />
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
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import HeroCatalogPicker from './HeroCatalogPicker.vue'
import { clearCustomNavHeaderBackgroundDirectory, loadCustomNavHeaderBackgroundMeta, pickCustomNavHeaderBackgroundDirectory, saveCustomNavHeaderBackgroundDirectory, supportsCustomNavHeaderBackground, type CustomNavHeaderBackgroundSelection } from '~/utils/customNavHeaderBackground'
defineProps<{ active: boolean }>()
const { locale, t } = useI18n()
const label = (zh: string, en: string) => locale.value === 'en' ? en : zh
const settings = useHeroPreferences()
const draft = ref({ ...settings.preference.value })
const viewport = ref<'desktop' | 'mobile'>('desktop')
const error = ref(''), folderName = ref(loadCustomNavHeaderBackgroundMeta().folderName)
const supportsPicker = supportsCustomNavHeaderBackground()
let pending: CustomNavHeaderBackgroundSelection | null = null, remove = false
function select(id: string | null) { if (viewport.value === 'desktop') draft.value.desktopId = id; else draft.value.mobileId = id }
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

<style scoped>
.hero-preferences { display: grid; gap: .85rem; }
.hero-preferences__sources { display: flex; gap: .4rem; flex-wrap: wrap; }
.hero-preferences__sources button { padding: .55rem .8rem; border: 1px solid var(--gf-border-strong); border-radius: .5rem; background: transparent; color: var(--gf-text-main); cursor: pointer; }
.hero-preferences__sources button[aria-pressed='true'], .hero-preferences__sources button:hover { background: var(--gf-accent-soft); }
.hero-preferences__sources button:focus-visible { outline: 2px solid var(--gf-accent); outline-offset: 2px; }
</style>
