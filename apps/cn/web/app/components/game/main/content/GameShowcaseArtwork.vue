<template>
  <div ref="root" class="game-home-showcase__media relative h-full w-full" :data-motion="interacted" :data-direction="direction">
    <div
      v-for="entry in retained" :key="entry.key" :data-key="entry.key" :data-active="entry.key === displayedKey"
      class="game-home-showcase__frame pointer-events-none absolute inset-0" :class="entry.key === displayedKey ? 'visible z-10' : 'invisible z-0'" aria-hidden="true"
    >
      <ManagedAssetImage
        v-if="entry.artwork.kind === 'managed'"
        :object-key="entry.artwork.desktop_object_key" :mobile-object-key="entry.artwork.mobile_object_key" fallback="" alt=""
        class="game-home-showcase__image absolute inset-0 block h-full w-full" :style="focalStyle(entry)"
        @load="loaded($event, entry.key)" @exhausted="settled(entry.key)" @pending="pending(entry.key)"
      />
      <SteamAssetImage
        v-else-if="entry.artwork.url && !failed.has(entry.key)" :src="entry.artwork.url" alt=""
        class="game-home-showcase__image absolute inset-0 block h-full w-full"
        @load="loaded($event, entry.key)" @error="exhausted(entry.key)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import ManagedAssetImage from '~/components/common/ManagedAssetImage.vue'
import SteamAssetImage from '~/components/common/SteamAssetImage.vue'
import type { GameShowcaseItem } from '~/types/game'
import { showcaseFocalPoint } from '~/utils/gameShowcasePresentation'

const props = defineProps<{ items: GameShowcaseItem[]; activeKey: string; direction: 'next' | 'previous' }>()
const emit = defineEmits<{ display: [state: { key: string; ready: boolean }] }>()
const root = ref<HTMLElement | null>(null)
const visited = ref(new Set([props.activeKey]))
const ready = new Set<string>()
const failed = ref(new Set<string>())
const displayedKey = ref(props.activeKey)
const interacted = ref(false)
const retained = computed(() => props.items.filter(entry => visited.value.has(entry.key)))
const focalStyle = (entry: GameShowcaseItem) => ({ objectPosition: `${showcaseFocalPoint(entry.artwork.focal_x)}% ${showcaseFocalPoint(entry.artwork.focal_y)}%` })

function report() { emit('display', { key: displayedKey.value, ready: displayedKey.value === props.activeKey && ready.has(props.activeKey) }) }
function pending(key: string) { ready.delete(key); report() }
function settled(key: string) {
  ready.add(key)
  if (key === props.activeKey) displayedKey.value = key
  report()
}
function exhausted(key: string) { failed.value.add(key); settled(key) }
function settleMissingSteam(key: string) {
  const entry = props.items.find(entry => entry.key === key)
  if (entry?.artwork.kind === 'steam' && !entry.artwork.url) exhausted(key)
}
async function decode(image: HTMLImageElement, key: string) {
  const source = image.currentSrc || image.src
  try { await image.decode() } catch { /* Resource errors follow the existing image fallback owner. */ }
  if (image.isConnected && image.naturalWidth > 0 && source === (image.currentSrc || image.src)) settled(key)
}
function loaded(event: Event, key: string) { void decode(event.target as HTMLImageElement, key) }
watch(() => props.activeKey, (key) => {
  interacted.value = true
  visited.value.add(key)
  if (ready.has(key)) displayedKey.value = key
  report()
  settleMissingSteam(key)
}, { flush: 'sync' })
onMounted(() => {
  root.value?.querySelectorAll<HTMLImageElement>('img').forEach(image => {
    if (image.complete && image.naturalWidth > 0) void decode(image, image.closest<HTMLElement>('[data-key]')!.dataset.key!)
  })
  settleMissingSteam(props.activeKey)
})
</script>
