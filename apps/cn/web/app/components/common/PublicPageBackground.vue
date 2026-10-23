<template>
  <div class="gf-public-background fixed inset-0 pointer-events-none" data-public-background data-pattern-status="default" aria-hidden="true" />
  <div class="gf-public-background__fade fixed inset-x-0 top-0 pointer-events-none" data-page-dots aria-hidden="true">
    <div class="gf-public-background__dots gf-public-background__dots--left absolute inset-y-0 left-0" />
    <div class="gf-public-background__dots gf-public-background__dots--right absolute inset-y-0 right-0" />
  </div>
</template>

<style scoped>
.gf-public-background {
  z-index: 0;
  background: var(--gf-page-background);
}

.gf-public-background__fade {
  z-index: 0;
  height: min(var(--gf-page-dot-height), 100svh);
  -webkit-mask-image: linear-gradient(to bottom, var(--gf-page-dot-mask-opaque) 18%, transparent);
  mask-image: linear-gradient(to bottom, var(--gf-page-dot-mask-opaque) 18%, transparent);
}

.gf-public-background__dots {
  width: var(--gf-page-dot-edge-width);
  background-image: radial-gradient(circle, rgba(var(--gf-page-dot-color), var(--gf-page-dot-alpha)) var(--gf-page-dot-radius), transparent var(--gf-page-dot-radius));
  background-repeat: repeat;
  background-size: var(--gf-page-dot-spacing) var(--gf-page-dot-spacing);
}

.gf-public-background__dots--left {
  -webkit-mask-image: linear-gradient(to right, var(--gf-page-dot-mask-opaque) 15%, transparent);
  mask-image: linear-gradient(to right, var(--gf-page-dot-mask-opaque) 15%, transparent);
}

.gf-public-background__dots--right {
  background-position: right top;
  -webkit-mask-image: linear-gradient(to left, var(--gf-page-dot-mask-opaque) 15%, transparent);
  mask-image: linear-gradient(to left, var(--gf-page-dot-mask-opaque) 15%, transparent);
}

@media (max-width: 767px) {
  .gf-public-background__fade {
    -webkit-mask-image: linear-gradient(to bottom, var(--gf-page-dot-mask-opaque), transparent);
    mask-image: linear-gradient(to bottom, var(--gf-page-dot-mask-opaque), transparent);
  }

  .gf-public-background__dots--left {
    width: 100%;
    -webkit-mask-image: none;
    mask-image: none;
  }

  .gf-public-background__dots--right { display: none; }
}
</style>
