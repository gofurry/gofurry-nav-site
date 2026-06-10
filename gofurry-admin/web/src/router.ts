import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'
import BootstrapView from './views/BootstrapView.vue'
import LoginView from './views/LoginView.vue'
import ShellView from './views/ShellView.vue'
import ResourceView from './views/ResourceView.vue'
import GameCollectView from './views/GameCollectView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', component: BootstrapView },
    { path: '/login', component: LoginView },
    {
      path: '/',
      component: ShellView,
      children: [
        { path: '', redirect: '/nav/sayings' },
        { path: 'game/collect', component: GameCollectView },
        { path: ':section/:resource', component: ResourceView },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.loadState()

  if (!auth.initialized && to.path !== '/setup') {
    return '/setup'
  }
  if (auth.initialized && !auth.authenticated && to.path !== '/login') {
    return '/login'
  }
  if (auth.initialized && auth.authenticated && (to.path === '/login' || to.path === '/setup')) {
    return '/nav/sayings'
  }
  return true
})

export default router
