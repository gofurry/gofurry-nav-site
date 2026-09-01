#!/usr/bin/env node
import { formatInsightChangeWhen, insightChangeOrder } from '../app/utils/insightChanges.ts'
import { formatCnyMinorAmount, publicPriceDisplay } from '../app/utils/insightPrices.ts'
import { formatInsightRatio, normalizeInsightSlice } from '../app/utils/insightDimensions.ts'

const free = publicPriceDisplay({
  date: '2026-08-28', state: 'free', currency: null,
  initial_amount: null, final_amount: null, discount_percent: null,
})
assert(free.kind === 'free' && free.amount === null, 'free price lost its distinct state')

const pricedZero = publicPriceDisplay({
  date: '2026-08-29', state: 'priced', currency: 'CNY',
  initial_amount: 5800, final_amount: 0, discount_percent: 100,
})
assert(pricedZero.kind === 'priced' && pricedZero.amount === 0, 'priced zero was confused with free or unavailable')
assert(formatCnyMinorAmount(pricedZero.amount, 'zh') === '¥0.00', 'priced zero was not formatted as a real CNY price')

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

console.log('[insights] public price, timeline, and dimension semantics passed')

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
