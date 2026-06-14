<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/20 px-4">
    <div
        class="game-search-filter-panel w-full max-w-2xl overflow-hidden rounded-2xl p-6 shadow"
    >
      <div class="space-y-2 overflow-y-auto scrollbar-hide max-h-[calc(80vh-3rem)]">

        <!-- 标题 & 操作 -->
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-bold">{{ t("game.search.advancedFilter") }}</h2>
          <div class="flex gap-2">
            <div
                class="game-search-filter-action game-search-filter-action--ghost"
                @click="emit('close')"
            >
              {{ t("common.cancel") }}
            </div>
            <div
                class="game-search-filter-action game-search-filter-action--primary"
                @click="onSearch"
            >
              {{ t("common.query") }}
            </div>
          </div>
        </div>

        <!-- 关键词 & 页大小 -->
        <div class="flex gap-4 items-center w-full">
          <div class="grid grid-cols-1 w-[75%]">
            <label class="text-xs text-gray-500">{{ t("common.keyword") }}</label>
            <input
                v-model="props.query.content"
                class="game-search-filter-input ml-1 mt-1 w-full rounded-lg px-3 py-2 focus:outline-none"
            />
          </div>
          <div class="grid grid-cols-1 w-[18%]">
            <label class="text-xs text-gray-500">{{ t("common.pageSize") }}</label>
            <input
                v-model.number="props.query.pageSize"
                min="1"
                class="game-search-filter-input mt-1 w-full rounded-lg px-3 py-2 focus:outline-none"
            />
          </div>
        </div>

        <!-- 发布时间 -->
        <div>
          <label class="text-xs text-gray-500">{{ t("game.search.publishTime") }}</label>
          <div class="flex gap-2 mt-1">
            <VueDatePicker
                v-model="publishStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                class="dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="publishEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                class="dp-custom-theme w-1/2"
            />
          </div>
        </div>

        <!-- 更新时间 -->
        <div>
          <label class="text-xs text-gray-500">{{ t("game.search.updateTime") }}</label>
          <div class="flex gap-2 mt-1">
            <VueDatePicker
                v-model="updateStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                class="dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="updateEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                class="dp-custom-theme w-1/2"
            />
          </div>
        </div>

        <!-- 排序 -->
        <div>
          <label class="text-xs text-gray-500">{{ t("common.sort") }}</label>
          <div class="flex flex-wrap gap-2">
            <span
                v-for="item in sortOptions"
                :key="item.key"
                @click="toggleSort(item.key)"
                :class="[
                'game-search-filter-chip',
                item.selected
                  ? 'game-search-filter-chip--active'
                  : 'game-search-filter-chip--idle'
              ]"
            >
              {{ t(item.label) }}
            </span>
          </div>
        </div>

        <!-- 标签 -->
        <div>
          <label class="text-xs text-gray-500">{{ t("common.tag") }}</label>
          <div class="mt-2 space-y-2">
            <div v-for="group in categoryGroups" :key="group.id">
              <div class="mb-1 text-sm font-semibold text-stone-700 dark:text-slate-200/80">
                {{ group.name }}
              </div>
              <div class="flex flex-wrap gap-2">
                <span
                    v-for="tag in (group.expanded ? group.children : group.children.slice(0, group.limit))"
                    :key="tag.id"
                    @click="toggleTag(tag)"
                    :class="[
                    'game-search-filter-chip',
                    tag.selected
                      ? 'game-search-filter-chip--active'
                      : 'game-search-filter-chip--idle'
                  ]"
                >
                  {{ tag.name }} {{ tag.game_count }}
                </span>
              </div>
              <div
                  v-if="group.children.length > group.limit"
                  class="mt-1 cursor-pointer select-none text-xs text-stone-600 transition hover:text-stone-900 dark:text-slate-300/64 dark:hover:text-slate-100/88"
                  @click="group.expanded = !group.expanded"
              >
                {{ group.expanded ? t("common.collapse") : t("common.expand") }}
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import type { GameTagRecord, SearchPageQueryRequest } from '@/types/game'
import { formatLocalDateTime } from '@/utils/util'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  tagGroups: GameTagRecord[]
  query: SearchPageQueryRequest
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'search'): void
}>()

// =============== 时间 ===============
const publishStart = ref<Date>(
    props.query.pub_start_time
        ? new Date(props.query.pub_start_time.replace(' ', 'T'))
        : new Date(2000, 0, 1)
)
const publishEnd = ref<Date>(
    props.query.pub_end_time
        ? new Date(props.query.pub_end_time.replace(' ', 'T'))
        : new Date()
)
const updateStart = ref<Date>(
    props.query.update_start_time
        ? new Date(props.query.update_start_time.replace(' ', 'T'))
        : new Date(2000, 0, 1)
)
const updateEnd = ref<Date>(
    props.query.update_end_time
        ? new Date(props.query.update_end_time.replace(' ', 'T'))
        : new Date()
)

// =============== 排序 ===============
const sortOptions = reactive([
  {
    key: 'highestRating',
    label: 'game.search.highestRating',
    selected: props.query.score ?? false
  },
  {
    key: 'mostComments',
    label: 'game.search.mostComments',
    selected: props.query.remark_order ?? false
  },
  {
    key: 'latestInfo',
    label: 'game.search.latestInfo',
    selected: props.query.time_order ?? false
  }
])

const toggleSort = (key: string) => {
  const item = sortOptions.find(i => i.key === key)
  if (!item) return

  item.selected = !item.selected

  props.query.score = !!sortOptions.find(i => i.key === 'highestRating')?.selected
  props.query.remark_order = !!sortOptions.find(i => i.key === 'mostComments')?.selected
  props.query.time_order = !!sortOptions.find(i => i.key === 'latestInfo')?.selected
}

// =============== 分类 & 标签 ===============
type CategoryGroup = GameTagRecord & {
  children: (GameTagRecord & { selected: boolean })[]
  expanded: boolean
  limit: number
}

const categoryGroups = ref<CategoryGroup[]>([])

const buildCategoryGroups = () => {
  const groups: CategoryGroup[] = props.tagGroups
      .filter(t => Number(t.prefix) === -1)
      .sort((a, b) => Number(a.id) - Number(b.id))
      .map(g => ({
        ...g,
        children: [],
        expanded: false,
        limit: 16
      }))

  const tags = props.tagGroups.filter(t => Number(t.prefix) !== -1)

  groups.forEach(group => {
    group.children = tags
        .filter(t => Number(t.prefix) === Number(group.id))
        .map(t => ({
          ...t,
          selected: (props.query.tag_list ?? []).includes(Number(t.id))
        }))
  })

  categoryGroups.value = groups
}

const toggleTag = (tag: any) => {
  tag.selected = !tag.selected
  props.query.tag_list = categoryGroups.value
      .flatMap(g => g.children)
      .filter(t => t.selected)
      .map(t => Number(t.id))
}

// =============== watch & 生命周期 ===============
onMounted(buildCategoryGroups)

watch(() => props.tagGroups, buildCategoryGroups, { deep: true })

const formatDateTime = formatLocalDateTime

watch([publishStart, publishEnd], () => {
  props.query.pub_start_time = formatDateTime(publishStart.value)
  props.query.pub_end_time = formatDateTime(publishEnd.value)
})

watch([updateStart, updateEnd], () => {
  props.query.update_start_time = formatDateTime(updateStart.value)
  props.query.update_end_time = formatDateTime(updateEnd.value)
})

const onSearch = () => {
  emit('search')
  emit('close')
}
</script>

<style scoped>
.scrollbar-hide {
  scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

.game-search-filter-panel {
  max-height: 80vh;
  border: 1px solid rgba(126, 92, 58, 0.16);
  background: rgba(255, 250, 242, 0.94);
  color: rgba(45, 35, 28, 0.92);
}

.game-search-filter-action {
  cursor: pointer;
  border-radius: 0.65rem;
  padding: 0.38rem 0.75rem;
  font-size: 0.86rem;
  transition: background-color 180ms ease, color 180ms ease, border-color 180ms ease;
}

.game-search-filter-action--ghost:hover {
  background: rgba(255, 239, 213, 0.72);
}

.game-search-filter-action--primary {
  background: rgba(124, 45, 18, 0.86);
  color: rgba(255, 250, 242, 0.96);
}

.game-search-filter-action--primary:hover {
  background: rgba(99, 39, 15, 0.96);
}

.game-search-filter-input {
  border: 1px solid rgba(126, 92, 58, 0.12);
  background: rgba(255, 239, 213, 0.56);
  color: rgba(45, 35, 28, 0.92);
}

.game-search-filter-input:focus {
  border-color: rgba(120, 87, 56, 0.32);
  box-shadow: 0 0 0 2px rgba(120, 87, 56, 0.10);
}

.game-search-filter-chip {
  cursor: pointer;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.26rem 0.62rem;
  font-size: 0.76rem;
  transition: background-color 180ms ease, border-color 180ms ease, color 180ms ease;
}

.game-search-filter-chip--idle {
  border-color: rgba(126, 92, 58, 0.12);
  background: rgba(255, 239, 213, 0.48);
  color: rgba(87, 43, 20, 0.82);
}

.game-search-filter-chip--idle:hover {
  border-color: rgba(180, 96, 24, 0.30);
  background: rgba(255, 224, 186, 0.72);
}

.game-search-filter-chip--active {
  border-color: rgba(126, 92, 58, 0.20);
  background: rgba(124, 45, 18, 0.86);
  color: rgba(255, 250, 242, 0.96);
}

::v-deep(.dp-custom-theme) {
  --dp-background-color: rgba(255, 250, 242, 0.96);
  --dp-text-color: #3f3428;
  --dp-hover-color: rgba(255, 224, 186, 0.82);
  --dp-hover-text-color: #3f3428;
  --dp-hover-icon-color: #785738;
  --dp-primary-color: #7c2d12;
  --dp-primary-disabled-color: rgba(124, 45, 18, 0.34);
  --dp-primary-text-color: #fffaf2;
  --dp-secondary-color: rgba(126, 92, 58, 0.34);
  --dp-border-color: rgba(126, 92, 58, 0.18);
  --dp-menu-border-color: rgba(126, 92, 58, 0.18);
  --dp-border-color-hover: rgba(120, 87, 56, 0.36);
  --dp-border-color-focus: rgba(120, 87, 56, 0.36);
  --dp-disabled-color: rgba(255, 239, 213, 0.48);
  --dp-scroll-bar-background: rgba(255, 250, 242, 0.96);
  --dp-scroll-bar-color: rgba(126, 92, 58, 0.30);
  --dp-success-color: #16a34a;
  --dp-success-color-disabled: #a3d9b1;
  --dp-icon-color: rgba(87, 43, 20, 0.82);
  --dp-danger-color: #dc2626;
  --dp-marker-color: #7c2d12;
  --dp-tooltip-color: #fffaf2;
  --dp-disabled-color-text: #9ca3af;
  --dp-highlight-color: rgba(124, 45, 18, 0.12);
  --dp-range-between-dates-background-color: var(--dp-hover-color);
  --dp-range-between-dates-text-color: var(--dp-hover-text-color);
  --dp-range-between-border-color: var(--dp-hover-color);
}

:global(.games-search-page.games-page--dark) .game-search-filter-panel {
  border-color: rgba(226, 232, 240, 0.16);
  background: rgba(30, 41, 59, 0.96);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-filter-action--ghost:hover {
  background: rgba(226, 232, 240, 0.10);
}

:global(.games-search-page.games-page--dark) .game-search-filter-action--primary {
  background: rgba(203, 213, 225, 0.18);
  color: rgba(241, 245, 249, 0.92);
}

:global(.games-search-page.games-page--dark) .game-search-filter-action--primary:hover {
  background: rgba(203, 213, 225, 0.26);
}

:global(.games-search-page.games-page--dark) .game-search-filter-input {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(15, 23, 42, 0.42);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-filter-input:focus {
  border-color: rgba(203, 213, 225, 0.36);
  box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.18);
}

:global(.games-search-page.games-page--dark) .game-search-filter-chip--idle {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(226, 232, 240, 0.055);
  color: rgba(190, 208, 222, 0.76);
}

:global(.games-search-page.games-page--dark) .game-search-filter-chip--idle:hover {
  border-color: rgba(203, 213, 225, 0.34);
  background: rgba(226, 232, 240, 0.12);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-filter-chip--active {
  border-color: rgba(203, 213, 225, 0.24);
  background: rgba(203, 213, 225, 0.18);
  color: rgba(241, 245, 249, 0.92);
}

:global(.games-search-page.games-page--dark) ::v-deep(.dp-custom-theme) {
  --dp-background-color: rgb(30, 41, 59);
  --dp-text-color: rgba(226, 232, 240, 0.88);
  --dp-hover-color: rgba(226, 232, 240, 0.10);
  --dp-hover-text-color: rgba(241, 245, 249, 0.92);
  --dp-hover-icon-color: rgba(203, 213, 225, 0.78);
  --dp-primary-color: rgba(203, 213, 225, 0.24);
  --dp-primary-disabled-color: rgba(203, 213, 225, 0.10);
  --dp-primary-text-color: rgba(241, 245, 249, 0.92);
  --dp-secondary-color: rgba(148, 163, 184, 0.24);
  --dp-border-color: rgba(226, 232, 240, 0.14);
  --dp-menu-border-color: rgba(226, 232, 240, 0.14);
  --dp-border-color-hover: rgba(203, 213, 225, 0.34);
  --dp-border-color-focus: rgba(203, 213, 225, 0.34);
  --dp-disabled-color: rgba(15, 23, 42, 0.42);
  --dp-scroll-bar-background: rgb(30, 41, 59);
  --dp-scroll-bar-color: rgba(148, 163, 184, 0.40);
  --dp-icon-color: rgba(203, 213, 225, 0.78);
  --dp-marker-color: rgba(203, 213, 225, 0.70);
  --dp-tooltip-color: rgb(30, 41, 59);
  --dp-disabled-color-text: rgba(148, 163, 184, 0.56);
  --dp-highlight-color: rgba(203, 213, 225, 0.14);
}
</style>
