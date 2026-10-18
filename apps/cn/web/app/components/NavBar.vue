<template>
  <header ref="navigation" class="gf-nav pointer-events-none absolute inset-x-0 top-0 z-[70] w-full md:fixed"
    :class="{ 'gf-nav--overlay': navOverlayDesktop }" :inert="inactive" :aria-hidden="inactive || undefined">
    <div class="gf-nav__inner relative mx-auto flex w-full max-w-[1700px] items-center justify-between gap-3 px-3 py-3 md:px-5 md:py-4 lg:px-6">
      <NuxtLink :to="localePath('/')" class="gf-nav__island gf-nav__brand pointer-events-auto flex shrink-0 items-center gap-2 px-3 md:px-4"
        aria-label="GoFurry" @click="closeMenus()">
        <img :src="logo" alt="" class="h-8 w-8" />
        <span class="gf-nav__brand-text">GoFurry</span>
      </NuxtLink>
      <div class="gf-nav__island gf-nav__actions pointer-events-auto flex shrink-0 items-center gap-0.5 px-2 lg:gap-1 lg:px-3">
        <nav class="gf-nav__links hidden items-center gap-0.5 md:flex lg:gap-1" :aria-label="t('navbar.expandNav')">
          <template v-for="link in navLinks" :key="link.label">
            <a v-if="link.external" :href="link.href" target="_blank" rel="noopener noreferrer"
              class="gf-nav__link inline-flex items-center justify-center" :aria-label="link.label" :title="link.label">
              <component :is="link.icon" :size="20" weight="regular" aria-hidden="true" />
            </a>
            <NuxtLink v-else :to="link.to" class="gf-nav__link inline-flex items-center justify-center"
              :class="{ 'gf-nav__link--active': isActive(link) }" :aria-current="isActive(link) ? 'page' : undefined"
              :aria-label="link.label" :title="link.label" @click="closeMenus()">
              <component :is="link.icon" :size="20" weight="regular" aria-hidden="true" />
            </NuxtLink>
          </template>
        </nav>
        <span class="gf-nav__separator mx-1 hidden md:block" aria-hidden="true" />
        <a href="https://github.com/gofurry/gofurry-nav-site" target="_blank" rel="noopener noreferrer"
          class="gf-nav__github-link hidden items-center justify-center md:inline-flex" aria-label="GitHub" title="GitHub">
          <PhGithubLogo :size="20" weight="regular" aria-hidden="true" />
        </a>
        <button type="button" class="gf-nav__icon-button inline-flex items-center justify-center"
          :aria-label="themeToggleLabel" :title="themeToggleLabel" @click="toggleThemeIcon">
          <PhSun v-if="themeStore.theme === 'light'" :size="20" weight="regular" aria-hidden="true" />
          <PhMoon v-else :size="20" weight="regular" aria-hidden="true" />
        </button>
        <button type="button" class="gf-nav__icon-button gf-nav__mode-button hidden items-center justify-center md:inline-flex"
          :aria-label="t('navbar.mode')" :title="t('navbar.mode')" @click="openModeModal">
          <PhGearSix :size="20" weight="regular" aria-hidden="true" />
        </button>
        <button type="button" class="gf-nav__icon-button gf-nav__language inline-flex items-center justify-center"
          :aria-label="currentLang === 'zh' ? 'Switch to English' : '切换至中文'" :title="currentLang === 'zh' ? 'English' : '中文'"
          @click="switchLang">
          {{ currentLang === 'zh' ? 'EN' : '中' }}
        </button>
        <button ref="mobileMenuToggle" type="button" class="gf-nav__mobile-toggle inline-flex items-center justify-center md:hidden"
          :class="{ 'gf-nav__mobile-toggle--open': mobileMenuOpen }" :aria-expanded="mobileMenuOpen" :aria-controls="menuId"
          :aria-label="t(mobileMenuOpen ? 'navbar.closeNav' : 'navbar.expandNav')" @click="mobileMenuOpen = !mobileMenuOpen">
          <PhList v-if="!mobileMenuOpen" :size="20" weight="regular" aria-hidden="true" />
          <PhX v-else :size="20" weight="regular" aria-hidden="true" />
        </button>
      </div>
    </div>
    <Transition name="mobile-menu" @before-leave="deactivateMenu" @before-enter="activateMenu">
      <div v-if="mobileMenuOpen" :id="menuId" ref="menuPanel"
        class="gf-nav__mobile-panel pointer-events-auto absolute left-3 right-3 top-full z-[90] -mt-1 max-h-[calc(100dvh-6rem)] overflow-y-auto overscroll-contain p-4 md:hidden">
        <nav class="grid grid-cols-4 gap-2" :aria-label="t('navbar.expandNav')">
          <template v-for="link in navLinks" :key="link.label">
            <a v-if="link.external" :href="link.href" target="_blank" rel="noopener noreferrer"
              class="gf-nav__mobile-link flex items-center justify-center py-3" :aria-label="link.label" :title="link.label" @click="closeMenus(true)">
              <component :is="link.icon" :size="20" weight="regular" aria-hidden="true" />
            </a>
            <NuxtLink v-else :to="link.to" class="gf-nav__mobile-link flex items-center justify-center py-3"
              :class="{ 'gf-nav__mobile-link--active': isActive(link) }" :aria-current="isActive(link) ? 'page' : undefined"
              :aria-label="link.label" :title="link.label" @click="closeMenus(true)">
              <component :is="link.icon" :size="20" weight="regular" aria-hidden="true" />
            </NuxtLink>
          </template>
        </nav>
        <div class="gf-nav__mobile-divider mt-3 flex flex-col gap-2 pt-3">
          <a href="https://github.com/gofurry/gofurry-nav-site" target="_blank" rel="noopener noreferrer"
            class="gf-nav__mobile-link flex items-center gap-3 px-3 py-3" @click="closeMenus(true)">
            <PhGithubLogo :size="20" weight="regular" aria-hidden="true" />GitHub
          </a>
          <button type="button" class="gf-nav__mobile-action flex items-center justify-between gap-3 px-3 py-3" @click="openModeModal">
            <span class="flex items-center gap-3"><PhGearSix :size="20" weight="regular" aria-hidden="true" />{{ t('navbar.mode') }}</span>
            <span class="gf-nav__mobile-action-value truncate">{{ mode || '--' }}</span>
          </button>
        </div>
      </div>
    </Transition>
    <ModeSettingModal :show="showModeModal" :mode="mode" :return-focus="modeReturnFocus"
      @cancel="showModeModal = false" @save="saveMode" />
  </header>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch, type Component } from 'vue'
import { PhBinoculars, PhFlask, PhGameController, PhGearSix, PhGithubLogo, PhHouse, PhList, PhMoon, PhSun, PhX } from '@phosphor-icons/vue'
import { useThemeStore } from '@/stores/theme'
import { useI18n } from 'vue-i18n'
import logo from '@/assets/svgs/logo-mini.svg'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const switchLocalePath = useSwitchLocalePath()
const themeStore = useThemeStore()
defineProps<{ navOverlayDesktop?: boolean, inactive?: boolean }>()
const emit = defineEmits<{ busy: [value: boolean] }>()
const menuId = useId()
const navigation = ref<HTMLElement | null>(null)
const menuPanel = ref<HTMLElement | null>(null)
const mobileMenuToggle = ref<HTMLElement | null>(null)
const modeReturnFocus = ref<HTMLElement | null>(null)
const showModeModal = ref(false)
const mobileMenuOpen = ref(false)
const mode = ref('')
let stopModeSubscription: (() => void) | null = null
let mobileQuery: MediaQueryList | null = null
const currentLang = computed<'zh' | 'en'>(() => locale.value === 'en' ? 'en' : 'zh')
type NavLink = { label: string, icon: Component, to?: string, activePath?: string, href?: string, external?: boolean }
const navLinks = computed<NavLink[]>(() => [
  { label: t('sidebar.nav'), icon: PhHouse, to: localePath('/'), activePath: '/' },
  { label: t('sidebar.games'), icon: PhGameController, to: localePath('/games'), activePath: '/games' },
  { label: t('sidebar.insights'), icon: PhBinoculars, to: localePath('/insights'), activePath: '/insights' },
  { label: currentLang.value === 'zh' ? '深度兽研' : 'DeepFurry', icon: PhFlask, href: 'https://www.deepfurry.com', external: true },
])
const themeToggleLabel = computed(() => currentLang.value === 'zh' ? '切换明暗主题图标' : 'Toggle theme icon')
const normalizeRoutePath = (path: string) => path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/'
function isActive(link: NavLink) {
  if (!link.to) return false
  const current = normalizeRoutePath(route.path)
  return link.activePath === '/' ? current === '/' : current === link.activePath || current.startsWith(`${link.activePath}/`)
}
function closeMenus(restoreFocus = false) {
  if (restoreFocus && mobileMenuOpen.value) mobileMenuToggle.value?.focus({ preventScroll: true })
  mobileMenuOpen.value = false
}
function activateMenu(element: Element) {
  const panel = element as HTMLElement
  panel.inert = false
  panel.removeAttribute('aria-hidden')
}
function deactivateMenu(element: Element) {
  const panel = element as HTMLElement
  if (panel.contains(document.activeElement)) mobileMenuToggle.value?.focus({ preventScroll: true })
  panel.inert = true
  panel.setAttribute('aria-hidden', 'true')
}
function handleKey(event: KeyboardEvent) {
  if (event.key === 'Escape' && mobileMenuOpen.value && !showModeModal.value) {
    event.preventDefault()
    closeMenus(true)
  }
}
function outsideClick(event: MouseEvent) {
  if (event.target instanceof Element && event.target.closest('.gf-modal-backdrop')) return
  // A toggle replaces its icon before document bubbling; the original path is stable.
  if (!showModeModal.value && navigation.value && !event.composedPath().includes(navigation.value)) closeMenus()
}
async function handleBreakpoint() {
  const focused = document.activeElement instanceof HTMLElement && navigation.value?.contains(document.activeElement)
  const modalWasOpen = showModeModal.value
  const menuWasOpen = mobileMenuOpen.value
  showModeModal.value = false
  closeMenus()
  await nextTick()
  if (focused || modalWasOpen || menuWasOpen) navigation.value?.querySelector<HTMLElement>('.gf-nav__brand')?.focus({ preventScroll: true })
}
onMounted(() => {
  mode.value = readMode()
  themeStore.initTheme()
  stopModeSubscription = subscribeModeChange(({ mode: next }) => { mode.value = next })
  mobileQuery = window.matchMedia('(max-width: 767px)')
  mobileQuery.addEventListener('change', handleBreakpoint)
  document.addEventListener('keydown', handleKey)
  document.addEventListener('click', outsideClick)
})
onUnmounted(() => {
  stopModeSubscription?.()
  mobileQuery?.removeEventListener('change', handleBreakpoint)
  document.removeEventListener('keydown', handleKey)
  document.removeEventListener('click', outsideClick)
})
watch(() => route.fullPath, () => { closeMenus(true); showModeModal.value = false })
watch([mobileMenuOpen, showModeModal], () => emit('busy', mobileMenuOpen.value || showModeModal.value), { flush: 'sync' })
function saveMode(value: string) { showModeModal.value = false; writeMode(value.trim().slice(0, 32)) }
async function switchLang() {
  const nextPath = switchLocalePath(currentLang.value === 'zh' ? 'en' : 'zh')
  if (nextPath) {
    const target = router.resolve(nextPath)
    await router.push({ path: target.path, query: route.query, hash: route.hash })
  }
  closeMenus(true)
}
function openModeModal(event: MouseEvent) {
  modeReturnFocus.value = event.currentTarget as HTMLElement
  showModeModal.value = true
}
function toggleThemeIcon() { themeStore.setTheme(themeStore.theme === 'light' ? 'dark' : 'light') }
</script>
