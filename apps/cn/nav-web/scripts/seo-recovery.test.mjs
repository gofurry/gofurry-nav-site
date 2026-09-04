#!/usr/bin/env node

import { readFileSync } from 'node:fs'
import { authoritativePageStatus } from '../app/utils/authoritativePageError.ts'
import { parsePositiveEntityRouteId } from '../app/utils/routeIdentity.ts'
import { siteEntityPath, siteTargetPath } from '../app/utils/siteRoutes.ts'
import { buildSiteGroupSeo, MAX_DESCRIPTION_LENGTH } from '../app/utils/seo.ts'
import { parseGameInventory, parseSiteGroupInventory, parseSiteInventory } from '../server/utils/sitemapInventory.ts'

for (const value of ['1', '110', '999999']) {
  assert(parsePositiveEntityRouteId(value) === value, `valid entity ID ${value} was rejected`)
}
for (const value of ['', '0', '-1', 'abc', '1.5', '%7Bfoo%7D']) {
  assert(parsePositiveEntityRouteId(value) === null, `invalid entity ID ${value} was accepted`)
}

assert(siteEntityPath(225) === '/site/225', 'Site entity path changed')
assert(siteTargetPath(225, 'example.com') === '/site/225?domain=example.com', 'Site target path is not a query view')
assert(siteTargetPath(225, '{"domain":[]}') === '/site/225?domain=%7B%22domain%22%3A%5B%5D%7D', 'Site target encoding is unstable')

assert(authoritativePageStatus({ statusCode: 404 }, 'game') === 404, 'HTTP 404 was not preserved')
assert(authoritativePageStatus({ response: { status: 404 } }, 'site') === 404, 'nested fetch HTTP 404 was not preserved')
assert(authoritativePageStatus(new Error('查询站内游戏主档案失败: game not found'), 'game') === 404, 'Game envelope not-found was not classified')
assert(authoritativePageStatus(new Error('站点不存在'), 'site') === 404, 'Site envelope not-found was not classified')
assert(authoritativePageStatus(new Error('target 不属于当前 site'), 'site') === 404, 'Site target not-found was not classified')
for (const error of [
  { statusCode: 500 },
  { status: 502 },
  new Error('fetch failed'),
  new Error('database timeout'),
]) {
  assert(authoritativePageStatus(error, 'game') === 503, 'temporary Game failure did not map to 503')
  assert(authoritativePageStatus(error, 'site') === 503, 'temporary Site failure did not map to 503')
}

for (const locale of ['zh', 'en']) {
  const community = buildSiteGroupSeo({ name: 'Furry Community', id: 10, locale })
  const tools = buildSiteGroupSeo({ name: 'Furry Tools', id: 11, locale })
  assert(community.description !== tools.description, `${locale} Site Group descriptions are not unique`)
  assert(community.description.includes('Furry Community'), `${locale} Site Group description lost its identity`)
  assert(tools.description.includes('Furry Tools'), `${locale} Site Group description lost its identity`)
  for (const seo of [community, tools]) {
    assert(seo.title.length > 0 && seo.description.length > 0, `${locale} Site Group SEO is empty`)
    assert(seo.description.length <= MAX_DESCRIPTION_LENGTH, `${locale} Site Group description exceeds the shared maximum`)
  }
}

assert(parseSiteInventory({ code: 1, data: { items: [{ id: 1, domains: ['example.com'] }] } })[0].id === '1', 'Nav Sites inventory rejected valid data')
assert(parseSiteGroupInventory({ code: 1, data: [{ id: '2' }] })[0].id === '2', 'Site Groups inventory rejected valid data')
assert(parseGameInventory({ code: 1, data: [{ game_id: '3' }] })[0].id === '3', 'Games inventory rejected valid data')

for (const invalid of [
  () => parseSiteInventory({ code: 0, data: { items: [] } }),
  () => parseSiteInventory({ code: 1, data: [] }),
  () => parseSiteGroupInventory({ code: 1, data: {} }),
  () => parseGameInventory({ code: 1, data: [{ name: 'missing id' }] }),
]) {
  assertThrows(invalid, 'invalid sitemap inventory did not fail closed')
}

for (const path of [
  '../scripts/perf/shared.mjs',
  '../docs/performance/budget.json',
  '../docs/performance/baseline.json',
  '../docs/performance/README.md',
]) {
  const source = readFileSync(new URL(path, import.meta.url), 'utf8')
  assert(!source.includes('steam-zone') && !source.includes('/steam'), `${path} retains the retired Steam-page performance scenario`)
}

const nuxtConfig = readFileSync(new URL('../nuxt.config.ts', import.meta.url), 'utf8')
for (const route of ["'/games/prize'", "'/games/prize/**'", "'/en/games/prize'", "'/en/games/prize/**'"]) {
  assert(nuxtConfig.includes(route), `Prize route rule ${route} is missing`)
}
assert(nuxtConfig.match(/'X-Robots-Tag': 'noindex, follow'/g)?.length === 4, 'Prize route rules do not consistently set X-Robots-Tag')

console.log('[seo-recovery] route identity, authoritative status, and sitemap inventory contracts passed')

function assertThrows(action, message) {
  try {
    action()
  } catch {
    return
  }
  throw new Error(message)
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
