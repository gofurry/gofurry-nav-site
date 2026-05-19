<template>
  <div :class="pageClass" :style="pageVars">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute -left-24 top-8 h-72 w-72 rounded-full bg-[var(--about-orb-a)] blur-3xl"></div>
      <div class="absolute right-[-8rem] top-24 h-96 w-96 rounded-full bg-[var(--about-orb-b)] blur-3xl"></div>
      <div class="absolute bottom-16 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-[var(--about-orb-c)] blur-3xl"></div>
    </div>

    <main class="relative mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-8 sm:px-6 md:gap-12 md:px-8 md:py-12">
      <section class="about-hero">
        <div class="relative grid gap-8 lg:grid-cols-[minmax(0,1.35fr)_minmax(300px,0.85fr)] lg:items-end">
          <div class="space-y-6">
            <div class="space-y-4">
              <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--about-accent)]">
                {{ hero.kicker }}
              </p>
              <h1 class="max-w-4xl text-4xl font-semibold leading-tight md:text-6xl md:leading-[1.02]">
                {{ hero.title }}
              </h1>
              <p class="max-w-3xl text-sm leading-7 text-[var(--about-muted)] md:text-base">
                {{ hero.lead }}
              </p>
            </div>

            <div class="flex flex-wrap gap-3">
              <span
                  v-for="tag in hero.tags"
                  :key="tag"
                  class="rounded-lg border border-[var(--about-chip-border)] bg-[var(--about-chip-bg)] px-4 py-2 text-sm text-[var(--about-chip-text)]"
              >
                {{ tag }}
              </span>
            </div>

            <div class="flex flex-wrap gap-3">
              <NuxtLink
                  to="/nav"
                  class="inline-flex items-center justify-center rounded-xl bg-[var(--about-action-bg)] px-5 py-3 text-sm font-semibold text-[var(--about-action-text)] transition hover:brightness-105"
              >
                {{ hero.primaryAction }}
              </NuxtLink>
              <a
                  href="https://github.com/gofurry/gofurry-nav-site"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex items-center justify-center rounded-xl border border-[var(--about-button-border)] bg-[var(--about-button-bg)] px-5 py-3 text-sm font-semibold text-[var(--about-button-text)] transition hover:bg-[var(--about-button-hover)]"
              >
                {{ hero.secondaryAction }}
              </a>
            </div>
          </div>

          <div class="grid gap-3">
            <article
                v-for="item in overview"
                :key="item.label"
                class="about-compact"
            >
              <div class="text-xs uppercase tracking-[0.24em] text-[var(--about-accent-soft)]">
                {{ item.label }}
              </div>
              <p class="mt-2 text-sm leading-6 text-[var(--about-muted)]">
                {{ item.desc }}
              </p>
            </article>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article
            v-for="feature in features"
            :key="feature.title"
            class="about-card group"
        >
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-[var(--about-icon-bg)] shadow-sm">
            <img :src="feature.icon" :alt="feature.title" class="h-6 w-6 opacity-95" />
          </div>
          <h2 class="mt-5 text-xl font-semibold text-[var(--about-heading)]">
            {{ feature.title }}
          </h2>
          <p class="mt-3 text-sm leading-7 text-[var(--about-muted)]">
            {{ feature.desc }}
          </p>
          <div class="mt-auto pt-5 text-xs uppercase tracking-[0.24em] text-[var(--about-accent)]">
            {{ feature.meta }}
          </div>
        </article>
      </section>

      <section class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <article class="about-panel">
          <div class="text-xs uppercase tracking-[0.28em] text-[var(--about-accent)]">
            {{ teamSection.kicker }}
          </div>
          <div class="mt-5 flex items-center gap-4">
            <img
                :src="teamSection.member.avatar"
                :alt="teamSection.member.name"
                class="h-20 w-20 rounded-2xl object-cover ring-2 ring-[var(--about-avatar-ring)]"
            />
            <div>
              <div class="text-2xl font-semibold text-[var(--about-heading)]">
                {{ teamSection.member.name }}
              </div>
              <div class="mt-1 text-sm text-[var(--about-muted)]">
                {{ teamSection.member.role }}
              </div>
            </div>
          </div>
          <p class="mt-5 text-sm leading-7 text-[var(--about-muted)]">
            {{ teamSection.desc }}
          </p>
          <a
              :href="teamSection.member.link"
              target="_blank"
              rel="noopener noreferrer"
              class="mt-6 inline-flex items-center rounded-lg border border-[var(--about-button-border)] bg-[var(--about-button-bg)] px-4 py-2 text-sm font-medium text-[var(--about-button-text)] transition hover:bg-[var(--about-button-hover)]"
          >
            {{ teamSection.member.action }}
          </a>
        </article>

        <article class="about-panel">
          <div class="text-xs uppercase tracking-[0.28em] text-[var(--about-accent)]">
            {{ contact.kicker }}
          </div>
          <h2 class="mt-3 text-2xl font-semibold text-[var(--about-heading)]">
            {{ contact.title }}
          </h2>
          <p class="mt-4 text-sm leading-7 text-[var(--about-muted)]">
            {{ contact.desc }}
          </p>

          <div class="mt-6 grid gap-3">
            <a
                v-for="item in contact.links"
                :key="item.label"
                :href="item.href"
                :target="item.external ? '_blank' : undefined"
                :rel="item.external ? 'noopener noreferrer' : undefined"
                class="about-link"
            >
              <span>{{ item.label }}</span>
              <span class="font-medium text-[var(--about-accent)]">{{ item.value }}</span>
            </a>
          </div>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { i18n } from '@/main'
import { useThemeStore } from '@/stores/theme'
import compassIcon from '@/assets/svgs/compass.svg'
import gamepadIcon from '@/assets/svgs/gamepad.svg'
import apiIcon from '@/assets/svgs/api.svg'
import featherIcon from '@/assets/svgs/feather.svg'

const themeStore = useThemeStore()
const isZh = computed(() => i18n.global.locale.value === 'zh')
const isDark = computed(() => themeStore.theme === 'dark')

onMounted(() => {
  themeStore.initTheme()
})

const pageClass = computed(() => [
  'relative flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500',
  isDark.value ? 'bg-[#070a10] text-slate-100' : 'bg-[#f7efe2] text-slate-950'
])

const pageVars = computed(() => isDark.value
  ? {
      '--about-surface': 'rgba(12, 17, 27, 0.88)',
      '--about-surface-strong': 'rgba(6, 9, 15, 0.92)',
      '--about-border': 'rgba(148, 163, 184, 0.16)',
      '--about-rule': 'rgba(148, 163, 184, 0.18)',
      '--about-heading': 'rgb(248 250 252)',
      '--about-muted': 'rgb(148 163 184)',
      '--about-accent': 'rgb(251 191 36)',
      '--about-accent-soft': 'rgb(253 230 138)',
      '--about-chip-bg': 'rgba(255, 255, 255, 0.06)',
      '--about-chip-border': 'rgba(255, 255, 255, 0.1)',
      '--about-chip-text': 'rgb(226 232 240)',
      '--about-icon-bg': 'rgba(255, 255, 255, 0.08)',
      '--about-avatar-ring': 'rgba(255, 255, 255, 0.12)',
      '--about-action-bg': 'rgb(251 191 36)',
      '--about-action-text': 'rgb(15 23 42)',
      '--about-button-bg': 'rgba(255, 255, 255, 0.06)',
      '--about-button-border': 'rgba(255, 255, 255, 0.12)',
      '--about-button-text': 'rgb(248 250 252)',
      '--about-button-hover': 'rgba(255, 255, 255, 0.1)',
      '--about-orb-a': 'rgba(251, 191, 36, 0.12)',
      '--about-orb-b': 'rgba(56, 189, 248, 0.1)',
      '--about-orb-c': 'rgba(99, 102, 241, 0.08)'
    }
  : {
      '--about-surface': 'rgba(255, 255, 255, 0.7)',
      '--about-surface-strong': 'rgba(15, 23, 42, 0.94)',
      '--about-border': 'rgba(194, 120, 3, 0.18)',
      '--about-rule': 'rgba(194, 120, 3, 0.2)',
      '--about-heading': 'rgb(15 23 42)',
      '--about-muted': 'rgb(71 85 105)',
      '--about-accent': 'rgb(217 119 6)',
      '--about-accent-soft': 'rgb(180 83 9)',
      '--about-chip-bg': 'rgba(255, 255, 255, 0.66)',
      '--about-chip-border': 'rgba(251, 146, 60, 0.24)',
      '--about-chip-text': 'rgb(51 65 85)',
      '--about-icon-bg': 'rgba(15, 23, 42, 0.92)',
      '--about-avatar-ring': 'rgba(15, 23, 42, 0.12)',
      '--about-action-bg': 'rgb(251 191 36)',
      '--about-action-text': 'rgb(15 23 42)',
      '--about-button-bg': 'rgba(15, 23, 42, 0.04)',
      '--about-button-border': 'rgba(15, 23, 42, 0.12)',
      '--about-button-text': 'rgb(15 23 42)',
      '--about-button-hover': 'rgba(15, 23, 42, 0.07)',
      '--about-orb-a': 'rgba(251, 146, 60, 0.18)',
      '--about-orb-b': 'rgba(251, 191, 36, 0.22)',
      '--about-orb-c': 'rgba(56, 189, 248, 0.12)'
    }
)

const hero = computed(() => (
  isZh.value
    ? {
        kicker: 'About GoFurry',
        title: '一个面向兽圈的导航、游戏情报与长期内容入口。',
        lead: 'GoFurry 现在更像一个清晰的发现层：把站点导航、热门兽游、兽圈档案和深度兽研放在同一条路径里，让分散的信息更容易被找到、理解和持续维护。',
        tags: ['站点导航', '热门兽游', '兽圈档案', '深度兽研'],
        primaryAction: '进入站点导航',
        secondaryAction: '查看开源仓库',
      }
    : {
        kicker: 'About GoFurry',
        title: 'A discovery gateway for furry sites, games, archives, and long-form research.',
        lead: 'GoFurry is becoming a clearer discovery layer: navigation, game intelligence, community archives, and DeepFurry research are organized into one maintainable public path.',
        tags: ['Navigation', 'Furry Games', 'Archive', 'DeepFurry'],
        primaryAction: 'Open Navigation',
        secondaryAction: 'View Repository',
      }
))

const overview = computed(() => (
  isZh.value
    ? [
        { label: 'Current Focus', desc: '先把主要页面、搜索体验和内容结构打磨稳定，再逐步扩展开放能力。' },
        { label: 'Design Direction', desc: '轻量、克制、可持续，不做拥挤的门户，也不把探索变成信息噪音。' },
        { label: 'Open Source', desc: '站点前端持续在 GitHub 上公开迭代，方便反馈、审阅和协作。' },
      ]
    : [
        { label: 'Current Focus', desc: 'Stabilize the core pages, search experience, and content structure before expanding open capabilities.' },
        { label: 'Design Direction', desc: 'Lightweight, restrained, and sustainable, without turning discovery into portal noise.' },
        { label: 'Open Source', desc: 'The frontend keeps evolving in public on GitHub for feedback, review, and collaboration.' },
      ]
))

const features = computed(() => (
  isZh.value
    ? [
        {
          icon: compassIcon,
          title: '站点导航',
          desc: '整理兽圈相关网站、工具、社区与内容入口，让图标、分类、搜索和访问路径保持清晰。',
          meta: 'Navigation',
        },
        {
          icon: gamepadIcon,
          title: '热门兽游',
          desc: '围绕 Steam 兽游建立索引、更新和详情页，帮助玩家更快发现作品，也帮助作者被看见。',
          meta: 'Game Intel',
        },
        {
          icon: apiIcon,
          title: '兽圈档案',
          desc: '沉淀长期有价值的资料、站点变化和项目记录，让信息不只停留在一次性的浏览里。',
          meta: 'Archive',
        },
        {
          icon: featherIcon,
          title: '深度兽研',
          desc: '承接更长篇、更研究化的表达，给文化、作品和社区议题留出更安静的阅读空间。',
          meta: 'Research',
        },
      ]
    : [
        {
          icon: compassIcon,
          title: 'Navigation',
          desc: 'Curated furry websites, tools, communities, and content entry points with clearer icons, categories, search, and routes.',
          meta: 'Navigation',
        },
        {
          icon: gamepadIcon,
          title: 'Furry Games',
          desc: 'Steam furry game indexes, updates, and detail pages that help players discover work and help creators become visible.',
          meta: 'Game Intel',
        },
        {
          icon: apiIcon,
          title: 'Community Archive',
          desc: 'Long-lived records for useful resources, site changes, and project traces so information does not vanish after one visit.',
          meta: 'Archive',
        },
        {
          icon: featherIcon,
          title: 'DeepFurry',
          desc: 'A quieter reading space for long-form, research-oriented writing around culture, works, and community topics.',
          meta: 'Research',
        },
      ]
))

const teamSection = computed(() => (
  isZh.value
    ? {
        kicker: 'People Behind It',
        desc: '目前项目仍然以较小规模持续推进，但会保持稳定打磨与公开迭代。站点、内容和体验都会围绕真实使用场景慢慢变得更完整。',
        member: {
          name: '福狼',
          role: '开发者 / 维护者',
          avatar: 'https://qcdn.go-furry.com/game/creator/100/avatar.jpg',
          link: 'https://github.com/gofurry',
          action: '前往个人主页',
        },
      }
    : {
        kicker: 'People Behind It',
        desc: 'The project is still being pushed forward by a small footprint, but with consistent iteration and public-facing refinement. The site, content, and UX will keep improving around real usage patterns.',
        member: {
          name: 'Fu Lang',
          role: 'Developer / Maintainer',
          avatar: 'https://qcdn.go-furry.com/game/creator/100/avatar.jpg',
          link: 'https://github.com/gofurry',
          action: 'Open project page',
        },
      }
))

const contact = computed(() => (
  isZh.value
    ? {
        kicker: 'Contact',
        title: '想反馈站点、补充资料或交流合作，可以从这里找到我们。',
        desc: '欢迎提交问题、建议站点、补充游戏信息，或者讨论开放平台与 DeepFurry 后续方向。',
        links: [
          { label: '邮箱', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry/gofurry-nav-site', href: 'https://github.com/gofurry/gofurry-nav-site', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
    : {
        kicker: 'Contact',
        title: 'For feedback, resource updates, or collaboration, you can reach us here.',
        desc: 'Issues, site suggestions, game information, and conversations around the open platform or DeepFurry direction are welcome.',
        links: [
          { label: 'Email', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry/gofurry-nav-site', href: 'https://github.com/gofurry/gofurry-nav-site', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
))
</script>

<style scoped>
.about-hero,
.about-panel,
.about-card,
.about-compact {
  border: 1px solid var(--about-border);
  background: var(--about-surface);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18px);
  transition: background-color 0.5s ease, border-color 0.5s ease, box-shadow 0.5s ease, transform 0.3s ease;
}

.about-hero {
  border-radius: 18px;
  padding: clamp(1.5rem, 3vw, 3rem);
}

.about-card,
.about-panel,
.about-compact {
  border-radius: 14px;
}

.about-card {
  display: flex;
  min-height: 16rem;
  flex-direction: column;
  padding: 1.5rem;
}

.about-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.16);
}

.about-panel {
  padding: clamp(1.5rem, 2vw, 2rem);
}

.about-compact {
  padding: 1rem;
}

.about-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-radius: 0.75rem;
  background: var(--about-button-bg);
  padding: 0.875rem 1rem;
  color: var(--about-button-text);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.about-link:hover {
  background: var(--about-button-hover);
}
</style>
