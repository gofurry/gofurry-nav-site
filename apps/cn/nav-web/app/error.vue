<template>
  <NuxtLayout>
    <section class="not-found-page">
      <GoFurryGridBackground palette="nav-content" />
      <div class="not-found-page__content">
        <p class="not-found-page__eyebrow">{{ copy.eyebrow }}</p>
        <h1 class="not-found-page__code">{{ statusCode }}</h1>
        <h2 class="not-found-page__title">{{ copy.title }}</h2>
        <p class="not-found-page__description">{{ copy.description }}</p>

        <div class="not-found-page__actions">
          <button class="gf-button gf-button--primary" type="button" @click="goHome">
            {{ copy.home }}
          </button>
          <button class="gf-button gf-button--surface" type="button" @click="goBack">
            {{ copy.back }}
          </button>
        </div>
      </div>
    </section>
  </NuxtLayout>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'
import GoFurryGridBackground from '~/components/common/GoFurryGridBackground.vue'

const props = defineProps<{
  error: NuxtError
}>()

const { locale } = useI18n()
const statusCode = computed(() => props.error?.statusCode || 500)
const isNotFound = computed(() => statusCode.value === 404)
const homePath = computed(() => locale.value === 'en' ? '/en' : '/')

const copy = computed(() => {
  if (locale.value === 'en') {
    return {
      eyebrow: isNotFound.value ? 'Page Not Found' : 'Something went wrong',
      title: isNotFound.value ? 'This trail leads nowhere.' : 'The page is temporarily unavailable.',
      description: isNotFound.value
        ? 'The address may be outdated, moved, or never existed.'
        : 'The service hit an unexpected error. You can return home and try again in a moment.',
      home: 'Back Home',
      back: 'Go Back',
      seoTitle: isNotFound.value ? 'Page Not Found - GoFurry' : 'Service Error - GoFurry',
    }
  }

  return {
    eyebrow: isNotFound.value ? 'Page Not Found' : '页面暂时不可用',
    title: isNotFound.value ? '这条路径没有找到对应页面' : '服务刚刚踩到了一片松动的地板',
    description: isNotFound.value
      ? '这个地址可能已经移动、过期，或者从未存在。'
      : '页面发生了意外错误，可以先返回首页，稍后再试一次。',
    home: '返回首页',
    back: '返回上一页',
    seoTitle: isNotFound.value ? 'Page Not Found - GoFurry' : 'Service Error - GoFurry',
  }
})

useHead(() => ({
  title: copy.value.seoTitle,
  meta: [
    { name: 'robots', content: 'noindex, nofollow' },
  ],
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
</script>

<style scoped>
.not-found-page {
  position: relative;
  display: flex;
  min-height: calc(100vh - 72px);
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 7rem 1.25rem 5rem;
  color: var(--gf-text-main);
  letter-spacing: 0.018em;
}

.not-found-page::before {
  content: "";
  position: absolute;
  inset: 18% auto auto 50%;
  z-index: 1;
  width: min(34rem, 70vw);
  height: min(34rem, 70vw);
  transform: translateX(-50%);
  border-radius: 999px;
  background:
    radial-gradient(circle, rgba(249, 115, 22, 0.14), transparent 62%),
    radial-gradient(circle at 70% 24%, rgba(255, 250, 242, 0.24), transparent 42%);
  filter: blur(12px);
}

.not-found-page__content {
  position: relative;
  z-index: 2;
  width: min(100%, 46rem);
  border: 1px solid var(--gf-border);
  border-radius: 28px;
  background:
    linear-gradient(145deg, rgba(255, 250, 242, 0.94), rgba(255, 237, 213, 0.82)),
    var(--gf-surface-strong);
  box-shadow: 0 28px 80px rgba(91, 62, 28, 0.18);
  padding: clamp(2rem, 5vw, 4rem);
  text-align: center;
}

.not-found-page__eyebrow {
  margin: 0;
  color: var(--gf-accent);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.not-found-page__code {
  margin: 0.35rem 0 0;
  color: rgba(124, 45, 18, 0.94);
  font-size: clamp(5rem, 18vw, 10rem);
  font-weight: 900;
  letter-spacing: -0.045em;
  line-height: 0.96;
}

.not-found-page__title {
  margin: 1.5rem 0 0;
  font-size: clamp(1.7rem, 4vw, 2.75rem);
  font-weight: 850;
  letter-spacing: 0.02em;
  line-height: 1.22;
}

.not-found-page__description {
  margin: 1rem auto 0;
  max-width: 34rem;
  color: var(--gf-text-muted);
  font-size: 1rem;
  line-height: 1.8;
}

.not-found-page__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.8rem;
  margin-top: 2rem;
}

:global(html.dark .not-found-page__content) {
  background:
    linear-gradient(145deg, rgba(15, 23, 42, 0.92), rgba(30, 41, 59, 0.82)),
    var(--gf-surface-strong);
  box-shadow: 0 28px 80px rgba(2, 6, 23, 0.42);
}

:global(html.dark .not-found-page::before) {
  display: none;
}

:global(html.dark .not-found-page__code) {
  color: rgba(226, 232, 240, 0.94);
}

@media (max-width: 640px) {
  .not-found-page {
    min-height: calc(100vh - 56px);
    padding: 5.5rem 1rem 4rem;
  }

  .not-found-page__content {
    border-radius: 22px;
  }

  .not-found-page__actions {
    flex-direction: column;
  }

  .not-found-page__actions .gf-button {
    width: 100%;
  }
}
</style>
