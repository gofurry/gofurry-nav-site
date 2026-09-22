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
                v-model="draft.content"
                class="game-search-filter-input ml-1 mt-1 w-full px-3 py-2 focus:outline-none"
            />
          </div>
          <div class="grid grid-cols-1 w-[18%]">
            <label class="game-search-filter-label">{{ t("common.pageSize") }}</label>
            <input
                v-model.number="draft.pageSize"
                min="1"
                class="game-search-filter-input mt-1 w-full px-3 py-2 focus:outline-none"
            />
          </div>
        </div>

        <!-- 发行状态 -->
        <div>
          <label class="game-search-filter-label">{{ t("game.search.releaseStatus") }}</label>
          <div class="mt-2 flex flex-wrap gap-2" role="radiogroup" :aria-label="t('game.search.releaseStatus')">
            <button
                v-for="item in availabilityOptions"
                :key="item.value"
                type="button"
                role="radio"
                :aria-checked="draft.availability === item.value"
                @click="setAvailability(item.value)"
                :class="[
                  'game-search-filter-chip',
                  draft.availability === item.value
                    ? 'game-search-filter-chip--active'
                    : 'game-search-filter-chip--idle'
                ]"
            >
              {{ t(item.label) }}
            </button>
          </div>
        </div>

        <!-- 首次可用时间 -->
        <div v-if="draft.availability === 'available'">
          <label class="game-search-filter-label">{{ t("game.search.firstAvailableTime") }}</label>
          <div class="game-search-date-range mt-1">
            <VueDatePicker
                v-model="publishStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
            />
            <VueDatePicker
                v-model="publishEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
            />
          </div>
        </div>

        <!-- 预计发售时间 -->
        <div v-else>
          <label class="game-search-filter-label">{{ t("game.search.plannedReleaseTime") }}</label>
          <div class="game-search-date-range mt-1">
            <VueDatePicker
                v-model="plannedStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
            />
            <VueDatePicker
                v-model="plannedEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
            />
          </div>
        </div>

        <!-- 更新时间 -->
        <div>
          <label class="game-search-filter-label">{{ t("game.search.updateTime") }}</label>
          <div class="game-search-date-range mt-1">
            <VueDatePicker
                v-model="updateStart"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
            />
            <VueDatePicker
                v-model="updateEnd"
                :enable-time-picker="true"
                format="yyyy-MM-dd HH:mm:ss"
                :teleport="false"
                class="game-date-picker dp-custom-theme"
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
              {{ t(item.key === 'latestInfo' && draft.availability === 'upcoming' ? 'game.search.plannedReleaseOrder' : item.label) }}
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
import { buildGameTagGroups, type GameTagGroup } from '@/utils/gameTagDomain'
import type { GameSearchAvailability, GameTagCategory, SearchPageQueryRequest } from '@/types/game'
import { formatLocalDateTime } from '@/utils/util'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  tagGroups: GameTagCategory[]
  query: SearchPageQueryRequest
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'search', query: SearchPageQueryRequest): void
}>()

// The parent owns committed criteria. This instance's draft is discarded on Cancel.
const draft = reactive<SearchPageQueryRequest>({ ...props.query, tag_list: [...(props.query.tag_list ?? [])] })

// =============== 时间 ===============
const parseOptionalDate = (value?: string) => value ? new Date(value.replace(' ', 'T')) : null

const publishStart = ref<Date | null>(parseOptionalDate(draft.pub_start_time))
const publishEnd = ref<Date | null>(parseOptionalDate(draft.pub_end_time))
const plannedStart = ref<Date | null>(parseOptionalDate(draft.planned_start_time))
const plannedEnd = ref<Date | null>(parseOptionalDate(draft.planned_end_time))
const updateStart = ref<Date | null>(parseOptionalDate(draft.update_start_time))
const updateEnd = ref<Date | null>(parseOptionalDate(draft.update_end_time))

const availabilityOptions: Array<{ value: GameSearchAvailability, label: string }> = [
  { value: 'available', label: 'game.search.released' },
  { value: 'upcoming', label: 'game.search.upcoming' },
]

const setAvailability = (availability: GameSearchAvailability) => {
  draft.availability = availability
  if (availability === 'available') {
    plannedStart.value = null
    plannedEnd.value = null
    draft.planned_start_time = undefined
    draft.planned_end_time = undefined
    return
  }
  publishStart.value = null
  publishEnd.value = null
  draft.pub_start_time = undefined
  draft.pub_end_time = undefined
}

// =============== 排序 ===============
const sortOptions = computed(() => [
  {
    key: 'highestRating',
    field: 'score' as const,
    label: 'game.search.highestRating',
    selected: draft.score ?? false
  },
  {
    key: 'mostComments',
    field: 'remark_order' as const,
    label: 'game.search.mostComments',
    selected: draft.remark_order ?? false
  },
  {
    key: 'latestInfo',
    field: 'time_order' as const,
    label: 'game.search.latestInfo',
    selected: draft.time_order ?? false
  }
])

const toggleSort = (key: string) => {
  const item = sortOptions.value.find(i => i.key === key)
  if (!item) return

  draft[item.field] = !item.selected
}

// =============== 分类 & 标签 ===============
const categoryGroups = ref<GameTagGroup[]>([])

const buildCategoryGroups = () => {
  categoryGroups.value = buildGameTagGroups(props.tagGroups, draft.tag_list ?? [], categoryGroups.value)
}

const toggleTag = (tag: GameTagGroup['children'][number]) => {
  const id = Number(tag.id)
  const selected = draft.tag_list ?? []
  draft.tag_list = selected.includes(id) ? selected.filter(value => value !== id) : [...selected, id]
}

// =============== watch & 生命周期 ===============
onMounted(buildCategoryGroups)

watch([() => props.tagGroups, () => draft.tag_list], buildCategoryGroups, { deep: true })

const formatDateTime = formatLocalDateTime

const formatOptionalDateTime = (value: Date | null) => value ? formatDateTime(value) : undefined

watch([publishStart, publishEnd], () => {
  draft.pub_start_time = formatOptionalDateTime(publishStart.value)
  draft.pub_end_time = formatOptionalDateTime(publishEnd.value)
})

watch([plannedStart, plannedEnd], () => {
  draft.planned_start_time = formatOptionalDateTime(plannedStart.value)
  draft.planned_end_time = formatOptionalDateTime(plannedEnd.value)
})

watch([updateStart, updateEnd], () => {
  draft.update_start_time = formatOptionalDateTime(updateStart.value)
  draft.update_end_time = formatOptionalDateTime(updateEnd.value)
})

const onSearch = () => {
  // Include cleared optional fields so applying a snapshot cannot retain old dates.
  emit('search', {
    ...draft,
    pub_start_time: draft.pub_start_time,
    pub_end_time: draft.pub_end_time,
    planned_start_time: draft.planned_start_time,
    planned_end_time: draft.planned_end_time,
    update_start_time: draft.update_start_time,
    update_end_time: draft.update_end_time,
    tag_list: [...(draft.tag_list ?? [])],
  })
  emit('close')
}
</script>
