<template>
  <div class="games-page workshop-page relative isolate min-h-screen overflow-hidden">
    <GoFurryGridBackground :fixed="false" palette="games" />

    <main class="relative z-10 mx-auto flex w-full max-w-[1800px] flex-1 flex-col px-4 pb-20 pt-8 sm:px-6 lg:pt-10 xl:px-8">
      <WorkshopNav />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import WorkshopNav from '@/components/workshop/WorkshopNav.vue'

const { t } = useI18n()

useHead(() => ({
  title: `${t('workshop.tabs.discussion')} - ${t('workshop.title')}`,
}))
</script>
