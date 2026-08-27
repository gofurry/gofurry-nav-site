<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import {
  Activity,
  Bell,
  Boxes,
  ChevronLeft,
  ChevronRight,
  Database,
  Gamepad2,
  Globe2,
  Layers3,
  LogOut,
  MessageSquareText,
  Quote,
  Tags,
} from 'lucide-vue-next'
import { resources } from '../resources'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const collapsed = ref(false)

const navResources = computed(() => resources.filter((item) => item.section === 'nav'))
const gameResources = computed(() => resources.filter((item) => item.section === 'game'))

const resourceIcons: Record<string, Component> = {
  sayings: Quote,
  'update-notices': Bell,
  'collector-domains': Globe2,
  sites: Database,
  'site-groups': Layers3,
  'site-group-maps': Boxes,
  'featured-sites': Globe2,
  games: Gamepad2,
  comments: MessageSquareText,
  prizes: Boxes,
  tags: Tags,
  'tag-maps': Layers3,
}

async function logout() {
  await auth.logout()
  router.push('/login')
}

function isActive(path: string) {
  return route.path === path
}

function iconFor(key: string) {
  return resourceIcons[key] ?? Database
}
</script>

<template>
  <div class="admin-shell" :class="{ 'admin-shell--collapsed': collapsed }">
    <aside class="admin-sidebar">
      <div class="admin-sidebar__brand">
        <div class="admin-sidebar__identity">
          <div class="admin-sidebar__brand-copy">
            <div class="admin-sidebar__eyebrow">gofurry</div>
            <div class="admin-sidebar__title">Admin</div>
          </div>
        </div>
        <button
          class="icon-button admin-sidebar__toggle"
          type="button"
          :title="collapsed ? '展开菜单' : '收起菜单'"
          :aria-label="collapsed ? '展开菜单' : '收起菜单'"
          @click="collapsed = !collapsed"
        >
          <ChevronRight v-if="collapsed" :size="18" />
          <ChevronLeft v-else :size="18" />
        </button>
      </div>

      <nav class="admin-nav" aria-label="管理菜单">
        <section class="admin-nav__section">
          <div class="admin-nav__section-title">采集控制</div>
          <div class="admin-nav__items">
            <RouterLink
              to="/collection"
              class="admin-nav__link"
              :class="{ 'admin-nav__link--active': isActive('/collection') }"
              title="Collection Center"
            >
              <Activity :size="18" />
              <span>Collection Center</span>
            </RouterLink>
          </div>
        </section>

        <section class="admin-nav__section">
          <div class="admin-nav__section-title">导航库</div>
          <div class="admin-nav__items">
            <RouterLink
              v-for="item in navResources"
              :key="item.key"
              :to="`/${item.section}/${item.key}`"
              class="admin-nav__link"
              :class="{ 'admin-nav__link--active': isActive(`/${item.section}/${item.key}`) }"
              :title="item.title"
            >
              <component :is="iconFor(item.key)" :size="18" />
              <span>{{ item.title }}</span>
            </RouterLink>
          </div>
        </section>

        <section class="admin-nav__section">
          <div class="admin-nav__section-title">游戏库</div>
          <div class="admin-nav__items">
            <RouterLink
              v-for="item in gameResources"
              :key="item.key"
              :to="`/${item.section}/${item.key}`"
              class="admin-nav__link"
              :class="{ 'admin-nav__link--active': isActive(`/${item.section}/${item.key}`) }"
              :title="item.title"
            >
              <component :is="iconFor(item.key)" :size="18" />
              <span>{{ item.title }}</span>
            </RouterLink>
          </div>
        </section>
      </nav>

      <button class="admin-nav__link admin-sidebar__logout" type="button" title="登出" @click="logout">
        <LogOut :size="18" />
        <span>登出</span>
      </button>
    </aside>

    <main class="admin-workspace">
      <RouterView />
    </main>
  </div>
</template>
