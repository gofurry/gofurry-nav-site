<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  if (!password.value.trim()) {
    errorMessage.value = '请输入口令'
    return
  }
  loading.value = true
  try {
    await auth.login(password.value)
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
    <div class="w-full max-w-lg border border-[var(--line-strong)] bg-[var(--panel)]/80 p-8 shadow-[0_0_0_1px_rgba(111,224,255,0.06),0_24px_80px_rgba(0,0,0,0.45)]">
      <div class="mb-6">
        <div class="text-xs uppercase tracking-[0.3em] text-[var(--accent)]">gofurry Admin</div>
        <h1 class="mt-3 text-3xl font-semibold">运维后台登录</h1>
        <p class="mt-2 text-sm text-[var(--text-muted)]">唯一账户，只输入口令。登录态基于 HttpOnly Cookie。</p>
      </div>
      <div class="space-y-4">
        <input v-model="password" type="password" placeholder="输入口令" class="w-full border border-[var(--line)] bg-black/20 px-4 py-3 outline-none focus:border-[var(--accent)]" />
        <button class="w-full border border-[var(--accent)] bg-[var(--accent)]/10 px-4 py-3" @click="submit">{{ loading ? '登录中…' : '登录' }}</button>
      </div>
      <div v-if="errorMessage" class="mt-4 text-sm text-[var(--danger)]">{{ errorMessage }}</div>
    </div>
  </main>
</template>
