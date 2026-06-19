<template>
  <div class="gf-app-shell relative flex min-h-screen flex-col transition-colors duration-500">
    <div :class="navBarWrapperClass">
      <NavBar :nav-overlay-desktop="isNavPage" />
    </div>
    <main class="relative flex min-w-0 flex-1 flex-col">
      <slot />
      <div v-if="showFooter" class="relative mt-auto">
        <div class="pointer-events-none absolute inset-x-0 top-0 z-10 h-4 -translate-y-1/2 bg-black/30"></div>
        <Footer />
      </div>
    </main>
    <ClientOnly>
      <MobileBottomTabBar />
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import { NAV_PAGE_REVEAL_EVENT } from '@/utils/navPageReveal'
import MobileBottomTabBar from '@/components/common/MobileBottomTabBar.vue'

const route = useRoute()
const navPageRevealed = ref(true)
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isNavPage = computed(() => normalizedPath.value === '/')
const navBarWrapperClass = computed(() => (
  isNavPage.value
    ? 'md:absolute md:inset-x-0 md:top-0 md:z-[70] md:w-full'
    : ''
))

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
