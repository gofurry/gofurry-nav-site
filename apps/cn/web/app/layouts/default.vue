<template>
  <div class="gf-app-shell relative isolate flex min-h-screen flex-col">
    <PublicPageBackground />
    <NavBar :nav-overlay-desktop="isNavPage" :inactive="mobileNavigationActive" @busy="navigationBusy = $event" />
    <main class="relative flex min-w-0 flex-1 flex-col pb-20 md:pb-0">
      <slot />
      <div class="gf-footer-shell relative mt-auto">
        <Footer />
      </div>
    </main>
    <ClientOnly>
      <MobileBottomTabBar :navigation-busy="navigationBusy" @handoff="mobileNavigationActive = $event" />
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import MobileBottomTabBar from '@/components/common/MobileBottomTabBar.vue'
import PublicPageBackground from '@/components/common/PublicPageBackground.vue'

const route = useRoute()
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isNavPage = computed(() => normalizedPath.value === '/')
const mobileNavigationActive = ref(false)
const navigationBusy = ref(false)

</script>
