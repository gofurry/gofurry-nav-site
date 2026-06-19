<template>
  <div class="relative flex w-full flex-1 flex-col overflow-hidden bg-[#f5ead8] text-slate-900">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute -left-20 top-10 h-72 w-72 rounded-full bg-orange-300/20 blur-3xl"></div>
      <div class="absolute right-0 top-0 h-96 w-96 rounded-full bg-amber-200/35 blur-3xl"></div>
      <div class="absolute bottom-10 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-sky-200/15 blur-3xl"></div>
    </div>

    <main class="relative mx-auto flex w-full max-w-7xl flex-1 flex-col gap-10 px-4 py-8 sm:px-6 md:gap-14 md:px-8 md:py-12">
      <section class="relative overflow-hidden rounded-xl border border-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.94),rgba(51,65,85,0.92)_52%,rgba(194,120,3,0.88))] px-6 py-8 text-white shadow-[0_30px_80px_rgba(15,23,42,0.22)] md:px-10 md:py-12">
        <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.22),transparent_32%),radial-gradient(circle_at_bottom_left,rgba(251,191,36,0.18),transparent_28%)]"></div>
        <div class="relative grid gap-8 lg:grid-cols-[minmax(0,1.5fr)_minmax(320px,0.9fr)] lg:items-center">
          <div class="space-y-6">
            <div class="space-y-4">
              <h1 class="max-w-4xl text-4xl font-semibold leading-tight md:text-5xl md:leading-[1.05]">
                {{ hero.title }}
              </h1>
              <p class="max-w-3xl text-sm leading-7 text-slate-200 md:text-base">
                {{ hero.lead }}
              </p>
            </div>

            <div class="flex flex-wrap gap-3">
              <span
                  v-for="tag in hero.tags"
                  :key="tag"
                  class="rounded-lg border border-white/12 bg-white/8 px-4 py-2 text-sm text-slate-100"
              >
                {{ tag }}
              </span>
            </div>

            <div class="flex flex-wrap gap-3">
              <RouterLink
                  to="/nav"
                  class="inline-flex items-center justify-center rounded-xl bg-orange-300 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-orange-200"
              >
                {{ hero.primaryAction }}
              </RouterLink>
              <a
                  href="https://github.com/gofurry/gofurry-nav-site"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex items-center justify-center rounded-xl border border-white/18 bg-white/8 px-5 py-3 text-sm font-semibold text-white transition hover:bg-white/14"
              >
                {{ hero.secondaryAction }}
              </a>
            </div>
          </div>

          <div class="grid gap-4 self-center sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
            <article
                v-for="stat in stats"
                :key="stat.label"
                class="flex h-full flex-col justify-center rounded-lg border border-white/12 bg-white/10 p-5 backdrop-blur-sm"
            >
              <div class="text-xs uppercase tracking-[0.24em] text-orange-100/80">
                {{ stat.label }}
              </div>
              <p class="mt-2 text-sm leading-6 text-slate-200">
                {{ stat.desc }}
              </p>
            </article>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article
            v-for="feature in features"
            :key="feature.title"
            class="group relative flex h-full flex-col overflow-hidden rounded-lg border border-orange-200/60 bg-white/70 p-6 shadow-[0_18px_40px_rgba(148,163,184,0.12)] backdrop-blur-sm transition duration-300 hover:-translate-y-1 hover:shadow-[0_24px_50px_rgba(194,120,3,0.18)]"
        >
          <div class="absolute inset-x-0 top-0 h-1 bg-[linear-gradient(90deg,#fb923c,#fbbf24,#38bdf8)] opacity-80"></div>
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-900/90 shadow-lg shadow-slate-900/10">
            <img :src="feature.icon" :alt="feature.title" class="h-6 w-6 opacity-95" />
          </div>
          <h2 class="mt-5 text-xl font-semibold text-slate-900">
            {{ feature.title }}
          </h2>
          <p class="mt-3 text-sm leading-7 text-slate-600">
            {{ feature.desc }}
          </p>
          <div class="mt-auto pt-5 text-xs uppercase tracking-[0.24em] text-orange-600">
            {{ feature.meta }}
          </div>
        </article>
      </section>

      <section class="grid gap-6 lg:grid-cols-2">
        <article class="rounded-xl border border-slate-900/8 bg-slate-900 p-6 text-slate-100 shadow-[0_22px_50px_rgba(15,23,42,0.2)] md:p-8">
          <div class="text-xs uppercase tracking-[0.28em] text-orange-300/90">
            {{ teamSection.kicker }}
          </div>
          <div class="mt-5 flex items-center gap-4">
            <img
                :src="teamSection.member.avatar"
                :alt="teamSection.member.name"
                class="h-20 w-20 rounded-2xl object-cover ring-2 ring-white/10"
            />
            <div>
              <div class="text-2xl font-semibold text-white">
                {{ teamSection.member.name }}
              </div>
              <div class="mt-1 text-sm text-slate-300">
                {{ teamSection.member.role }}
              </div>
            </div>
          </div>
          <p class="mt-5 text-sm leading-7 text-slate-300">
            {{ teamSection.desc }}
          </p>
          <a
              :href="teamSection.member.link"
              target="_blank"
              rel="noopener noreferrer"
              class="mt-6 inline-flex items-center rounded-lg border border-white/14 bg-white/6 px-4 py-2 text-sm font-medium text-white transition hover:bg-white/12"
          >
            {{ teamSection.member.action }}
          </a>
        </article>

        <article class="rounded-xl border border-orange-200/70 bg-white/70 p-6 shadow-[0_18px_48px_rgba(148,163,184,0.14)] backdrop-blur-sm md:p-8">
          <div class="text-xs uppercase tracking-[0.28em] text-orange-500">
            {{ contact.kicker }}
          </div>
          <h2 class="mt-3 text-2xl font-semibold text-slate-900">
            {{ contact.title }}
          </h2>
          <p class="mt-4 text-sm leading-7 text-slate-600">
            {{ contact.desc }}
          </p>

          <div class="mt-6 grid gap-3">
            <a
                v-for="item in contact.links"
                :key="item.label"
                :href="item.href"
                :target="item.external ? '_blank' : undefined"
                :rel="item.external ? 'noopener noreferrer' : undefined"
                class="flex items-center justify-between rounded-lg bg-slate-950/[0.04] px-4 py-3 text-sm text-slate-700 transition hover:bg-orange-100/80"
            >
              <span>{{ item.label }}</span>
              <span class="font-medium text-orange-600">{{ item.value }}</span>
            </a>
          </div>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { i18n } from '@/main'
import compassIcon from '@/assets/svgs/compass.svg'
import gamepadIcon from '@/assets/svgs/gamepad.svg'
import apiIcon from '@/assets/svgs/api.svg'
import featherIcon from '@/assets/svgs/feather.svg'

const isZh = computed(() => i18n.global.locale.value === 'zh')

const hero = computed(() => (
  isZh.value
    ? {
        title: 'gofurry 正在把导航、游戏情报与开放生态整理成一个更清晰的入口。',
        lead: '我们希望把分散的兽圈站点、项目、游戏和长期内容串成一条更容易理解的路径。它不是复杂的门户，而是一个轻量、稳定、持续演进的公共入口。',
        tags: ['导航与发现', '游戏情报', '开放平台', 'DeepFurry'],
        primaryAction: '进入导航站',
        secondaryAction: '查看 GitHub',
      }
    : {
        title: 'gofurry brings navigation, game intelligence, and open ecosystem entry points into one clearer experience.',
        lead: 'We want scattered furry sites, projects, games, and long-form work to feel connected instead of fragmented. gofurry is designed as a lightweight, stable, and sustainable public-facing gateway.',
        tags: ['Navigation', 'Game Intel', 'Open Platform', 'DeepFurry'],
        primaryAction: 'Open Navigation',
        secondaryAction: 'View GitHub',
      }
))

const stats = computed(() => (
  isZh.value
    ? [
        { label: 'Core Surface', desc: '导航、游戏、开放平台与 DeepFurry 共同构成当前的核心展示面。' },
        { label: 'Design Goal', desc: '一个更轻量、更顺手、也更持续更新的兽圈入口。' },
        { label: 'Operating Style', desc: '以公共性、开放性和长期可维护性作为主要方向。' },
        { label: 'Focus', desc: '比起堆叠功能，我们更重视结构清晰和信息密度。' },
      ]
    : [
        { label: 'Core Surface', desc: 'Navigation, games, open platform, and DeepFurry form the current core surface.' },
        { label: 'Design Goal', desc: 'A lighter, clearer, and more maintainable furry gateway.' },
        { label: 'Operating Style', desc: 'Public value, openness, and long-term sustainability shape the direction.' },
        { label: 'Focus', desc: 'We value clarity and information density more than feature sprawl.' },
      ]
))

const features = computed(() => (
  isZh.value
    ? [
        {
          icon: compassIcon,
          title: '导航',
          desc: '把分散的网站入口、分组和基础状态整合成清晰的浏览路径，帮助用户更快找到目标内容。',
          meta: 'Discovery Layer',
        },
        {
          icon: gamepadIcon,
          title: '游戏',
          desc: '围绕 Steam 兽游构建更新情报、检索、详情和作者视角，让内容浏览和深入追踪都更顺手。',
          meta: 'Game Intel',
        },
        {
          icon: apiIcon,
          title: '开放平台',
          desc: '承接后续开放能力、公共服务和外部扩展页面，作为 gofurry 的延展入口。',
          meta: 'Open Surface',
        },
        {
          icon: featherIcon,
          title: 'DeepFurry',
          desc: '面向更长篇、更研究化、更具表达深度的内容，承担宣传页与长期叙事延伸。',
          meta: 'Long-form Research',
        },
      ]
    : [
        {
          icon: compassIcon,
          title: 'Navigation',
          desc: 'A clearer path across scattered furry sites, groups, and status signals, so discovery feels intentional instead of noisy.',
          meta: 'Discovery Layer',
        },
        {
          icon: gamepadIcon,
          title: 'Games',
          desc: 'Steam furry game updates, search, details, and creator views are organized for both fast browsing and deeper tracking.',
          meta: 'Game Intel',
        },
        {
          icon: apiIcon,
          title: 'Open Platform',
          desc: 'A future-facing surface for public capabilities, open integrations, and external ecosystem pages built around gofurry.',
          meta: 'Open Surface',
        },
        {
          icon: featherIcon,
          title: 'DeepFurry',
          desc: 'A space for longer-form thinking, research-oriented writing, and a more expressive public-facing narrative.',
          meta: 'Long-form Research',
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
        title: '如果你想交流项目、内容或合作方向，可以从这里找到我们。',
        desc: '欢迎反馈站点问题、提出想法，或者交流关于开放平台与 DeepFurry 的后续方向。',
        links: [
          { label: '邮箱', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry', href: 'https://github.com/gofurry', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
    : {
        kicker: 'Contact',
        title: 'If you want to talk about the project, content, or possible collaboration, start here.',
        desc: 'Feedback, ideas, and conversations around the open platform or DeepFurry direction are always welcome.',
        links: [
          { label: 'Email', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry', href: 'https://github.com/gofurry', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
))
</script>
