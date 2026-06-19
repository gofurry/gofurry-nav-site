import { createRouter, createWebHashHistory } from 'vue-router'
import ConsoleView from './views/ConsoleView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'console',
      component: ConsoleView,
    },
  ],
})
