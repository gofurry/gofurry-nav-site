<template>
  <div class="gf-public-background" data-public-background data-pattern-status="default" aria-hidden="true">
    <div class="gf-public-background__pattern" />
  </div>
</template>

<style scoped>
.gf-public-background {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: var(--gf-page-background);
}

.gf-public-background__pattern {
  position: absolute;
  inset: 0;
  background-image: var(--gf-page-pattern);
  background-position: top left;
  background-repeat: repeat;
  background-size: var(--gf-page-pattern-size);
  opacity: var(--gf-page-pattern-opacity);
}
</style>
