<script setup lang="ts">
import {
  Activity,
  AlertTriangle,
  CloudCog,
  GitBranch,
  LayoutDashboard,
  LogOut,
  Network,
  RefreshCw,
  Server,
  ShieldCheck,
} from 'lucide-vue-next'
import { computed, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useSessionStore } from './stores/session'

const session = useSessionStore()
const route = useRoute()
const passcode = ref('')
const loginError = ref('')

const navItems = [
  { to: '/', label: '总览', icon: LayoutDashboard },
  { to: '/nodes', label: '节点', icon: Server },
  { to: '/services', label: '服务', icon: Activity },
  { to: '/alerts', label: '告警', icon: AlertTriangle },
  { to: '/sync', label: '同步', icon: GitBranch },
  { to: '/peers', label: 'Peer', icon: Network },
  { to: '/deployments', label: '部署', icon: CloudCog },
]

const pageTitle = computed(() => {
  const current = navItems.find((item) => route.path === item.to || route.path.startsWith(`${item.to}/`))
  return current?.label || '节点'
})

onMounted(() => {
  void session.refresh()
})

async function submitLogin() {
  loginError.value = ''
  try {
    await session.login(passcode.value)
    passcode.value = ''
  } catch (error) {
    loginError.value = error instanceof Error ? error.message : '登录失败'
  }
}
</script>

<template>
  <main v-if="session.authenticated" class="ops-shell min-h-screen text-[var(--ops-ink)]">
    <aside class="ops-sidebar fixed inset-y-0 left-0 hidden w-64 lg:block">
      <div class="flex h-16 items-center gap-3 border-b border-[var(--ops-border)] px-5">
        <div class="brand-mark flex size-9 items-center justify-center">
          <ShieldCheck class="size-5" />
        </div>
        <div>
          <p class="text-sm font-semibold">GoFurry Ops</p>
          <p class="text-xs text-[var(--ops-muted)]">Center Console</p>
        </div>
      </div>
      <nav class="space-y-1 p-3">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-link"
          :class="{ active: route.path === item.to || route.path.startsWith(`${item.to}/`) }"
        >
          <component :is="item.icon" class="size-4" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <section class="lg:pl-64">
      <header
        class="ops-header sticky top-0 z-20 flex min-h-16 items-center justify-between px-4 md:px-6"
      >
        <div>
          <p class="text-xs font-medium uppercase tracking-[0.18em] text-[var(--ops-muted)]">Ops Center</p>
          <h1 class="text-xl font-semibold md:text-2xl">{{ pageTitle }}</h1>
        </div>
        <div class="flex items-center gap-2">
          <button class="icon-button" title="刷新认证状态" @click="session.refresh()">
            <RefreshCw class="size-4" />
          </button>
          <button class="icon-button" title="退出" @click="session.logout()">
            <LogOut class="size-4" />
          </button>
        </div>
      </header>

      <nav class="ops-mobile-nav flex gap-2 overflow-x-auto px-4 py-2 lg:hidden">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-pill"
          :class="{ active: route.path === item.to || route.path.startsWith(`${item.to}/`) }"
        >
          <component :is="item.icon" class="size-4" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="mx-auto w-full max-w-7xl px-4 py-5 md:px-6">
        <RouterView />
      </div>
    </section>
  </main>

  <main v-else class="ops-shell flex min-h-screen items-center justify-center p-4 text-[var(--ops-ink)]">
    <form class="panel login-panel w-full max-w-sm space-y-5 p-6" @submit.prevent="submitLogin">
      <div class="space-y-2">
        <div class="brand-mark flex size-10 items-center justify-center">
          <ShieldCheck class="size-5" />
        </div>
        <h1 class="text-xl font-semibold">GoFurry Ops Center</h1>
        <p class="text-sm text-[var(--ops-muted)]">输入控制台口令</p>
      </div>
      <label class="block space-y-2">
        <span class="text-sm font-medium">Passcode</span>
        <input v-model="passcode" class="input" type="password" autocomplete="current-password" />
      </label>
      <p v-if="loginError || session.error" class="text-sm text-[var(--ops-red)]">{{ loginError || session.error }}</p>
      <button class="button-primary w-full" type="submit" :disabled="session.loading">
        {{ session.loading ? '登录中' : '登录' }}
      </button>
    </form>
  </main>
</template>
