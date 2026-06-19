<template>
  <main
      class="flex-1 bg-[#f2e3d0]"
      :style="{
      backgroundImage: `url(${bgGrid})`,
      backgroundRepeat: 'repeat'
    }"
  >
    <CreatorGallery />
  </main>
</template>

<script setup lang="ts">
import CreatorGallery from '@/components/game/creators/CreatorGallery.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'
</script>
