<template>
  <div :class="pageClass" :style="pageVars">
    <GoFurryGridBackground profile="light" />
    <div class="pointer-events-none absolute inset-x-0 top-0 h-48 bg-[var(--about-top-veil)]" />

    <main class="relative mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-8 sm:px-6 md:px-8 md:py-12">
      <h1 class="sr-only">{{ pageSeo.heading }}</h1>
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
          <div class="mt-6 flex flex-wrap gap-3">
            <a
                :href="teamSection.member.link"
                target="_blank"
                rel="noopener noreferrer"
                class="about-panel-action"
            >
              {{ teamSection.member.action }}
            </a>
            <a
                :href="teamSection.issueLink"
                target="_blank"
                rel="noopener noreferrer"
                class="about-panel-action about-panel-action--accent"
            >
              {{ teamSection.issueAction }}
            </a>
          </div>
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
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
const themeStore = useThemeStore()
const { locale } = useI18n()
const isZh = computed(() => locale.value === 'zh')
const isDark = computed(() => themeStore.theme === 'dark')
const pageSeo = computed(() => (
  isZh.value
    ? {
        heading: '关于 GoFurry 兽人控导航站',
        title: '关于 GoFurry 兽人控导航站 - 维护者、反馈与收录合作',
        description: '了解 GoFurry 兽人控导航站的维护者、项目定位、站点收录原则、反馈方式与合作入口。GoFurry 专注兽人文化资源导航和用户视角的站点观测数据。'
      }
    : {
        heading: 'About GoFurry Navigation',
        title: 'About GoFurry Navigation - Maintainer, feedback, and listing collaboration',
        description: 'Learn about GoFurry Navigation, its maintainer, project direction, listing feedback channels, and collaboration paths for furry culture resources and user-perspective site observation data.'
      }
))

useSeoMeta({
  title: () => pageSeo.value.title,
  description: () => pageSeo.value.description,
  ogTitle: () => pageSeo.value.title,
  ogDescription: () => pageSeo.value.description,
})

onMounted(() => {
  themeStore.initTheme()
})

const pageClass = computed(() => [
  'relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500',
  isDark.value ? 'bg-[#08101b] text-slate-100' : 'bg-[#f6ebdc] text-slate-950'
])

const pageVars = computed(() => isDark.value
  ? {
      '--about-surface': 'rgba(9, 16, 27, 0.74)',
      '--about-surface-strong': 'rgba(8, 13, 22, 0.84)',
      '--about-border': 'rgba(123, 154, 189, 0.2)',
      '--about-rule': 'rgba(123, 154, 189, 0.18)',
      '--about-heading': 'rgb(239 246 255)',
      '--about-muted': 'rgb(179 195 214)',
      '--about-accent': 'rgb(142 214 255)',
      '--about-accent-soft': 'rgb(176 227 255)',
      '--about-chip-bg': 'rgba(255, 255, 255, 0.05)',
      '--about-chip-border': 'rgba(148, 163, 184, 0.14)',
      '--about-chip-text': 'rgb(226 232 240)',
      '--about-avatar-ring': 'rgba(148, 163, 184, 0.18)',
      '--about-action-bg': 'rgba(251, 146, 60, 0.22)',
      '--about-action-hover': 'rgba(251, 146, 60, 0.3)',
      '--about-action-text': 'rgb(255 237 213)',
      '--about-action-border': 'rgba(251, 146, 60, 0.34)',
      '--about-action-border-hover': 'rgba(251, 146, 60, 0.48)',
      '--about-button-bg': 'rgba(255, 255, 255, 0.04)',
      '--about-button-border': 'rgba(148, 163, 184, 0.14)',
      '--about-button-text': 'rgb(239 246 255)',
      '--about-button-hover': 'rgba(255, 255, 255, 0.09)',
      '--about-link-bg': 'rgba(255, 255, 255, 0.06)',
      '--about-link-border': 'rgba(148, 163, 184, 0.12)',
      '--about-link-hover': 'rgba(255, 255, 255, 0.1)',
      '--about-top-veil': 'linear-gradient(180deg, rgba(7, 13, 23, 0.46), rgba(7, 13, 23, 0))'
    }
  : {
      '--about-surface': 'rgba(255, 249, 241, 0.72)',
      '--about-surface-strong': 'rgba(255, 252, 247, 0.88)',
      '--about-border': 'rgba(168, 112, 46, 0.18)',
      '--about-rule': 'rgba(168, 112, 46, 0.18)',
      '--about-heading': 'rgb(15 23 42)',
      '--about-muted': 'rgb(71 85 105)',
      '--about-accent': 'rgb(190 112 28)',
      '--about-accent-soft': 'rgb(180 83 9)',
      '--about-chip-bg': 'rgba(255, 255, 255, 0.58)',
      '--about-chip-border': 'rgba(168, 112, 46, 0.18)',
      '--about-chip-text': 'rgb(51 65 85)',
      '--about-avatar-ring': 'rgba(15, 23, 42, 0.12)',
      '--about-action-bg': 'rgba(251, 146, 60, 0.18)',
      '--about-action-hover': 'rgba(251, 146, 60, 0.26)',
      '--about-action-text': 'rgb(15 23 42)',
      '--about-action-border': 'rgba(251, 146, 60, 0.28)',
      '--about-action-border-hover': 'rgba(251, 146, 60, 0.42)',
      '--about-button-bg': 'rgba(255, 255, 255, 0.4)',
      '--about-button-border': 'rgba(15, 23, 42, 0.09)',
      '--about-button-text': 'rgb(15 23 42)',
      '--about-button-hover': 'rgba(255, 255, 255, 0.62)',
      '--about-link-bg': 'rgba(251, 146, 60, 0.12)',
      '--about-link-border': 'rgba(180, 83, 9, 0.13)',
      '--about-link-hover': 'rgba(251, 146, 60, 0.18)',
      '--about-top-veil': 'linear-gradient(180deg, rgba(255, 248, 239, 0.54), rgba(255, 248, 239, 0))'
    }
)

const teamSection = computed(() => (
  isZh.value
    ? {
        kicker: 'People Behind It',
        desc: '兽人控导航站站长，Golang 软件工程师。兽人控导航站是一个开放的兽圈站点观测平台，主要提供站点的导航与用户视角的网络观测数据，无用户系统、不中转、不收集任何用户的敏感信息。目前项目处于小规模的持续迭代中，如有任何疑问欢迎前往Github提交Issue，十分感谢您的贡献。',
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: '提交您的建议',
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
        desc: 'Site owner of GoFurry Navigation and Golang software engineer. GoFurry Navigation is an open furry-site observability platform focused on navigation and user-perspective web observation data, with no user system, no traffic proxying, and no collection of sensitive user information.',
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: 'Submit Suggestion',
        member: {
          name: 'Faolan',
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
        title: '站点反馈、收录信息补充、交流合作',
        desc: '欢迎提交问题、建议站点、补充收录信息，或者讨论兽人控导航站的后续方向。',
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
.about-panel {
  border: 1px solid var(--about-border);
  background: var(--about-surface);
  box-shadow: 0 20px 48px rgba(15, 23, 42, 0.1);
  backdrop-filter: blur(14px);
  transition: background-color 0.5s ease, border-color 0.5s ease, box-shadow 0.5s ease, transform 0.3s ease;
  border-radius: 20px;
}

.about-panel {
  padding: clamp(1.5rem, 2vw, 2rem);
}

.about-panel-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  border: 1px solid var(--about-button-border);
  background: var(--about-button-bg);
  padding: 0.6rem 0.95rem;
  color: var(--about-button-text);
  font-size: 0.875rem;
  font-weight: 600;
  transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
}

.about-panel-action--accent {
  border-color: var(--about-action-border);
  background: var(--about-action-bg);
  color: var(--about-action-text);
}

.about-panel-action:hover {
  background: var(--about-button-hover);
  transform: translateY(-1px);
}

.about-panel-action--accent:hover {
  border-color: var(--about-action-border-hover);
  background: var(--about-action-hover);
}

.about-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-radius: 0.75rem;
  border: 1px solid var(--about-link-border);
  background: var(--about-link-bg);
  padding: 0.875rem 1rem;
  color: var(--about-button-text);
  transition: border-color 0.2s ease, background-color 0.2s ease, color 0.2s ease;
}

.about-link:hover {
  background: var(--about-link-hover);
}

@media (max-width: 767px) {
  .about-panel {
    border-radius: 18px;
  }
}
</style>
