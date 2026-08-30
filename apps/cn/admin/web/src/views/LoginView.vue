<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  if (!username.value.trim() || !password.value) {
    errorMessage.value = '请输入用户名和口令'
    return
  }
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/nav/sayings')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center px-4">
    <div class="w-full max-w-lg border-l-2 border-[var(--accent)] bg-[var(--panel)] p-8">
      <div class="mb-6">
        <div class="text-xs uppercase tracking-[0.3em] text-[var(--accent)]">gofurry Admin</div>
        <h1 class="mt-3 text-3xl font-semibold">运维后台登录</h1>
        <p class="mt-2 text-sm text-[var(--text-muted)]">使用内部 Admin 账号登录；登录态基于 HttpOnly Cookie。</p>
      </div>
      <div class="space-y-4">
        <input v-model="username" autocomplete="username" placeholder="用户名" class="ui-control w-full px-4 py-3" />
        <input v-model="password" type="password" placeholder="输入口令" class="ui-control w-full px-4 py-3" />
        <button class="ui-button ui-button--primary w-full px-4 py-3" @click="submit">{{ loading ? '登录中…' : '登录' }}</button>
      </div>
      <div v-if="errorMessage" class="mt-4 text-sm text-[var(--danger)]">{{ errorMessage }}</div>
    </div>
  </main>
</template>
