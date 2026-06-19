import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: { path: '/nav', query: { mode: 'sfw' } } },
  { path: '/home', redirect: { path: '/nav', query: { mode: 'sfw' } } },
  { path: '/panel', component: () => import('@/pages/nav/Dashboard.vue') },
  { path: '/about', component: () => import('@/pages/other/About.vue') },
  { path: '/nav', component: () => import('@/pages/nav/NavPage.vue') },
  { path: '/updates', component: () => import('@/pages/other/Updates.vue') },
  { path: '/games', component: () => import('@/pages/game/GamesPage.vue') },
  { path: '/site/:id', component: () => import('@/pages/nav/Site.vue') },
  { path: '/games/:id', component: () => import('@/pages/game/GameDetail.vue') },
  { path: '/games/search', component: () => import('@/pages/game/GameSearch.vue') },
  { path: '/games/creator', component: () => import('@/pages/game/GameCreators.vue') },
  { path: '/games/news/more', component: () => import('@/pages/game/MoreGameNews.vue') },
  { path: '/games/prize', component: () => import('@/pages/game/Lottery.vue') },
  { path: '/games/prize/activation', component: () => import('@/pages/game/LotteryActivation.vue') },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
