<template>
  <section class="updates-summary" :aria-label="label">
    <div
      class="updates-summary-shell"
    >
      <div class="updates-summary-item">
        <span class="updates-summary-label">{{ latestLabel }}</span>
        <strong class="updates-summary-value">
          {{ latestValue }}
        </strong>
      </div>

      <span class="updates-summary-divider" aria-hidden="true">
        <img :src="dividerSrc" alt="" />
      </span>

      <div class="updates-summary-item">
        <span class="updates-summary-label">{{ countLabel }}</span>
        <strong class="updates-summary-value">
          {{ countValue }}
        </strong>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  latestLabel: string
  latestValue: string
  countLabel: string
  countValue: number
  dividerSrc: string
}>()
</script>
