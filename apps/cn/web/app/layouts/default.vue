<template>
  <div class="gf-app-shell relative isolate flex min-h-screen flex-col">
    <PublicPageBackground />
    <NavBar :nav-overlay-desktop="isNavPage" :inactive="mobileNavigationActive" @busy="navigationBusy = $event" />
    <main class="relative flex min-w-0 flex-1 flex-col pb-20 md:pb-0">
      <slot />
      <div v-if="showFooter" class="gf-footer-shell relative mt-auto">
        <Footer />
      </div>
    </main>
    <ClientOnly>
      <MobileBottomTabBar :navigation-busy="navigationBusy" @handoff="mobileNavigationActive = $event" />
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import { NAV_PAGE_REVEAL_EVENT } from '@/utils/navPageReveal'
import MobileBottomTabBar from '@/components/common/MobileBottomTabBar.vue'
import PublicPageBackground from '@/components/common/PublicPageBackground.vue'

const route = useRoute()
const navPageRevealed = ref(true)
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isNavPage = computed(() => normalizedPath.value === '/')
const mobileNavigationActive = ref(false)
const navigationBusy = ref(false)

const showFooter = computed(() => {
  if (isNavPage.value) {
    return navPageRevealed.value
  }

  return true
})

function handleNavPageReveal(event: Event) {
  const customEvent = event as CustomEvent<{ visible?: boolean }>
  navPageRevealed.value = customEvent.detail?.visible ?? true
}

watch(
  () => route.path,
  () => {
    navPageRevealed.value = isNavPage.value
      ? import.meta.client && window.innerWidth < 768
      : true
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener(NAV_PAGE_REVEAL_EVENT, handleNavPageReveal)
})

onUnmounted(() => {
  window.removeEventListener(NAV_PAGE_REVEAL_EVENT, handleNavPageReveal)
})
</script>
