<template>
  <div ref="root" class="site-detail-selector relative min-w-0" @focusout="onFocusOut" @keydown.esc.stop.prevent="close(true)">
    <button
      ref="trigger"
      type="button"
      data-site-target-trigger
      class="gf-button gf-button--surface w-full"
      aria-haspopup="listbox"
      aria-controls="site-target-listbox"
      :aria-expanded="open"
      @click="toggle"
      @keydown.down.prevent="show"
      @keydown.up.prevent="show"
    >{{ t('siteDetail.selectTarget') }} <span aria-hidden="true">⌄</span></button>
    <ul
      v-if="open"
      id="site-target-listbox"
      role="listbox"
      :aria-label="t('siteDetail.selectTarget')"
      class="site-detail-target-list absolute right-0 z-40 mt-2 w-full overflow-y-auto"
    >
      <li v-for="(item, index) in targets" :key="item.target" role="presentation">
        <button
          type="button"
          role="option"
          :aria-selected="item.target === selected"
          :tabindex="index === focused ? 0 : -1"
          :data-site-target-option="item.target"
          class="site-detail-target-option block w-full text-left"
          @focus="focused = index"
          @click="choose(item.target)"
          @keydown="onKey($event, index)"
        >
          <span class="block break-all">{{ item.target }}</span>
          <span v-if="item.relation" class="site-detail-note mt-1 block break-words">{{ item.relation }}</span>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onBeforeUnmount, ref } from 'vue'
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
const props = defineProps<{ targets: SiteTargetPresentation['targetList']; selected: string }>()
const emit = defineEmits<{ select: [target: string] }>()
const { t } = useI18n()
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const open = ref(false)
const focused = ref(0)
async function focusOption(index: number) {
  focused.value = index
  await nextTick()
  // Native focus also reveals offscreen options inside the bounded listbox.
  root.value?.querySelectorAll<HTMLButtonElement>('[role="option"]')[index]?.focus()
}
function show() {
  open.value = true
  void focusOption(Math.max(0, props.targets.findIndex(item => item.target === props.selected)))
}
function close(returnFocus = false) {
  open.value = false
  if (returnFocus) trigger.value?.focus({ preventScroll: true })
}
function toggle() {
  if (open.value) close(true)
  else show()
}
function choose(target: string) {
  close(true)
  if (target !== props.selected) emit('select', target)
}
function onKey(event: KeyboardEvent, index: number) {
  const count = props.targets.length
  const next = event.key === 'ArrowDown' ? (index + 1) % count : event.key === 'ArrowUp' ? (index + count - 1) % count
    : event.key === 'Home' ? 0 : event.key === 'End' ? count - 1 : -1
  if (next >= 0) {
    event.preventDefault()
    void focusOption(next)
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    const target = props.targets[index]?.target
    if (target) choose(target)
  }
}
function onOutside(event: MouseEvent) {
  if (open.value && event.target instanceof Node && !root.value?.contains(event.target)) close(true)
}
function onFocusOut(event: FocusEvent) {
  if (event.relatedTarget instanceof Node && !root.value?.contains(event.relatedTarget)) close()
}
// Click runs after native pointer focus transfer. Returning focus on pointerdown
// would be undone by the browser when the outside surface receives that click.
onMounted(() => document.addEventListener('click', onOutside))
onBeforeUnmount(() => document.removeEventListener('click', onOutside))
</script>
