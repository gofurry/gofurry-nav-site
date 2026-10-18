<template>
  <div
    aria-hidden="true"
    class="gofurry-grid-background"
    :class="{
      'is-fixed': fixed && profile === 'full',
      'is-light': profile === 'light'
    }"
    :data-palette="palette"
    :data-profile="profile"
  />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  fixed?: boolean
  palette?: 'default' | 'nav-content' | 'games'
  profile?: 'full' | 'light'
}>(), {
  fixed: true,
  palette: 'default',
  profile: 'full',
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

.gofurry-grid-background.is-light {
  position: absolute;
  background:
    repeating-linear-gradient(90deg, rgba(105, 78, 57, 0.16) 0, rgba(105, 78, 57, 0.16) 1px, transparent 1px, transparent 38px),
    repeating-linear-gradient(0deg, rgba(105, 78, 57, 0.13) 0, rgba(105, 78, 57, 0.13) 1px, transparent 1px, transparent 38px),
    linear-gradient(180deg, rgba(255, 250, 244, 0.92), rgba(242, 227, 208, 0.96)),
    #f2e3d0;
}

.gofurry-grid-background::before,
.gofurry-grid-background::after {
  content: "";
  position: absolute;
  inset: -1px;
  opacity: 0.88;
  will-change: opacity;
}

.gofurry-grid-background.is-light::before,
.gofurry-grid-background.is-light::after {
  content: none;
}

.gofurry-grid-background::before {
  background: repeating-linear-gradient(
    90deg,
    rgba(105, 78, 57, 0.28) 0,
    rgba(105, 78, 57, 0.28) 1px,
    transparent 1px,
    transparent 38px
  );
  -webkit-mask-image: repeating-linear-gradient(0deg, #000 0, #000 3px, transparent 3px, transparent 7px);
  mask-image: repeating-linear-gradient(0deg, #000 0, #000 3px, transparent 3px, transparent 7px);
}

.gofurry-grid-background::after {
  background: repeating-linear-gradient(
    0deg,
    rgba(105, 78, 57, 0.24) 0,
    rgba(105, 78, 57, 0.24) 1px,
    transparent 1px,
    transparent 38px
  );
  -webkit-mask-image: repeating-linear-gradient(90deg, #000 0, #000 3px, transparent 3px, transparent 7px);
  mask-image: repeating-linear-gradient(90deg, #000 0, #000 3px, transparent 3px, transparent 7px);
}

:global(html.dark .gofurry-grid-background){
  background:
    radial-gradient(circle at 50% 42%, rgba(45, 212, 191, 0.14), transparent 92%),
    linear-gradient(180deg, rgba(8, 14, 28, 0.96), rgba(18, 28, 45, 0.985)),
    #0b1220;
}

:global(html.dark .gofurry-grid-background.is-light){
  background:
    repeating-linear-gradient(90deg, rgba(148, 163, 184, 0.16) 0, rgba(148, 163, 184, 0.16) 1px, transparent 1px, transparent 38px),
    repeating-linear-gradient(0deg, rgba(148, 163, 184, 0.14) 0, rgba(148, 163, 184, 0.14) 1px, transparent 1px, transparent 38px),
    linear-gradient(180deg, rgba(8, 14, 28, 0.96), rgba(18, 28, 45, 0.985)),
    #0b1220;
}

:global(html.dark .gofurry-grid-background::before){
  background: repeating-linear-gradient(
    90deg,
    rgba(148, 163, 184, 0.28) 0,
    rgba(148, 163, 184, 0.28) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(html.dark .gofurry-grid-background::after){
  background: repeating-linear-gradient(
    0deg,
    rgba(148, 163, 184, 0.26) 0,
    rgba(148, 163, 184, 0.26) 1px,
    transparent 1px,
    transparent 38px
  );
}

.gofurry-grid-background[data-palette="nav-content"] {
  background: #efe0ca;
}

.gofurry-grid-background[data-palette="games"] {
  background:
    radial-gradient(circle at 18% 8%, rgba(255, 246, 229, 0.34), transparent 26%),
    radial-gradient(circle at 86% 20%, rgba(225, 198, 154, 0.20), transparent 30%),
    #efe0ca;
}

.gofurry-grid-background.is-light[data-palette="nav-content"] {
  background:
    repeating-linear-gradient(90deg, rgba(95, 69, 49, 0.18) 0, rgba(95, 69, 49, 0.18) 1px, transparent 1px, transparent 38px),
    repeating-linear-gradient(0deg, rgba(95, 69, 49, 0.15) 0, rgba(95, 69, 49, 0.15) 1px, transparent 1px, transparent 38px),
    #efe0ca;
}

.gofurry-grid-background[data-palette="nav-content"]::before {
  background: repeating-linear-gradient(
    90deg,
    rgba(95, 69, 49, 0.34) 0,
    rgba(95, 69, 49, 0.34) 1px,
    transparent 1px,
    transparent 38px
  );
}

.gofurry-grid-background[data-palette="games"]::before {
  background: repeating-linear-gradient(
    90deg,
    rgba(95, 69, 49, 0.32) 0,
    rgba(95, 69, 49, 0.32) 1px,
    transparent 1px,
    transparent 38px
  );
}

.gofurry-grid-background[data-palette="nav-content"]::after {
  background: repeating-linear-gradient(
    0deg,
    rgba(95, 69, 49, 0.30) 0,
    rgba(95, 69, 49, 0.30) 1px,
    transparent 1px,
    transparent 38px
  );
}

.gofurry-grid-background[data-palette="games"]::after {
  background: repeating-linear-gradient(
    0deg,
    rgba(95, 69, 49, 0.28) 0,
    rgba(95, 69, 49, 0.28) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(html.dark .gofurry-grid-background[data-palette="nav-content"]) {
  background: #18273c;
}

:global(html.dark .gofurry-grid-background[data-palette="games"]) {
  background:
    radial-gradient(circle at 18% 8%, rgba(20, 184, 166, 0.10), transparent 26%),
    radial-gradient(circle at 86% 20%, rgba(245, 158, 11, 0.08), transparent 30%),
    linear-gradient(180deg, rgba(8, 15, 28, 0.98), rgba(14, 27, 46, 0.99)),
    #0b1728;
}

:global(html.dark .gofurry-grid-background.is-light[data-palette="nav-content"]) {
  background:
    repeating-linear-gradient(90deg, rgba(148, 163, 184, 0.17) 0, rgba(148, 163, 184, 0.17) 1px, transparent 1px, transparent 38px),
    repeating-linear-gradient(0deg, rgba(148, 163, 184, 0.15) 0, rgba(148, 163, 184, 0.15) 1px, transparent 1px, transparent 38px),
    #18273c;
}

:global(html.dark .gofurry-grid-background[data-palette="nav-content"]::before) {
  background: repeating-linear-gradient(
    90deg,
    rgba(148, 163, 184, 0.30) 0,
    rgba(148, 163, 184, 0.30) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(html.dark .gofurry-grid-background[data-palette="games"]::before) {
  background: repeating-linear-gradient(
    90deg,
    rgba(125, 211, 252, 0.16) 0,
    rgba(125, 211, 252, 0.16) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(html.dark .gofurry-grid-background[data-palette="nav-content"]::after) {
  background: repeating-linear-gradient(
    0deg,
    rgba(148, 163, 184, 0.28) 0,
    rgba(148, 163, 184, 0.28) 1px,
    transparent 1px,
    transparent 38px
  );
}

:global(html.dark .gofurry-grid-background[data-palette="games"]::after) {
  background: repeating-linear-gradient(
    0deg,
    rgba(251, 191, 36, 0.11) 0,
    rgba(251, 191, 36, 0.11) 1px,
    transparent 1px,
    transparent 38px
  );
}
</style>
