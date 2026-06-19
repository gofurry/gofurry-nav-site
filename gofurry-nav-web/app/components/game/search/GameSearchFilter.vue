<template>
  <div class="game-search-filter-overlay fixed inset-0 z-50 flex items-center justify-center px-4">
    <div
        class="game-search-filter-panel w-full max-w-2xl overflow-hidden p-6"
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
                class="game-search-filter-input ml-1 mt-1 w-full px-3 py-2 focus:outline-none"
            />
          </div>
          <div class="grid grid-cols-1 w-[18%]">
            <label class="game-search-filter-label">{{ t("common.pageSize") }}</label>
            <input
                v-model.number="props.query.pageSize"
                min="1"
                class="game-search-filter-input mt-1 w-full px-3 py-2 focus:outline-none"
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
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="publishEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
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
                :teleport="false"
                class="game-date-picker dp-custom-theme w-1/2"
            />
            <VueDatePicker
                v-model="updateEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
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
