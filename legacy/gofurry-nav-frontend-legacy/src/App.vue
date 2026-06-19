<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <NavBar />
    <main class="relative flex-1 flex flex-col min-w-0">
      <RouterView v-slot="{ Component }">
        <component :is="Component" class="min-h-0 flex-1" />
      </RouterView>
      <div v-if="showFooter" class="relative mt-auto">
        <div class="pointer-events-none absolute inset-x-0 top-0 z-10 h-4 -translate-y-1/2 bg-black/30"></div>
        <Footer />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import Footer from '@/components/Footer.vue'
import { NAV_PAGE_REVEAL_EVENT } from '@/utils/navPageReveal'

const route = useRoute()
const navPageRevealed = ref(true)

const showFooter = computed(() => {
  if (route.path === '/nav') {
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
  (path) => {
    navPageRevealed.value = path === '/nav'
      ? window.innerWidth < 768
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
