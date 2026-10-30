<template>
  <div class="gf-app-shell relative isolate flex min-h-screen flex-col">
    <PublicPageBackground />
    <nav v-if="transitionMode && !isNavPage" class="relative px-4 py-4 sm:px-8" aria-label="GoFurry">
      <NuxtLink :to="localePath('/')" no-prefetch class="gf-button gf-button--stationary">GoFurry</NuxtLink>
    </nav>
    <NavBar v-else-if="!transitionMode" :nav-overlay-desktop="isNavPage" :inactive="mobileNavigationActive" @busy="navigationBusy = $event" />
    <main class="relative flex min-w-0 flex-1 flex-col" :class="transitionMode ? '' : 'pb-20 md:pb-0'">
      <slot />
      <div class="gf-footer-shell relative mt-auto">
        <TransitionFooter v-if="transitionMode" />
        <Footer v-else />
      </div>
    </main>
    <ClientOnly v-if="!transitionMode">
      <MobileBottomTabBar :navigation-busy="navigationBusy" @handoff="mobileNavigationActive = $event" />
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import MobileBottomTabBar from '@/components/common/MobileBottomTabBar.vue'
import PublicPageBackground from '@/components/common/PublicPageBackground.vue'
import TransitionFooter from '~/components/transition/TransitionFooter.vue'

const route = useRoute()
const transitionMode = useTransitionMode()
const localePath = useLocalePath()
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isNavPage = computed(() => normalizedPath.value === '/')
const mobileNavigationActive = ref(false)
const navigationBusy = ref(false)

</script>
