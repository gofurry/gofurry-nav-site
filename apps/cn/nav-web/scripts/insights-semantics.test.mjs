#!/usr/bin/env node
import { readFileSync } from 'node:fs'
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
