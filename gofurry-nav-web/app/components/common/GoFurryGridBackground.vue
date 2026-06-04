<template>
  <div
    aria-hidden="true"
    class="gofurry-grid-background"
    :class="{ 'is-fixed': fixed }"
  />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  fixed?: boolean
}>(), {
  fixed: true,
})
</script>

<style scoped>
.gofurry-grid-background {
  pointer-events: none;
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  background:
    radial-gradient(circle at 14% 16%, rgba(255, 238, 214, 0.52), transparent 28%),
    radial-gradient(circle at 82% 10%, rgba(229, 223, 215, 0.46), transparent 30%),
    linear-gradient(180deg, rgba(255, 250, 244, 0.90), rgba(242, 227, 208, 0.96)),
    #f2e3d0;
}

.gofurry-grid-background.is-fixed {
  position: fixed;
}

.gofurry-grid-background::before,
.gofurry-grid-background::after {
  content: "";
  position: absolute;
  inset: -1px;
  opacity: 0.72;
  will-change: opacity;
}

.gofurry-grid-background::before {
  background: repeating-linear-gradient(
    90deg,
    rgba(105, 78, 57, 0.18) 0,
    rgba(105, 78, 57, 0.18) 1px,
    transparent 1px,
    transparent 38px
  );
  -webkit-mask-image: repeating-linear-gradient(0deg, #000 0, #000 3px, transparent 3px, transparent 7px);
  mask-image: repeating-linear-gradient(0deg, #000 0, #000 3px, transparent 3px, transparent 7px);
}

.gofurry-grid-background::after {
  background: repeating-linear-gradient(
    0deg,
    rgba(105, 78, 57, 0.16) 0,
    rgba(105, 78, 57, 0.16) 1px,
    transparent 1px,
    transparent 38px
  );
  -webkit-mask-image: repeating-linear-gradient(90deg, #000 0, #000 3px, transparent 3px, transparent 7px);
  mask-image: repeating-linear-gradient(90deg, #000 0, #000 3px, transparent 3px, transparent 7px);
}

:global(.dark .gofurry-grid-background){
  background:
    radial-gradient(circle at 50% 42%, rgba(45, 212, 191, 0.14), transparent 92%),
    linear-gradient(180deg, rgba(8, 14, 28, 0.96), rgba(18, 28, 45, 0.985)),
    #0b1220;
}

:global(.dark .gofurry-grid-background::before){
  background: repeating-linear-gradient(
    90deg,
    rgba(148, 163, 184, 0.12) 0,
    rgba(148, 163, 184, 0.12) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(.dark .gofurry-grid-background::after){
  background: repeating-linear-gradient(
    0deg,
    rgba(148, 163, 184, 0.10) 0,
    rgba(148, 163, 184, 0.10) 1px,
    transparent 1px,
    transparent 38px
  );
}
</style>
