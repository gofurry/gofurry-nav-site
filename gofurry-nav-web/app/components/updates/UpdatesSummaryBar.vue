<template>
  <section class="flex justify-center py-[18px] pb-[30px]" :aria-label="label">
    <div
      class="inline-flex items-center justify-center gap-[22px] border-b px-6 py-3 backdrop-blur-[10px]"
      :class="wrapperClass"
    >
      <div class="inline-flex items-baseline gap-3">
        <span class="text-[0.76rem] uppercase" :class="labelClass">{{ latestLabel }}</span>
        <strong class="break-words text-[clamp(1rem,1.8vw,1.12rem)] font-[780] leading-[1.5]" :class="valueClass">
          {{ latestValue }}
        </strong>
      </div>

      <span class="inline-flex h-5 w-[88px] items-center justify-center" aria-hidden="true">
        <img class="block h-5 w-[88px]" :src="dividerSrc" alt="" />
      </span>

      <div class="inline-flex items-baseline gap-3">
        <span class="text-[0.76rem] uppercase" :class="labelClass">{{ countLabel }}</span>
        <strong class="break-words text-[clamp(1rem,1.8vw,1.12rem)] font-[780] leading-[1.5]" :class="valueClass">
          {{ countValue }}
        </strong>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  latestLabel: string
  latestValue: string
  countLabel: string
  countValue: number
  dividerSrc: string
  dark: boolean
}>()

const wrapperClass = computed(() => (
  props.dark
    ? 'border-[#97e0ec2e] bg-[rgba(6,20,29,0.22)] shadow-[inset_0_-1px_0_rgba(171,234,243,0.08),0_10px_30px_rgba(0,0,0,0.12)]'
    : 'border-[rgba(62,50,41,0.12)] bg-transparent'
))

const labelClass = computed(() => (
  props.dark ? 'text-[rgba(174,205,212,0.68)]' : 'text-[rgba(32,24,21,0.54)]'
))

const valueClass = computed(() => (
  props.dark ? 'text-[rgba(226,238,242,0.84)]' : 'text-[rgba(31,23,19,0.8)]'
))
</script>

<style scoped>
@media (max-width: 720px) {
  section > div {
    display: grid;
    justify-items: center;
    gap: 14px;
    padding: 12px 18px;
  }

  section > div > div {
    flex-wrap: wrap;
    justify-content: center;
    gap: 6px 10px;
  }
}
</style>
