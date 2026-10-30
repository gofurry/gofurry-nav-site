<template>
  <div v-if="transitionMode" class="gf-app-shell flex min-h-screen flex-col">
    <main class="mx-auto flex w-full max-w-3xl flex-1 flex-col items-center justify-center gap-6 px-4 py-12 text-center">
      <h1>GoFurry</h1>
      <p>{{ transitionNotice }}</p>
      <button class="gf-button gf-button--stationary" @click="goHome">GoFurry</button>
    </main>
    <div class="gf-footer-shell"><TransitionFooter /></div>
  </div>
  <NuxtLayout v-else name="error">
    <ErrorExperience
      :status-code="statusCode"
      :variant="variant"
      @home="goHome"
      @back="goBack"
      @retry="retry"
    />
  </NuxtLayout>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'
import ErrorExperience from '@/components/error/ErrorExperience.vue'
import TransitionFooter from '~/components/transition/TransitionFooter.vue'
import { transitionNotice } from '#shared/utils/transitionMode'
const transitionMode = useTransitionMode()

const props = defineProps<{ error: NuxtError }>()
const { locale } = useI18n()
const statusCode = computed(() => props.error?.statusCode ?? 500)
const variant = computed<'notFound' | 'serverError' | 'generic'>(() => {
  if (statusCode.value === 404) return 'notFound'
  if (statusCode.value >= 500 && statusCode.value < 600) return 'serverError'
  return 'generic'
})
const homePath = computed(() => locale.value === 'en' ? '/en' : '/')

useHead(() => ({
  title: `${statusCode.value} - GoFurry`,
  meta: [{ name: 'robots', content: 'noindex, nofollow' }],
}))

async function goHome() {
  await clearError({ redirect: homePath.value })
}

async function goBack() {
  await clearError()

  if (import.meta.client && window.history.length > 1) {
    window.history.back()
    return
  }

  await navigateTo(homePath.value)
}

function retry() {
  if (import.meta.client) window.location.reload()
}
</script>
