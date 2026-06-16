<template>
  <div class="game-detail-info space-y-6 text-sm">

    <section class="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-3">
      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.infoCollectedTime") }}:</span>
        <span>{{ game?.create_time || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.infoUpdatedTime") }}:</span>
        <span>{{ game?.update_time || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.releaseDate") }}:</span>
        <span>{{ game?.release_date || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.supportedPlatforms") }}:</span>
        <span>{{ formattedPlatform || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.supportedLanguages") }}:</span>
        <span>{{ game?.supported_languages || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.ageRestriction") }}:</span>
        <span>{{ game?.required_age || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.gameType") }}:</span>
        <span>{{ formattedType || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="game-detail-info-label w-28 shrink-0">{{ t("game.detail.freeToPlay") }}:</span>
        <span>{{ game ? (game.is_free ? t("common.yes") : t("common.no")) : t("game.panel.none") }}</span>
      </div>
    </section>

    <section v-if="game?.short_description" class="space-y-2">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.shortDescription") }}</h4>
      <p class="leading-relaxed break-words">{{ game.short_description }}</p>
    </section>

    <section class="space-y-3">
      <div>
        <h4 class="game-detail-subtitle mb-1 font-bold">{{ t("game.detail.developer") }}</h4>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="(developer, index) in game?.developers || []"
            :key="`developer-${index}`"
            class="game-detail-chip px-2 py-0.5 text-xs"
          >
            {{ developer }}
          </span>
          <span v-if="!game?.developers?.length" class="game-detail-empty text-sm">
            {{ t("game.panel.none") }}
          </span>
        </div>
      </div>

      <div>
        <h4 class="game-detail-subtitle mb-1 font-bold">{{ t("game.detail.publisher") }}</h4>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="(publisher, index) in game?.publishers || []"
            :key="`publisher-${index}`"
            class="game-detail-chip px-2 py-0.5 text-xs"
          >
            {{ publisher }}
          </span>
          <span v-if="!game?.publishers?.length" class="game-detail-empty text-sm">
            {{ t("game.panel.none") }}
          </span>
        </div>
      </div>
    </section>

    <section v-if="priceList.length" class="space-y-2">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.priceInfo") }}</h4>
      <div class="flex gap-x-1 sm:grid-cols-3 gap-2">
        <div
          v-for="(price, index) in priceList"
          :key="`price-${index}`"
          class="game-detail-price-chip flex items-center justify-center px-3 py-1"
        >
          <span class="font-medium">
            <strong>{{ countryMap[price.country] || price.country }}</strong>
            {{ price.price }}
          </span>
        </div>
      </div>
    </section>

    <section v-if="supportEntries.length" class="space-y-3">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.supportInfo") }}</h4>
      <div class="space-y-2">
        <div
          v-for="entry in supportEntries"
          :key="entry.key"
          class="flex flex-col gap-1 md:flex-row md:gap-2"
        >
          <span class="game-detail-info-label w-28 shrink-0">{{ entry.label }}:</span>
          <a
            v-if="entry.href"
            :href="entry.href"
            target="_blank"
            rel="noopener noreferrer"
            class="game-detail-link break-all"
          >
            {{ entry.value }}
          </a>
          <span v-else class="break-all">{{ entry.value }}</span>
        </div>
      </div>
    </section>

    <section v-if="game?.website" class="space-y-1">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.officialWebsite") }}</h4>
      <div class="game-detail-link break-all">
        <a
          :href="game.website"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ game.website }}
        </a>
      </div>
    </section>

    <section
      v-for="section in requirementSections"
      :key="section.key"
      class="space-y-4"
    >
      <h4 class="game-detail-subtitle font-bold">{{ section.title }}</h4>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="game-detail-requirement-card p-4 space-y-2">
          <div class="game-detail-info-label text-xs">{{ t("game.detail.minimum") }}</div>
          <div
            v-html="section.requirement.minimum || t('game.panel.none')"
            class="leading-relaxed"
          />
        </div>

        <div class="game-detail-requirement-card p-4 space-y-2">
          <div class="game-detail-info-label text-xs">{{ t("game.detail.recommended") }}</div>
          <div
            v-html="section.requirement.recommended || t('game.panel.none')"
            class="leading-relaxed"
          />
        </div>
      </div>
    </section>

    <section v-if="contentDescriptorItems.length" class="space-y-3">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.contentDescriptors") }}</h4>
      <div class="space-y-2">
        <div
          v-for="item in contentDescriptorItems"
          :key="item.key"
          class="flex flex-col gap-1 md:flex-row md:gap-2"
        >
          <span
            v-if="item.label"
            class="game-detail-info-label w-28 shrink-0"
          >
            {{ item.label }}:
          </span>
          <span class="break-words leading-relaxed">{{ item.value }}</span>
        </div>
      </div>
    </section>

    <section v-if="ratingItems.length" class="space-y-3">
      <h4 class="game-detail-subtitle font-bold">{{ t("game.detail.ratings") }}</h4>
      <div class="space-y-2">
        <div
          v-for="item in ratingItems"
          :key="item.key"
          class="flex flex-col gap-1 md:flex-row md:gap-2"
        >
          <span
            v-if="item.label"
            class="game-detail-info-label w-28 shrink-0"
          >
            {{ item.label }}:
          </span>
          <span class="break-words leading-relaxed">{{ item.value }}</span>
        </div>
      </div>
    </section>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { GameBaseInfoResponse, RequirementsModel } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  game: GameBaseInfoResponse | null
}>()

interface SupportEntry {
  key: string
  label: string
  value: string
  href?: string
}

interface RequirementSection {
  key: string
  title: string
  requirement: RequirementsModel
}

interface FlatMetaItem {
  key: string
  label: string
  value: string
}

const countryMap = computed<Record<string, string>>(() => ({
  US: t("game.detail.globalRegion"),
  CN: t("game.detail.chinaRegion"),
  HK: t("game.detail.hongKongRegion")
}))

const formattedPlatform = computed(() => (props.game?.platform ?? '').split(',').filter(Boolean).join(', '))

const formattedType = computed(() => formatType(props.game?.type ?? ''))

const priceList = computed(() => (props.game?.price_list ?? []).filter((item) => item.price))

const supportEntries = computed(() => buildSupportEntries(props.game?.support_info ?? {}))

const requirementSections = computed<RequirementSection[]>(() => {
  const requirements = props.game?.requirements
  if (!requirements) {
    return []
  }

  return [
    { key: 'pc', title: t("game.detail.pcRequirements"), requirement: requirements.pc },
    { key: 'mac', title: t("game.detail.macRequirements"), requirement: requirements.mac },
    { key: 'linux', title: t("game.detail.linuxRequirements"), requirement: requirements.linux },
  ].filter((section) => hasRequirementContent(section.requirement))
})

const contentDescriptorItems = computed(() => flattenExtraValue(props.game?.content_descriptors))

const ratingItems = computed(() => flattenExtraValue(props.game?.ratings))

function hasRequirementContent(requirement?: RequirementsModel | null) {
  return Boolean(requirement && (requirement.minimum || requirement.recommended))
}

function buildSupportEntries(supportInfo: Record<string, string>) {
  return Object.entries(supportInfo).flatMap(([key, rawValue]) => {
    const value = String(rawValue ?? '').trim()
    if (!value) {
      return []
    }

    const lowerKey = key.toLowerCase()
    let href = ''
    if (lowerKey.includes('email')) {
      href = `mailto:${value}`
    } else if (isHttpUrl(value)) {
      href = value
    }

    return [{
      key,
      label: formatSupportLabel(key),
      value,
      href,
    }]
  })
}

function flattenExtraValue(value: unknown, prefix = ''): FlatMetaItem[] {
  const primitive = primitiveToText(value)
  if (primitive) {
    return [{
      key: `${prefix || 'value'}:${primitive}`,
      label: formatMetaLabel(prefix),
      value: primitive,
    }]
  }

  if (Array.isArray(value)) {
    const primitiveValues = value
      .map((item) => primitiveToText(item))
      .filter((item): item is string => Boolean(item))

    const items: FlatMetaItem[] = []
    if (primitiveValues.length) {
      items.push({
        key: `${prefix || 'value'}:list:${primitiveValues.join('|')}`,
        label: formatMetaLabel(prefix),
        value: primitiveValues.join(' / '),
      })
    }

    for (const item of value) {
      if (primitiveToText(item)) {
        continue
      }
      items.push(...flattenExtraValue(item, prefix))
    }

    return items
  }

  if (isRecord(value)) {
    return Object.entries(value).flatMap(([key, item]) => {
      const nextPrefix = prefix ? `${prefix}.${key}` : key
      return flattenExtraValue(item, nextPrefix)
    })
  }

  return []
}

function primitiveToText(value: unknown) {
  switch (typeof value) {
    case 'string':
      return value.trim()
    case 'number':
    case 'boolean':
      return String(value)
    default:
      return ''
  }
}

function formatType(value: string) {
  const normalized = value.trim()
  if (!normalized) {
    return ''
  }
  return normalized
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

function formatSupportLabel(key: string) {
  switch (key.toLowerCase()) {
    case 'url':
      return t("game.detail.supportUrl")
    case 'email':
      return t("game.detail.supportEmail")
    default:
      return formatMetaLabel(key) || key
  }
}

function formatMetaLabel(path: string) {
  if (!path) {
    return ''
  }

  return path
    .split('.')
    .filter(Boolean)
    .map((segment) => {
      if (/^\d+$/.test(segment)) {
        return `#${segment}`
      }
      return segment
        .replace(/[_-]+/g, ' ')
        .replace(/\b\w/g, (char) => char.toUpperCase())
    })
    .join(' / ')
}

function isHttpUrl(value: string) {
  return /^https?:\/\//i.test(value)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
</script>
