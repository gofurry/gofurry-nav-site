<template>
  <div class="nav-transition-bar">
    <div class="nav-transition-bar__inner mx-auto">
      <div class="nav-transition-bar__time">
        {{ formattedDateTime }}
      </div>

      <div class="nav-transition-bar__weather">
        <div class="nav-transition-bar__weather-spacer"></div>
        <iframe
          allowtransparency="true"
          src="https://i.tianqi.com/index.php?c=code&id=73&icon=1&num=3&color=d1d5dc"
        ></iframe>
      </div>

      <div v-if="saying" class="nav-transition-bar__quote-shell">
        <div class="nav-transition-bar__quote-wrap">
          <div
            ref="quoteTriggerRef"
            class="nav-transition-bar__quote"
            tabindex="0"
            @mouseenter="showAuthorPopover"
            @mouseleave="hideAuthorPopover"
            @focus="showAuthorPopover"
            @blur="hideAuthorPopover"
          >
            {{ quoteDisplay }}
          </div>
        </div>
      </div>
      <div v-else>
        {{ locale === 'zh' ? '你的恩情狼不会忘记' : 'The pack remembers your kindness.' }}
      </div>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="authorPopoverVisible"
      class="nav-transition-bar__author"
      :style="authorPopoverStyle"
    >
      {{ quoteAuthor }}
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavHomeSaying } from '~/services/nav'
import type { SayingModel } from '~/types/nav'

const props = defineProps<{
  initialSaying?: SayingModel | null
}>()

const { locale } = useI18n()

const formattedDateTime = ref('')
const saying = ref<SayingModel | null>(props.initialSaying ?? null)
const quoteTriggerRef = ref<HTMLElement | null>(null)
const authorPopoverVisible = ref(false)
const authorPopoverStyle = ref<Record<string, string>>({
  left: '0px',
  top: '0px',
})

const quoteDisplay = computed(() => `"${saying.value?.content ?? ''}"`)
const quoteAuthor = computed(() => {
  const author = saying.value?.author?.trim()
  if (author) {
    return author
  }

  return locale.value === 'zh' ? '佚名' : 'Unknown'
})

function updateTime() {
  const formatLocale = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  formattedDateTime.value = new Intl.DateTimeFormat(formatLocale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date())
}

let timeTimer: number | null = null
let authorHideTimer: number | null = null

function updateAuthorPopoverPosition() {
  if (!quoteTriggerRef.value) {
    return
  }

  const rect = quoteTriggerRef.value.getBoundingClientRect()
  const popoverWidth = Math.max(96, Math.min(220, quoteAuthor.value.length * 14 + 24))
  const gap = 12
  const safeInset = 12
  const left = Math.max(safeInset, Math.min(rect.right - popoverWidth, window.innerWidth - popoverWidth - safeInset))
  const top = Math.min(rect.bottom + gap, window.innerHeight - 52)

  authorPopoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
  }
}

function showAuthorPopover() {
  if (authorHideTimer) {
    clearTimeout(authorHideTimer)
    authorHideTimer = null
  }
  updateAuthorPopoverPosition()
  authorPopoverVisible.value = true
}

function hideAuthorPopover() {
  authorHideTimer = window.setTimeout(() => {
    authorPopoverVisible.value = false
    authorHideTimer = null
  }, 80)
}

function syncAuthorPopover() {
  if (authorPopoverVisible.value) {
    updateAuthorPopoverPosition()
  }
}

function currentLang() {
  return locale.value === 'en' ? 'en' : 'zh'
}

async function loadSaying() {
  const response = await getNavHomeSaying(currentLang())
  saying.value = response.saying
}

watch(locale, () => {
  updateTime()
  void loadSaying()
})

watch(() => props.initialSaying, (nextSaying) => {
  saying.value = nextSaying ?? null
})

onMounted(async () => {
  updateTime()
  timeTimer = window.setInterval(updateTime, 60 * 1000)
  window.addEventListener('scroll', syncAuthorPopover, { passive: true, capture: true })
  window.addEventListener('resize', syncAuthorPopover)

  if (!saying.value) {
    await loadSaying()
  }
})

onUnmounted(() => {
  if (timeTimer) {
    window.clearInterval(timeTimer)
  }
  if (authorHideTimer) {
    window.clearTimeout(authorHideTimer)
  }
  window.removeEventListener('scroll', syncAuthorPopover, { capture: true })
  window.removeEventListener('resize', syncAuthorPopover)
})
</script>
