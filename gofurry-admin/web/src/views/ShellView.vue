<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { resources } from '../resources'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const navResources = computed(() => resources.filter((item) => item.section === 'nav'))
const gameResources = computed(() => resources.filter((item) => item.section === 'game'))

async function logout() {
  await auth.logout()
  router.push('/login')
}

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <div class="min-h-screen md:grid md:grid-cols-[260px_1fr]">
    <aside class="border-b border-[var(--line)] bg-[var(--bg-muted)]/80 p-5 md:min-h-screen md:border-b-0 md:border-r">
      <div class="mb-8">
        <div class="text-xs uppercase tracking-[0.3em] text-[var(--accent)]">gofurry</div>
        <div class="mt-2 text-2xl font-semibold">Admin</div>
      </div>

      <div class="space-y-6">
        <section>
          <div class="mb-2 text-xs uppercase tracking-[0.25em] text-[var(--text-muted)]">导航库</div>
          <div class="space-y-1">
            <RouterLink
              to="/nav/collect"
              class="block border px-3 py-2 text-sm"
              :class="isActive('/nav/collect') ? 'border-[var(--accent)] bg-[var(--accent)]/10' : 'border-transparent text-[var(--text-muted)] hover:border-[var(--line)] hover:text-[var(--text)]'"
            >
              采集观测
            </RouterLink>
            <RouterLink
              v-for="item in navResources"
              :key="item.key"
              :to="`/${item.section}/${item.key}`"
              class="block border px-3 py-2 text-sm"
              :class="isActive(`/${item.section}/${item.key}`) ? 'border-[var(--accent)] bg-[var(--accent)]/10' : 'border-transparent text-[var(--text-muted)] hover:border-[var(--line)] hover:text-[var(--text)]'"
            >
              {{ item.title }}
            </RouterLink>
          </div>
        </section>

        <section>
          <div class="mb-2 text-xs uppercase tracking-[0.25em] text-[var(--text-muted)]">游戏库</div>
          <div class="space-y-1">
            <RouterLink
              to="/game/collect"
              class="block border px-3 py-2 text-sm"
              :class="isActive('/game/collect') ? 'border-[var(--accent)] bg-[var(--accent)]/10' : 'border-transparent text-[var(--text-muted)] hover:border-[var(--line)] hover:text-[var(--text)]'"
            >
              采集观测
            </RouterLink>
            <RouterLink
              v-for="item in gameResources"
              :key="item.key"
              :to="`/${item.section}/${item.key}`"
              class="block border px-3 py-2 text-sm"
              :class="isActive(`/${item.section}/${item.key}`) ? 'border-[var(--accent)] bg-[var(--accent)]/10' : 'border-transparent text-[var(--text-muted)] hover:border-[var(--line)] hover:text-[var(--text)]'"
            >
              {{ item.title }}
            </RouterLink>
          </div>
        </section>

        <section>
          <div class="mb-2 text-xs uppercase tracking-[0.25em] text-[var(--text-muted)]">系统</div>
          <button class="w-full border border-[var(--line)] px-3 py-2 text-left text-sm text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--text)]" @click="logout">登出</button>
        </section>
      </div>
    </aside>

    <section class="p-4 md:p-6">
      <RouterView />
    </section>
  </div>
</template>
