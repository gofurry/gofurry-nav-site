<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/20 px-4">
    <div
        class="game-search-filter-panel w-full max-w-2xl overflow-hidden rounded-2xl p-6 shadow"
        :class="{ 'game-search-filter-panel--dark': isDarkTheme }"
    >
      <div class="space-y-2 overflow-y-auto scrollbar-hide max-h-[calc(80vh-3rem)]">

        <!-- 标题 & 操作 -->
        <div class="flex items-center justify-between">
          <h2 class="game-search-filter-title">{{ t("game.search.advancedFilter") }}</h2>
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
            <label class="game-search-filter-label">{{ t("common.keyword") }}</label>
            <input
                v-model="props.query.content"
                class="game-search-filter-input ml-1 mt-1 w-full rounded-lg px-3 py-2 focus:outline-none"
            />
          </div>
          <div class="grid grid-cols-1 w-[18%]">
            <label class="game-search-filter-label">{{ t("common.pageSize") }}</label>
            <input
                v-model.number="props.query.pageSize"
                min="1"
                class="game-search-filter-input mt-1 w-full rounded-lg px-3 py-2 focus:outline-none"
            />
          </div>
        </div>

        <!-- 发布时间 -->
        <div>
          <label class="game-search-filter-label">{{ t("game.search.publishTime") }}</label>
          <div class="flex gap-2 mt-1">
            <VueDatePicker
                v-model="publishStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :dark="isDarkTheme"
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="publishEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :dark="isDarkTheme"
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
          </div>
        </div>

        <!-- 更新时间 -->
        <div>
          <label class="game-search-filter-label">{{ t("game.search.updateTime") }}</label>
          <div class="flex gap-2 mt-1">
            <VueDatePicker
                v-model="updateStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :dark="isDarkTheme"
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="updateEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :dark="isDarkTheme"
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
          </div>
        </div>

        <!-- 排序 -->
        <div>
          <label class="game-search-filter-label">{{ t("common.sort") }}</label>
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
          <label class="game-search-filter-label">{{ t("common.tag") }}</label>
          <div class="mt-2 space-y-2">
            <div v-for="group in categoryGroups" :key="group.id">
              <div class="game-search-filter-group-title">
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
                  class="game-search-filter-expand"
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
import { computed, ref, reactive, watch, onMounted } from 'vue'
import type { GameTagRecord, SearchPageQueryRequest } from '@/types/game'
import { formatLocalDateTime } from '@/utils/util'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { i18n } from '@/main'
import { useThemeStore } from '@/stores/theme'

const { t } = i18n.global
const themeStore = useThemeStore()
const isDarkTheme = computed(() => themeStore.theme === 'dark')

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
  backdrop-filter: blur(18px);
}

.game-search-filter-panel--dark {
  border-color: rgba(226, 232, 240, 0.18);
  background: rgba(15, 23, 42, 0.92);
  color: rgba(226, 232, 240, 0.90);
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.38);
}

.game-search-filter-action {
  cursor: pointer;
  border-radius: 0.65rem;
  padding: 0.38rem 0.75rem;
  font-size: 0.86rem;
  transition: background-color 180ms ease, color 180ms ease, border-color 180ms ease;
}

.game-search-filter-title {
  color: rgba(45, 35, 28, 0.94);
  font-size: 1.08rem;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.25;
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

.game-search-filter-label {
  display: inline-flex;
  color: rgba(87, 83, 78, 0.72);
  font-size: 0.75rem;
  font-weight: 650;
  line-height: 1.25;
}

.game-search-filter-group-title {
  margin-bottom: 0.25rem;
  color: rgba(45, 35, 28, 0.86);
  font-size: 0.88rem;
  font-weight: 750;
  line-height: 1.35;
}

.game-search-filter-expand {
  margin-top: 0.25rem;
  display: inline-flex;
  cursor: pointer;
  user-select: none;
  color: rgba(87, 83, 78, 0.70);
  font-size: 0.75rem;
  font-weight: 650;
  transition: color 180ms ease;
}

.game-search-filter-expand:hover {
  color: rgba(45, 35, 28, 0.92);
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
  font-weight: 550;
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
  font-weight: 550;
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

:deep(.game-date-picker .dp__input_wrap) {
  border-radius: 0.78rem;
}

:deep(.game-date-picker .dp__input) {
  min-height: 2.35rem;
  border-radius: 0.78rem;
  border-color: rgba(126, 92, 58, 0.16);
  background: rgba(255, 247, 236, 0.78);
  color: rgba(45, 35, 28, 0.90);
  font-size: 0.86rem;
  font-weight: 650;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.62);
}

:deep(.game-date-picker .dp__input:focus),
:deep(.game-date-picker.dp__main:focus-within .dp__input) {
  border-color: rgba(120, 87, 56, 0.36);
  box-shadow: 0 0 0 2px rgba(120, 87, 56, 0.10), inset 0 1px 0 rgba(255, 255, 255, 0.70);
}

:deep(.game-date-picker .dp__input_icon),
:deep(.game-date-picker .dp__clear_icon) {
  color: rgba(87, 43, 20, 0.76);
}

:deep(.game-date-picker .dp__menu) {
  overflow: hidden;
  border: 1px solid rgba(126, 92, 58, 0.18);
  border-radius: 1rem;
  background: rgba(255, 250, 242, 0.98);
  box-shadow: 0 18px 42px rgba(91, 62, 28, 0.16);
}

:deep(.game-date-picker .dp__calendar_header) {
  color: rgba(45, 35, 28, 0.82);
  font-size: 0.76rem;
  font-weight: 800;
}

:deep(.game-date-picker .dp__month_year_select) {
  border-radius: 0.65rem;
  color: rgba(45, 35, 28, 0.94);
  font-weight: 800;
}

:deep(.game-date-picker .dp__month_year_select:hover),
:deep(.game-date-picker .dp__inner_nav:hover) {
  background: rgba(255, 224, 186, 0.58);
}

:deep(.game-date-picker .dp__cell_inner) {
  border-radius: 0.62rem;
  color: rgba(45, 35, 28, 0.88);
  font-size: 0.82rem;
  font-weight: 650;
}

:deep(.game-date-picker .dp__cell_inner:hover) {
  background: rgba(255, 224, 186, 0.62);
}

:deep(.game-date-picker .dp__cell_offset) {
  color: rgba(120, 113, 108, 0.44);
}

:deep(.game-date-picker .dp__today) {
  border-color: rgba(120, 87, 56, 0.30);
}

:deep(.game-date-picker .dp__active_date) {
  background: rgba(124, 45, 18, 0.88);
  color: rgba(255, 250, 242, 0.98);
}

:deep(.game-date-picker .dp__time_display),
:deep(.game-date-picker .dp__time_input) {
  color: rgba(45, 35, 28, 0.88);
  font-weight: 700;
}

:deep(.game-date-picker .dp__action_row) {
  border-top: 1px solid rgba(126, 92, 58, 0.12);
  padding: 0.55rem 0.75rem;
}

:deep(.game-date-picker .dp__selection_preview) {
  color: rgba(87, 83, 78, 0.72);
  font-size: 0.78rem;
}

:deep(.game-date-picker .dp__action_button) {
  border-radius: 0.62rem;
  font-weight: 750;
}

:deep(.game-date-picker .dp__action_cancel) {
  border-color: rgba(126, 92, 58, 0.18);
  color: rgba(87, 43, 20, 0.82);
}

:deep(.game-date-picker .dp__action_select) {
  background: rgba(124, 45, 18, 0.88);
  color: rgba(255, 250, 242, 0.98);
}

:global(.games-search-page.games-page--dark) .game-search-filter-panel {
  border-color: rgba(226, 232, 240, 0.16);
  background: rgba(15, 23, 42, 0.92);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-filter-title,
.game-search-filter-panel--dark .game-search-filter-title {
  color: rgba(248, 250, 252, 0.96);
}

.game-search-filter-panel--dark .game-search-filter-label {
  color: rgba(203, 213, 225, 0.76);
}

.game-search-filter-panel--dark .game-search-filter-group-title {
  color: rgba(241, 245, 249, 0.92);
}

.game-search-filter-panel--dark .game-search-filter-expand {
  color: rgba(190, 208, 222, 0.76);
}

.game-search-filter-panel--dark .game-search-filter-expand:hover {
  color: rgba(248, 250, 252, 0.94);
}

.game-search-filter-panel--dark .game-search-filter-input {
  border-color: rgba(226, 232, 240, 0.16);
  background: rgba(15, 23, 42, 0.48);
  color: rgba(226, 232, 240, 0.90);
}

.game-search-filter-panel--dark .game-search-filter-input:focus {
  border-color: rgba(203, 213, 225, 0.38);
  box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.18);
}

.game-search-filter-panel--dark .game-search-filter-chip--idle {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(226, 232, 240, 0.06);
  color: rgba(203, 213, 225, 0.80);
  font-weight: 550;
}

.game-search-filter-panel--dark .game-search-filter-chip--idle:hover {
  border-color: rgba(203, 213, 225, 0.34);
  background: rgba(226, 232, 240, 0.12);
  color: rgba(241, 245, 249, 0.92);
}

.game-search-filter-panel--dark .game-search-filter-chip--active {
  border-color: rgba(203, 213, 225, 0.26);
  background: rgba(203, 213, 225, 0.18);
  color: rgba(248, 250, 252, 0.94);
  font-weight: 550;
}

:global(.games-search-page.games-page--dark) .game-search-filter-label {
  color: rgba(203, 213, 225, 0.72);
}

:global(.games-search-page.games-page--dark) .game-search-filter-group-title {
  color: rgba(241, 245, 249, 0.90);
}

:global(.games-search-page.games-page--dark) .game-search-filter-expand {
  color: rgba(190, 208, 222, 0.72);
}

:global(.games-search-page.games-page--dark) .game-search-filter-expand:hover {
  color: rgba(241, 245, 249, 0.94);
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
  font-weight: 550;
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
  font-weight: 550;
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

.game-search-filter-panel--dark :deep(.game-date-picker .dp__input),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__input) {
  border-color: rgba(226, 232, 240, 0.16);
  background: rgba(15, 23, 42, 0.48);
  color: rgba(241, 245, 249, 0.92);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.045);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__input:focus),
.game-search-filter-panel--dark :deep(.game-date-picker.dp__main:focus-within .dp__input),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__input:focus),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker.dp__main:focus-within .dp__input) {
  border-color: rgba(203, 213, 225, 0.40);
  box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__input_icon),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__clear_icon),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__input_icon),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__clear_icon) {
  color: rgba(203, 213, 225, 0.80);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__menu),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__menu) {
  border-color: rgba(226, 232, 240, 0.16);
  background: rgba(15, 23, 42, 0.98);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.36);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__calendar_header),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__calendar_header) {
  color: rgba(203, 213, 225, 0.78);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__month_year_select),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__cell_inner),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__time_display),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__time_input),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__month_year_select),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__cell_inner),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__time_display),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__time_input) {
  color: rgba(226, 232, 240, 0.90);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__month_year_select:hover),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__inner_nav:hover),
.game-search-filter-panel--dark :deep(.game-date-picker .dp__cell_inner:hover),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__month_year_select:hover),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__inner_nav:hover),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__cell_inner:hover) {
  background: rgba(226, 232, 240, 0.11);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__cell_offset),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__cell_offset) {
  color: rgba(148, 163, 184, 0.48);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__today),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__today) {
  border-color: rgba(203, 213, 225, 0.34);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__active_date),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__active_date) {
  background: rgba(203, 213, 225, 0.22);
  color: rgba(248, 250, 252, 0.96);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__action_row),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__action_row) {
  border-top-color: rgba(226, 232, 240, 0.12);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__selection_preview),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__selection_preview) {
  color: rgba(203, 213, 225, 0.72);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__action_cancel),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__action_cancel) {
  border-color: rgba(226, 232, 240, 0.16);
  background: transparent;
  color: rgba(203, 213, 225, 0.84);
}

.game-search-filter-panel--dark :deep(.game-date-picker .dp__action_select),
:global(.games-search-page.games-page--dark) :deep(.game-date-picker .dp__action_select) {
  background: rgba(203, 213, 225, 0.20);
  color: rgba(248, 250, 252, 0.96);
}
</style>
