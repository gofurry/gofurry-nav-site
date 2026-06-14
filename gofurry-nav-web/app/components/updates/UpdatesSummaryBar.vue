<template>
  <section class="flex justify-center py-[18px] pb-[30px]" :aria-label="label">
    <div
      class="updates-summary-shell inline-flex items-center justify-center gap-[22px] border-b px-6 py-3 backdrop-blur-[10px]"
    >
      <div class="inline-flex items-baseline gap-3">
        <span class="updates-summary-label text-[0.76rem] uppercase">{{ latestLabel }}</span>
        <strong class="updates-summary-value break-words text-[clamp(1rem,1.8vw,1.12rem)] font-[780] leading-[1.5]">
          {{ latestValue }}
        </strong>
      </div>

      <span class="inline-flex h-5 w-[88px] items-center justify-center" aria-hidden="true">
        <img class="block h-5 w-[88px]" :src="dividerSrc" alt="" />
      </span>

      <div class="inline-flex items-baseline gap-3">
        <span class="updates-summary-label text-[0.76rem] uppercase">{{ countLabel }}</span>
        <strong class="updates-summary-value break-words text-[clamp(1rem,1.8vw,1.12rem)] font-[780] leading-[1.5]">
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

<style scoped>
.updates-summary-shell {
  border-color: var(--updates-summary-border);
  background: var(--updates-summary-bg);
  box-shadow: var(--updates-summary-shadow);
}

.updates-summary-label {
  color: var(--updates-summary-label);
}

.updates-summary-value {
  color: var(--updates-summary-value);
}

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
