<template>
  <span class="inline-flex items-center" @pointerenter="show" @pointerleave="leave">
    <button ref="trigger" type="button" data-site-help-trigger class="site-detail-help-trigger inline-flex items-center justify-center"
      :aria-label="label" :aria-describedby="open ? id : undefined" @focus="show" @blur="close" @click="show">
      <PhInfo class="site-detail-icon" aria-hidden="true" />
    </button>
    <Teleport v-if="open && host" :to="host">
      <span :id="id" ref="tip" role="tooltip" data-site-help-tooltip class="site-detail-help-tooltip fixed z-[100] break-words"
        :style="{ left: position.left + 'px', top: position.top + 'px' }" @pointerleave="leave">{{ text }}</span>
    </Teleport>
  </span>
</template>
<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef, useId } from 'vue'
import { PhInfo } from '@phosphor-icons/vue'
defineProps<{ label: string; text: string }>()
const id = useId(), open = ref(false), trigger = ref<HTMLButtonElement | null>(null), tip = ref<HTMLElement | null>(null)
const host = shallowRef<Element | null>(null), position = ref({ left: 0, top: 0 })
async function show() {
  open.value = true
  await nextTick()
  positionTooltip()
}
function positionTooltip() {
  if (!open.value || !trigger.value || !tip.value) return
  const anchor = trigger.value.getBoundingClientRect(), box = tip.value.getBoundingClientRect()
  position.value = { left: Math.max(12, Math.min(anchor.left, innerWidth - box.width - 12)),
    top: anchor.bottom + box.height < innerHeight - 12 ? anchor.bottom : Math.max(12, anchor.top - box.height) }
}
function close() { open.value = false }
function leave(event: PointerEvent) {
  if (event.relatedTarget instanceof Node && (tip.value?.contains(event.relatedTarget) || trigger.value?.contains(event.relatedTarget))) return
  if (document.activeElement !== trigger.value) close()
}
function escape(event: KeyboardEvent) { if (event.key === 'Escape') close() }
onMounted(() => {
  host.value = trigger.value?.closest('.site-detail-page') ?? null
  window.addEventListener('keydown', escape)
  window.addEventListener('resize', positionTooltip)
  window.addEventListener('scroll', positionTooltip, true)
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', escape)
  window.removeEventListener('resize', positionTooltip)
  window.removeEventListener('scroll', positionTooltip, true)
})
</script>
