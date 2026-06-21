<template>
  <div class="gf-static-page about-page relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="gf-static-page__top-veil" />

    <main class="gf-static-page__main gf-static-page__main--about">
      <h1 class="sr-only">{{ pageSeo.heading }}</h1>
      <section class="gf-static-page__grid gf-static-page__grid--about">
        <article class="about-panel">
          <div class="gf-static-kicker">
            {{ teamSection.kicker }}
          </div>
          <div class="about-person">
            <img
                :src="teamSection.member.avatar"
                :alt="teamSection.member.name"
                class="about-avatar"
            />
            <div>
              <div class="about-name">
                {{ teamSection.member.name }}
              </div>
              <div class="about-role">
                {{ teamSection.member.role }}
              </div>
            </div>
          </div>
          <p class="about-description">
            {{ teamSection.desc }}
          </p>
          <div class="about-actions">
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
            <NuxtLink
                :to="localePath('/about/faolan')"
                class="about-panel-action about-panel-action--resume"
            >
              {{ teamSection.resumeAction }}
            </NuxtLink>
          </div>
        </article>

        <article class="about-panel">
          <div class="gf-static-kicker">
            {{ contact.kicker }}
          </div>
          <h2 class="about-title">
            {{ contact.title }}
          </h2>
          <p class="about-description about-description--contact">
            {{ contact.desc }}
          </p>

          <div class="about-links">
            <a
                v-for="item in contact.links"
                :key="item.label"
                :href="item.href"
                :target="item.external ? '_blank' : undefined"
                :rel="item.external ? 'noopener noreferrer' : undefined"
                class="about-link"
            >
              <span>{{ item.label }}</span>
              <span class="about-link__value">{{ item.value }}</span>
            </a>
          </div>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
const { locale } = useI18n()
const localePath = useLocalePath()
const isZh = computed(() => locale.value === 'zh')
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

const teamSection = computed(() => (
  isZh.value
    ? {
        kicker: 'People Behind It',
        desc: '兽人控导航站站长，Golang 软件工程师。兽人控导航站是一个开放的兽圈站点观测平台，主要提供站点的导航与用户视角的网络观测数据，无用户系统、不中转、不收集任何用户的敏感信息。目前项目处于小规模的持续迭代中，如有任何疑问欢迎前往Github提交Issue，十分感谢您的贡献。',
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: '提交您的建议',
        resumeAction: '查看简历',
        member: {
          name: '福狼',
          role: '开发者 / 维护者',
          avatar: 'https://qcdn.go-furry.com/nav/about/faolan.jpg',
          link: 'https://github.com/gofurry',
          action: '前往个人主页',
        },
      }
    : {
        kicker: 'People Behind It',
        desc: 'Site owner of GoFurry Navigation and Golang software engineer. GoFurry Navigation is an open furry-site observability platform focused on navigation and user-perspective web observation data, with no user system, no traffic proxying, and no collection of sensitive user information.',
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: 'Submit Suggestion',
        resumeAction: 'View Resume',
        member: {
          name: 'Faolan',
          role: 'Developer / Maintainer',
          avatar: 'https://qcdn.go-furry.com/nav/about/faolan.jpg',
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
