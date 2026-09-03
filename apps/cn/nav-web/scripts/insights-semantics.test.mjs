#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { formatInsightChangeWhen, insightChangeOrder } from '../app/utils/insightChanges.ts'
import { formatCnyMinorAmount, formatMinorAmount, priceSegmentKey, publicPriceDisplay } from '../app/utils/insightPrices.ts'
import { formatInsightRatio, normalizeInsightSlice } from '../app/utils/insightDimensions.ts'
import { insightCompareReady, parseInsightCompareIDs } from '../app/utils/insightCompare.ts'

const freePoint = {
  date: '2026-08-28', state: 'free', currency: null,
  initial_amount: null, final_amount: null, discount_percent: null,
}
const free = publicPriceDisplay(freePoint)
assert(free.kind === 'free' && free.amount === null, 'free price lost its distinct state')

const pricedZeroPoint = {
  date: '2026-08-29', state: 'priced', currency: 'CNY',
  initial_amount: 5800, final_amount: 0, discount_percent: 100,
}
const pricedZero = publicPriceDisplay(pricedZeroPoint)
assert(pricedZero.kind === 'priced' && pricedZero.amount === 0, 'priced zero was confused with free or unavailable')
assert(formatCnyMinorAmount(pricedZero.amount, 'zh') === '¥0.00', 'priced zero was not formatted as a real CNY price')
assert(formatMinorAmount(599, 'USD', 'en') === '$5.99', 'regional currency formatting replaced or converted USD')
assert(priceSegmentKey(pricedZeroPoint) === 'priced:CNY' && priceSegmentKey(freePoint) === 'free', 'price identity segmentation collapsed free/priced currency semantics')

for (const state of ['unknown', 'unpriced']) {
  const unavailable = publicPriceDisplay({
    date: '2026-08-30', state, currency: null,
    initial_amount: null, final_amount: null, discount_percent: null,
  })
  assert(unavailable.kind === 'unavailable' && unavailable.amount === null, `${state} price became zero`)
}

const dayChange = {
  type: 'site.ipv6.enabled', date: '2026-08-30', occurred_at: null,
  entity: { id: 1, name: 'Fixture' }, detail: null,
}
assert(formatInsightChangeWhen(dayChange, 'zh') === '2026-08-30', 'day-precision event fabricated a time')

const exactChange = { ...dayChange, occurred_at: '2026-08-30T09:30:00Z' }
assert(formatInsightChangeWhen(exactChange, 'en') !== exactChange.date, 'exact event lost its time')
assert(insightChangeOrder(exactChange) > insightChangeOrder(dayChange), 'timeline ordering ignored exact timestamps')

assert(normalizeInsightSlice('country', 'cn') === 'CN', 'country slice did not normalize to its stable code')
assert(normalizeInsightSlice('nsfw', 'sfw') === 'sfw', 'boolean public slice mapping was rejected')
assert(normalizeInsightSlice('public_interest', 'false') === null, 'internal boolean leaked into public URL state')
assert(normalizeInsightSlice('tag', '123') === '123' && normalizeInsightSlice('tag', '0') === null, 'tag slice identity validation failed')
assert(formatInsightRatio(null) === '—', 'zero denominator was rendered as 0%')
assert(formatInsightRatio(0) === '0.0%', 'a real zero metric was rendered as unavailable')

const compareIDs = parseInsightCompareIDs('37,12,37,48')
assert(compareIDs?.join(',') === '37,12,48', 'Compare IDs did not preserve first appearance while deduplicating')
assert(insightCompareReady(compareIDs) && !insightCompareReady([37]), 'Compare builder state did not require 2–4 entities')
assert(parseInsightCompareIDs('1,2,3,4,5') === null, 'Compare accepted more than four entities')
assert(parseInsightCompareIDs('1,bad') === null, 'Compare accepted an invalid entity ID')
assert(parseInsightCompareIDs('9')?.join(',') === '9', 'Compare lost its one-entity preselected builder state')

const zh = JSON.parse(readFileSync(new URL('../i18n/locales/zh.json', import.meta.url), 'utf8'))
const en = JSON.parse(readFileSync(new URL('../i18n/locales/en.json', import.meta.url), 'utf8'))
assert(zh.sidebar.insights === '生态观测' && en.sidebar.insights === 'Ecosystem', 'public Ecosystem naming drifted')
assert(!JSON.stringify(zh).includes('洞察') && !JSON.stringify(en).includes('Insights'), 'retired public product naming remains in localized UI copy')
const insightsPageDirectory = new URL('../app/pages/insights/', import.meta.url)
const insightPageFiles = readdirSync(insightsPageDirectory, { recursive: true }).filter(path => String(path).endsWith('.vue'))
for (const path of insightPageFiles) {
  const source = readFileSync(new URL(String(path).replaceAll('\\', '/'), insightsPageDirectory), 'utf8')
  assert(!source.includes('insights-hero'), `${path} restored the retired large Ecosystem hero`)
  assert(!source.includes('GoFurryGridBackground'), `${path} bypassed the app-level background foundation`)
}
const layoutSource = readFileSync(new URL('../app/layouts/default.vue', import.meta.url), 'utf8')
assert(layoutSource.includes('<PublicPageBackground />'), 'default layout lost the public background owner')
const backgroundSource = readFileSync(new URL('../app/components/common/PublicPageBackground.vue', import.meta.url), 'utf8')
const globalStyles = readFileSync(new URL('../app/assets/css/main.css', import.meta.url), 'utf8')
const shellStyles = readFileSync(new URL('../app/assets/styles/components/shell.less', import.meta.url), 'utf8')
assert(backgroundSource.includes('data-pattern-status="default"') && backgroundSource.includes('mask-image: var(--gf-page-pattern)'), 'default layout lost its mask-based public pattern')
assert(globalStyles.includes("--gf-page-pattern: url('/web/background/gofurry-pattern.svg')") && globalStyles.includes('--gf-page-pattern-size: 160px 160px'), 'default public pattern contract drifted')
assert(!globalStyles.includes('--gf-page-pattern: none') && existsSync(new URL('../public/web/background/gofurry-pattern.svg', import.meta.url)), 'public pattern asset is missing or disabled')
for (const root of ['.nav-home-page', '.nav-content-shell', '.games-page', '.games-search-page', '.gf-static-page', '.lottery-page', '.lottery-activation-page']) {
  assert(shellStyles.includes(root), `${root} can hide the layout-owned public background`)
}
const mobileNavigationSource = readFileSync(new URL('../app/components/common/MobileBottomTabBar.vue', import.meta.url), 'utf8')
assert(mobileNavigationSource.includes("localePath('/insights')") && mobileNavigationSource.includes('isEcosystemActive'), 'mobile navigation lost the Ecosystem destination or active state')
assert(mobileNavigationSource.includes('@phosphor-icons/vue'), 'mobile navigation stopped using the primary system icon family')
const topNavigationSource = readFileSync(new URL('../app/components/NavBar.vue', import.meta.url), 'utf8')
assert(topNavigationSource.includes('@phosphor-icons/vue'), 'top navigation stopped using the primary system icon family')
assert(existsSync(new URL('../app/components/experimental/ambient/GoFurryGridBackground.vue', import.meta.url))
  && existsSync(new URL('../app/components/experimental/ambient/FallingLeavesCanvas.vue', import.meta.url)), 'retired ambient effects were deleted instead of preserved experimentally')
assert(zh.insights.entity.currentPlayers === '最近观测' && en.insights.entity.currentPlayers === 'Latest Observed', 'entity player observation was presented as realtime')
assert(zh.insights.entity.observedLow === 'GoFurry 观测最低价' && en.insights.entity.observedLow === 'GoFurry Observed Low', 'bounded observed-low naming drifted')
for (const forbidden of ['历史最低', 'Historical Low', 'All-time Low', 'Steam Historical Low']) {
  assert(!JSON.stringify([zh.insights, en.insights]).includes(forbidden), `forbidden observed-low wording leaked: ${forbidden}`)
}
for (const forbidden of ['winner', 'score', 'ranking', 'recommendation', '胜出', '评分', '排名', '推荐']) {
  assert(!JSON.stringify([zh.insights.siteCompare, zh.insights.gameCompare, en.insights.siteCompare, en.insights.gameCompare]).toLowerCase().includes(forbidden.toLowerCase()), `judgement wording leaked into Compare: ${forbidden}`)
}

console.log('[insights] public price, regional identity, timeline, dimension, and Compare semantics passed')

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
