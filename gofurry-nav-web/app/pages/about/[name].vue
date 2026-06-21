<template>
  <div class="gf-static-page about-page resume-page relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="gf-static-page__top-veil" />

    <main class="gf-static-page__main gf-static-page__main--resume">
      <NuxtLink :to="localePath('/about')" class="resume-back-link">
        <span aria-hidden="true">‹</span>
        <span>{{ copy.back }}</span>
      </NuxtLink>

      <section class="resume-hero about-panel">
        <img
            :src="avatarUrl"
            alt="Faolan"
            class="resume-avatar"
        />
        <div class="resume-hero__content">
          <div class="resume-identity-row">
            <h1 class="resume-name">Faolan</h1>
            <span class="resume-role-years">{{ copy.roleYears }}</span>
          </div>
          <p class="resume-summary">{{ copy.summary }}</p>
        </div>
      </section>

      <section class="resume-layout">
        <aside class="about-panel resume-facts-panel">
          <h2 class="resume-section-title">{{ copy.basicInfo }}</h2>
          <dl class="resume-fact-list">
            <div v-for="item in facts" :key="item.label" class="resume-fact">
              <dt>{{ item.label }}</dt>
              <dd>{{ item.value }}</dd>
            </div>
          </dl>
        </aside>

        <div class="resume-main-flow">
          <article class="about-panel resume-section">
            <div class="resume-section-heading">
              <h2 class="resume-section-title">{{ copy.workExperience }}</h2>
            </div>
            <ol class="resume-timeline">
              <li v-for="item in workExperiences" :key="item" class="resume-timeline-item">
                {{ item }}
              </li>
            </ol>
          </article>

          <article class="about-panel resume-section">
            <div class="resume-section-heading">
              <h2 class="resume-section-title">{{ copy.projectExperience }}</h2>
            </div>
            <ol class="resume-project-list">
              <li v-for="item in projects" :key="item.title" class="resume-project-item">
                <span>{{ item.title }}</span>
                <em v-if="item.note">{{ item.note }}</em>
              </li>
            </ol>
          </article>

          <article class="about-panel resume-section">
            <div class="resume-section-heading">
              <h2 class="resume-section-title">{{ copy.openSource }}</h2>
            </div>
            <div class="resume-links">
              <a
                  v-for="item in openSourceLinks"
                  :key="item.href"
                  :href="item.href"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="about-link"
              >
                <span>{{ item.label }}</span>
                <span class="about-link__value">{{ item.value }}</span>
              </a>
              <a
                  href="https://gopheratlas.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="about-link"
              >
                <span>{{ copy.personalBlog }}</span>
                <span class="about-link__value">gopheratlas.com</span>
              </a>
            </div>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const route = useRoute()
const localePath = useLocalePath()
const { locale } = useI18n()
const slug = String(route.params.name ?? '').toLowerCase()

if (slug !== 'faolan') {
  throw createError({
    statusCode: 404,
    statusMessage: 'Resume Not Found',
  })
}

const isZh = computed(() => locale.value === 'zh')
const avatarUrl = 'https://qcdn.go-furry.com/nav/about/faolan.jpg'

const copy = computed(() => (
  isZh.value
    ? {
        back: '返回关于页',
        roleYears: '软件工程师 · 2年',
        summary: '熟悉Golang/Vue的软件开发生态，了解并可以初步使用多门编程语言，可以熟练编写Web、爬虫、桌面端、命令行工具等多种软件。',
        basicInfo: '基础信息',
        workExperience: '在职经历',
        projectExperience: '项目经历',
        openSource: '开源与博客',
        personalBlog: '个人博客',
      }
    : {
        back: 'Back to About',
        roleYears: 'Software Engineer · 2 years',
        summary: 'Familiar with the Golang/Vue software development ecosystem, able to understand and make initial use of multiple programming languages, and comfortable building web apps, crawlers, desktop apps, command-line tools, and other software.',
        basicInfo: 'Basic Info',
        workExperience: 'Work Experience',
        projectExperience: 'Project Experience',
        openSource: 'Open Source & Blog',
        personalBlog: 'Personal Blog',
      }
))

const facts = computed(() => (
  isZh.value
    ? [
        { label: '姓名', value: 'Faolan' },
        { label: '性别', value: '男' },
        { label: '年龄', value: '24' },
        { label: '联系方式', value: '2660621624@qq.com' },
        { label: '职业', value: '软件工程师' },
        { label: '学校', value: '成都信息工程大学 - 本科 - 机器人工程' },
      ]
    : [
        { label: 'Name', value: 'Faolan' },
        { label: 'Gender', value: 'Male' },
        { label: 'Age', value: '24' },
        { label: 'Contact', value: '2660621624@qq.com' },
        { label: 'Profession', value: 'Software Engineer' },
        { label: 'Education', value: 'Chengdu University of Information Technology (CUIT) - Bachelor - Robotics Engineering' },
      ]
))

const workExperiences = computed(() => (
  isZh.value
    ? [
        '小厂 - 军工 - Golang 后端',
        '中国央企 - 网络工程师',
        '小厂 - 网络安全 / 军工 - Golang 全栈',
      ]
    : [
        'Small company - defense industry - Golang backend',
        'Chinese central state-owned enterprise - Network engineer',
        'Small company - cybersecurity / defense industry - Golang full-stack',
      ]
))

const projects = computed(() => (
  isZh.value
    ? [
        { title: 'GoFurry 导航站', note: '站长' },
        { title: '某 P2P 隐蔽控制网络', note: '涉密' },
        { title: '某中心化运维系统', note: '涉密' },
        { title: '分布式人脸识别门禁系统', note: '优秀毕设' },
      ]
    : [
        { title: 'GoFurry Navigation', note: 'site owner' },
        { title: 'P2P covert control network', note: 'confidential' },
        { title: 'Centralized operations platform', note: 'confidential' },
        { title: 'Distributed face-recognition access control system', note: 'excellent graduation project' },
      ]
))

const openSourceLinks = [
  {
    label: 'GoFiber Coraza',
    value: 'gofiber/contrib v3/coraza',
    href: 'https://github.com/gofiber/contrib/tree/main/v3/coraza',
  },
  {
    label: 'steam-go',
    value: 'gofurry/steam-go',
    href: 'https://github.com/gofurry/steam-go',
  },
]

const pageSeo = computed(() => (
  isZh.value
    ? {
        title: 'Faolan 简历 - GoFurry 兽人控导航站',
        description: 'Faolan 的软件工程师简历，包含教育背景、在职经历、项目经历、开源项目与个人博客。'
      }
    : {
        title: 'Faolan Resume - GoFurry Navigation',
        description: 'Resume of Faolan, including education, work experience, project experience, open-source projects, and personal blog.'
      }
))

useSeoMeta({
  title: () => pageSeo.value.title,
  description: () => pageSeo.value.description,
  ogTitle: () => pageSeo.value.title,
  ogDescription: () => pageSeo.value.description,
})
</script>
