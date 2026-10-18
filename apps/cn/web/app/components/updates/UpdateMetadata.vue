<script setup lang="ts">
import { PhArrowUpRight } from '@phosphor-icons/vue'
import { updateCommit } from '~/utils/updatePresentation'
import { formatUpdatesFullDate } from '~/utils/updatesDate'

const props = defineProps<{ publishedAt: string; commitSha: string | null }>()
const { locale } = useI18n()
const commit = computed(() => updateCommit(props.commitSha))
</script>

<template>
  <div class="updates-metadata flex flex-wrap items-center gap-x-5 gap-y-2">
    <time :datetime="publishedAt">{{ formatUpdatesFullDate(publishedAt, locale === 'en' ? 'en-US' : 'zh-CN') }}</time>
    <a v-if="commit" :href="commit.href" class="updates-commit inline-flex items-center gap-1" target="_blank" rel="noopener noreferrer" data-update-commit>
      <span>{{ commit.label }}</span><PhArrowUpRight aria-hidden="true" />
    </a>
  </div>
</template>
