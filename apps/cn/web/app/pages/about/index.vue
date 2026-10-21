<template>
  <div class="gf-static-page about-page relative isolate flex w-full flex-1 flex-col overflow-hidden">
    <div class="gf-static-page__top-veil" />

    <main class="gf-static-page__main gf-static-page__main--about">
      <h1 class="sr-only">{{ pageSeo.heading }}</h1>
      <section class="gf-static-page__grid gf-static-page__grid--about">
        <article
            v-for="(member, index) in teamSection.members"
            :key="member.name"
            class="about-panel"
            :class="{ 'lg:col-start-1 lg:row-start-2': index === 1 }"
        >
          <div class="gf-static-kicker">
            {{ teamSection.kicker }}
          </div>
          <div class="about-person">
            <img
                :src="member.avatar"
                :alt="member.name"
                class="about-avatar shrink-0"
            />
            <div>
              <h2 class="about-name">
                {{ member.name }}
              </h2>
              <div class="about-role">
                {{ member.role }}
              </div>
            </div>
          </div>
          <p class="about-description">
            {{ member.desc }}
            <template v-if="member.group"><br />{{ member.group }}</template>
          </p>
          <div v-if="member.link || member.resumeAction" class="about-actions">
            <a
                v-if="member.link"
                :href="member.link"
                target="_blank"
                rel="noopener noreferrer"
                class="about-panel-action"
            >
              {{ member.action }}
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
                v-if="member.resumeAction"
                :to="localePath('/about/faolan')"
                class="about-panel-action about-panel-action--resume"
            >
              {{ member.resumeAction }}
            </NuxtLink>
          </div>
        </article>

        <article class="about-panel lg:col-start-2 lg:row-start-1">
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
const { locale } = useI18n()
const localePath = useLocalePath()
const isZh = computed(() => locale.value === 'zh')
const pageSeo = computed(() => (
  isZh.value
    ? {
        heading: '关于 GoFurry 兽人控导航站',
        title: '关于 GoFurry 兽人控导航站 - 团队、反馈与收录合作',
        description: '了解 GoFurry 兽人控导航站的团队、项目定位、反馈方式与合作入口。我们共同维护兽人文化资源导航和用户视角的站点观测数据。'
      }
    : {
        heading: 'About GoFurry Navigation',
        title: 'About GoFurry Navigation - Team, feedback, and listing collaboration',
        description: 'Meet the GoFurry Navigation team and find feedback and collaboration channels for furry culture resources and user-perspective site observation data.'
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
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: '提交您的建议',
        members: [{
          name: '福狼',
          role: '开发者 / 维护者',
          avatar: '/web/about/faolan.avif',
          desc: '兽人控导航站站长，Golang 软件工程师。我们共同维护兽人文化资源导航，提供站点信息与用户视角的网络观测数据。欢迎补充收录和反馈建议。',
          group: null,
          link: 'https://github.com/gofurry',
          action: '前往个人主页',
          resumeAction: '查看简历',
        }, {
          name: '于空泽',
          role: '运营成员',
          avatar: '/web/about/yukongze.avif',
          desc: '导航站运营人员，主要负责上传兽游。如有遗漏，欢迎补充！',
          group: '导航站交流群1074661551',
          link: null,
          action: null,
          resumeAction: null,
        }],
      }
    : {
        kicker: 'People Behind It',
        issueLink: 'https://github.com/gofurry/gofurry-nav-site/issues',
        issueAction: 'Submit Suggestion',
        members: [{
          name: 'Faolan Grey',
          role: 'Developer / Maintainer',
          avatar: '/web/about/faolan.avif',
          desc: 'GoFurry Navigation site owner and Golang software engineer. We maintain a directory of furry resources with site information and user-perspective web observations. Suggestions and listing updates are welcome.',
          group: null,
          link: 'https://github.com/gofurry',
          action: 'Open project page',
          resumeAction: 'View Resume',
        }, {
          name: 'Yukongze',
          role: 'Operations member',
          avatar: '/web/about/yukongze.avif',
          desc: 'GoFurry operations member, mainly responsible for adding furry games. If anything is missing, please let us know!',
          group: 'Navigation QQ group: 1074661551',
          link: null,
          action: null,
          resumeAction: null,
        }],
      }
))

const contact = computed(() => (
  isZh.value
    ? {
        kicker: 'Contact',
        title: '站点反馈与交流合作',
        desc: '欢迎提交问题、建议站点、补充收录信息，或者讨论兽人控导航站的后续方向。',
        links: [
          { label: '邮箱', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry/gofurry-nav-site', href: 'https://github.com/gofurry/gofurry-nav-site', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
    : {
        kicker: 'Contact',
        title: 'Feedback & collaboration',
        desc: 'Get in touch to report issues, suggest sites, update game information, or discuss the project.',
        links: [
          { label: 'Email', value: '2660621624@qq.com', href: 'mailto:2660621624@qq.com', external: false },
          { label: 'GitHub', value: 'gofurry/gofurry-nav-site', href: 'https://github.com/gofurry/gofurry-nav-site', external: true },
          { label: 'DeepFurry', value: 'www.deepfurry.com', href: 'https://www.deepfurry.com', external: true },
        ],
      }
))
</script>
