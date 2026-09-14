<template>
  <NuxtLayout name="error">
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
