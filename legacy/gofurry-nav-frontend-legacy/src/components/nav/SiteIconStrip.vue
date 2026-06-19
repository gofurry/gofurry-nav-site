<template>
  <div class="flex h-full flex-col">
    <div
        v-if="displayItems.length > itemsPerPage"
        class="mb-4 flex items-center justify-center gap-2 rounded-lg bg-slate-900/70 px-2 py-1"
    >
      <button
          class="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-400 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="currentPage === 0"
          :title="t('customSites.prevPage')"
          @click="goPrevPage"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <span class="min-w-20 text-center text-xs text-slate-500">
        {{ t('customSites.pageInfo', { current: currentPage + 1, total: totalPages }) }}
      </span>
      <button
          class="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-400 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="currentPage >= totalPages - 1"
          :title="t('customSites.nextPage')"
          @click="goNextPage"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <div
        v-if="displayItems.length"
        ref="rowRef"
        class="grid min-h-[148px] flex-1 grid-rows-2 place-items-center gap-3"
        :style="{ gridTemplateColumns: `repeat(${columnCount}, minmax(0, ${itemWidth}px))` }"
    >
      <template v-for="item in pagedItems" :key="item.id">
        <a
            v-if="item.type === 'site'"
            :href="toExternalUrl(item.url)"
            target="_blank"
            rel="noreferrer"
            class="group relative block h-16 w-16 shrink-0 rounded-xl bg-white/20 p-2 ring-orange-500/20 duration-300 hover:bg-white/40 hover:ring-2"
            :class="reorderable && dragOverId === item.id ? 'ring-2 ring-amber-200' : ''"
            :title="item.name"
            :draggable="reorderable"
            @dragstart="startDrag($event, item.id)"
            @dragover.prevent="handleDragOver(item.id)"
            @drop.prevent="handleDrop(item.id)"
            @dragend="resetDragState"
            @click="handleVisit(item)"
        >
          <img
              v-if="!failedIcons[item.id]"
              :src='"https://favicon.im/"+item.url+"?larger=true"'
              :alt="item.name"
              class="h-full w-full rounded-xl object-cover"
              @error="markIconFailed(item.id)"
              loading="lazy"
          />
          <div
              v-else
              class="flex h-full w-full items-center justify-center rounded-xl bg-slate-900 text-lg font-semibold text-white"
          >
            {{ item.name.slice(0, 1).toUpperCase() }}
          </div>

          <div class="pointer-events-none absolute left-1/2 top-0 z-10 -translate-x-1/2 -translate-y-[calc(100%+0.5rem)] whitespace-nowrap rounded-lg bg-slate-950/88 px-2 py-1 text-xs text-white opacity-0 shadow-lg transition group-hover:opacity-100">
            {{ item.name }}
          </div>

          <div
              v-if="editable"
              class="absolute inset-x-1 bottom-1 flex items-center justify-between rounded-lg bg-slate-950/70 px-1 py-1 opacity-0 transition group-hover:opacity-100"
          >
            <button
                class="inline-flex h-6 w-6 items-center justify-center rounded-md text-white transition hover:bg-white/15"
                :title="t('customSites.editSite')"
                @click.prevent.stop="$emit('edit', item)"
            >
              <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536M9 11l6.768-6.768a2.5 2.5 0 113.536 3.536L12.536 14.536A4 4 0 019.707 15.707L6 16l.293-3.707A4 4 0 017.464 9.464L9 11z" />
              </svg>
            </button>
            <button
                class="inline-flex h-6 w-6 items-center justify-center rounded-md text-white transition hover:bg-white/15"
                :title="t('customSites.deleteSite')"
                @click.prevent.stop="$emit('remove', item.id)"
            >
              <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6M9 7V4a1 1 0 011-1h4a1 1 0 011 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </a>

        <button
            v-else
            class="group relative flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl border border-dashed border-slate-300 bg-white/70 text-slate-500 shadow-sm transition hover:-translate-y-0.5 hover:border-amber-300 hover:bg-white hover:text-slate-800 hover:shadow-md"
            :title="addTitle"
            @click="$emit('add')"
        >
          <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <div class="pointer-events-none absolute left-1/2 top-0 z-10 -translate-x-1/2 -translate-y-[calc(100%+0.5rem)] whitespace-nowrap rounded-lg bg-slate-950/88 px-2 py-1 text-xs text-white opacity-0 shadow-lg transition group-hover:opacity-100">
            {{ addTitle }}
          </div>
        </button>
      </template>
    </div>

    <div v-else class="flex min-h-[144px] flex-1 items-center justify-center rounded-xl border border-dashed border-slate-300 bg-white/60 px-4 text-center">
      <div>
        <p class="text-base font-medium text-slate-700">
          {{ emptyTitle }}
        </p>
        <p class="mt-2 text-sm text-slate-500">
          {{ emptyDescription }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'

export interface SiteStripItem {
  id: string
  name: string
  url: string
}

type DisplayItem =
    | (SiteStripItem & { type: 'site' })
    | { id: string; type: 'add' }

const props = withDefaults(defineProps<{
  sites: SiteStripItem[]
  emptyTitle: string
  emptyDescription: string
  editable?: boolean
  reorderable?: boolean
  showAddTile?: boolean
  addTitle?: string
}>(), {
  editable: false,
  reorderable: false,
  showAddTile: false,
  addTitle: '',
})

const emit = defineEmits<{
  (e: 'edit', site: SiteStripItem): void
  (e: 'remove', id: string): void
  (e: 'add'): void
  (e: 'reorder', payload: { draggedId: string; targetId: string }): void
}>()

const { t } = useI18n()

const itemWidth = 76
const itemGap = 12
const rowCount = 2
const currentPage = ref(0)
const itemsPerPage = ref(8)
const columnCount = ref(4)
const rowRef = ref<HTMLElement | null>(null)
const failedIcons = ref<Record<string, boolean>>({})
const draggedId = ref<string | null>(null)
const dragOverId = ref<string | null>(null)

const displayItems = computed<DisplayItem[]>(() => {
  const items: DisplayItem[] = props.sites.map(site => ({ ...site, type: 'site' as const }))
  if (props.showAddTile) {
    items.push({ id: '__add__', type: 'add' as const })
  }
  return items
})

const totalPages = computed(() =>
    Math.max(1, Math.ceil(displayItems.value.length / itemsPerPage.value))
)

const pagedItems = computed(() => {
  const start = currentPage.value * itemsPerPage.value
  return displayItems.value.slice(start, start + itemsPerPage.value)
})

function markIconFailed(id: string) {
  failedIcons.value = {
    ...failedIcons.value,
    [id]: true,
  }
}

function startDrag(event: DragEvent, id: string) {
  if (!props.reorderable) {
    return
  }

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', id)
  }
  draggedId.value = id
}

function handleDragOver(id: string) {
  if (!props.reorderable || draggedId.value === id) {
    return
  }

  dragOverId.value = id
}

function handleDrop(targetId: string) {
  if (!props.reorderable || !draggedId.value || draggedId.value === targetId) {
    resetDragState()
    return
  }

  emit('reorder', {
    draggedId: draggedId.value,
    targetId,
  })
  resetDragState()
}

function resetDragState() {
  draggedId.value = null
  dragOverId.value = null
}

function handleVisit(site: SiteStripItem) {
  recordRecentSite(site)
}

function goPrevPage() {
  currentPage.value = Math.max(0, currentPage.value - 1)
}

function goNextPage() {
  currentPage.value = Math.min(totalPages.value - 1, currentPage.value + 1)
}

function updateItemsPerPage() {
  const width = rowRef.value?.clientWidth ?? 0
  if (!width) {
    columnCount.value = 4
    itemsPerPage.value = 8
    return
  }

  const columns = Math.max(1, Math.floor((width + itemGap) / (itemWidth + itemGap)))
  columnCount.value = columns
  itemsPerPage.value = Math.max(rowCount, columns * rowCount)
}

let resizeObserver: ResizeObserver | null = null

watch(
    () => props.sites,
    async () => {
      await nextTick()
      updateItemsPerPage()
      currentPage.value = Math.min(currentPage.value, totalPages.value - 1)
    },
    { deep: true }
)

watch(itemsPerPage, () => {
  currentPage.value = Math.min(currentPage.value, totalPages.value - 1)
})

watch(rowRef, (element, previousElement) => {
  if (previousElement && resizeObserver) {
    resizeObserver.unobserve(previousElement)
  }

  if (element && resizeObserver) {
    resizeObserver.observe(element)
    updateItemsPerPage()
  }
})

onMounted(async () => {
  await nextTick()
  updateItemsPerPage()

  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      updateItemsPerPage()
    })

    if (rowRef.value) {
      resizeObserver.observe(rowRef.value)
    }
  }

  window.addEventListener('resize', updateItemsPerPage)
})

onUnmounted(() => {
  resetDragState()
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateItemsPerPage)
})
</script>
