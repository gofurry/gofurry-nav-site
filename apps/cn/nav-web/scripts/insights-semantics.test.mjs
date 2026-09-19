#!/usr/bin/env node
import { groupInsightChangeDates, formatInsightChangeDate } from '../app/utils/insightChangeTimeline.ts'
import { filterCompareSites, compareSelectedEntities } from '../app/utils/insightComparePicker.ts'
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { registerHooks } from 'node:module'
import { insightDimensionBars } from '../app/utils/insightDomain.ts'
import { formatInsightChangeWhen, insightChangeOrder } from '../app/utils/insightChanges.ts'
import { formatCnyMinorAmount, formatMinorAmount, priceSegmentKey, publicPriceDisplay } from '../app/utils/insightPrices.ts'
import { formatInsightRatio, normalizeInsightSlice } from '../app/utils/insightDimensions.ts'
import { insightCompareReady, parseInsightCompareIDs } from '../app/utils/insightCompare.ts'
import { formatGameInsightAxisDate, gameDetailInsightRanges } from '../app/utils/insightHistoryRanges.ts'
import { steamSharedAssetCandidates } from '../app/utils/steamAssets.ts'
import { insightsPrimaryItems, insightsDomainItems, isInsightsPrimaryActive, isInsightsDomainActive } from '../app/components/insights/navigation/navigation.ts'

// Resolve the same extensionless relative TS imports Nuxt supports, without a test dependency.
const tsImports = registerHooks({ resolve(specifier, context, nextResolve) {
  if (specifier.startsWith('./') && context.parentURL?.endsWith('.ts')) {
    const candidate = new URL(`${specifier}.ts`, context.parentURL)
    if (existsSync(candidate)) return nextResolve(candidate.href, context)
  }
  return nextResolve(specifier, context)
} })
const { overviewActivity, overviewSiteIdentities, overviewGeneratedAt, overviewSignal, formatOverviewDelta, overviewSiteMetricKeys, overviewExploreGroups, overviewChangesPath } = await import('../app/utils/insightOverview.ts')
const { selectGamePulse, pulseDiscountPrice } = await import('../app/utils/insightGamePulse.ts')
tsImports.deregister()

assert(overviewSiteMetricKeys.join(',') === 'tls13,ipv6,security_txt', 'Overview site signal contract changed')
assert(formatOverviewDelta(0.042) === '+4.2%' && formatOverviewDelta(-0.011) === '-1.1%', 'ratio delta lost its percentage-point scale or sign')
assert(formatOverviewDelta(null) === '—' && overviewSignal(null) === null, 'missing signal became zero')
assert(overviewSignal(-1) === 0 && overviewSignal(2) === 1 && overviewSignal(0.42) === 0.42, 'progress escaped its accessible 0..1 range')
const overviewNav = { generated_at: '2026-09-01T10:00:00Z', recent_changes: [
  { date: '2026-09-01', occurred_at: null, entity: { id: 1 } },
  { date: '2026-08-30', occurred_at: null, entity: { id: 1 } },
] }
const overviewGame = { generated_at: '2026-09-01T12:00:00Z', recent_changes: Array.from({ length: 6 }, (_, id) => ({ date: '2026-09-01', occurred_at: `2026-09-01T0${id}:00:00Z`, entity: { id } })) }
assert(overviewGeneratedAt(overviewNav, overviewGame) === '2026-09-01T10:00:00.000Z', 'snapshot time did not conservatively use the earlier response')
assert(overviewGeneratedAt(null, overviewGame) === '2026-09-01T12:00:00.000Z' && overviewGeneratedAt(null, null) === null, 'independent snapshot availability failed')
assert(overviewActivity(overviewNav, overviewGame).length === 5 && overviewActivity(overviewNav, overviewGame)[0].entity.id === 5, 'activity lost its one hero plus four limit or event ordering')
assert(overviewActivity(overviewNav, null).every(item => item.domain === 'site') && overviewSiteIdentities(overviewNav).length === 1, 'one-source activity or site identity deduplication failed')
const usPrice = { region: 'US', available: true, currency: 'USD', final_amount: 599, discount_percent: 50 }
const pulseGame = id => ({ id, online_count: { status: 'success', count: 0 }, prices: [usPrice], price: { ...usPrice, region: 'CN', currency: 'CNY' } })
const selectedPulse = selectGamePulse({ top_online: [pulseGame('1')], highest_discount: [pulseGame('1'), pulseGame('2')], latest_games: [pulseGame('1'), pulseGame('2'), pulseGame('3')] })
assert(selectedPulse.players.id === '1' && selectedPulse.discount.id === '2' && selectedPulse.latest.id === '3', 'Pulse candidate priority or deterministic dedupe changed')
assert(pulseDiscountPrice(pulseGame('1')).currency === 'USD', 'US-ranked discount silently used CN display price')
assert(selectGamePulse(null).players === null && selectGamePulse({ top_online: [{ ...pulseGame('1'), online_count: { status: 'unknown', count: 0 } }] }).players === null, 'missing player observations became zero')

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
assert(gameDetailInsightRanges.join(',') === '30d,90d,180d,1y,3y,5y', 'Game detail history ranges drifted or exposed all')
assert(formatGameInsightAxisDate('2026-09-05', '30d') === '09-05', 'short Game history date label was not compact')
assert(formatGameInsightAxisDate('2026-09-05', '1y') === '2026-09', 'one-year Game history date label lost its month')
assert(formatGameInsightAxisDate('2026-09-05', '5y') === '2026-09', 'long Game history date label lost its month')

const compareIDs = parseInsightCompareIDs('37,12,37,48')
assert(compareIDs?.join(',') === '37,12,48', 'Compare IDs did not preserve first appearance while deduplicating')
assert(insightCompareReady(compareIDs) && !insightCompareReady([37]), 'Compare builder state did not require 2–4 entities')
assert(parseInsightCompareIDs('1,2,3,4,5') === null, 'Compare accepted more than four entities')
assert(parseInsightCompareIDs('1,bad') === null, 'Compare accepted an invalid entity ID')
assert(parseInsightCompareIDs('9')?.join(',') === '9', 'Compare lost its one-entity preselected builder state')

const hashed2xAsset = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/123/digest/library_capsule_2x.jpg?version=1#cover'
const hashed2xURL = new URL(hashed2xAsset)
const assetCandidates = steamSharedAssetCandidates(hashed2xAsset, 'china')
assert(assetCandidates.length > 1, 'Steam shared CDN fallback candidates were not generated')
for (const candidate of assetCandidates) {
  const parsed = new URL(candidate)
  assert(parsed.pathname === hashed2xURL.pathname, `Steam CDN fallback changed hashed asset pathname: ${parsed.pathname}`)
  assert(parsed.search === hashed2xURL.search && parsed.hash === hashed2xURL.hash, 'Steam CDN fallback changed asset query/hash')
}

const zh = JSON.parse(readFileSync(new URL('../i18n/locales/zh.json', import.meta.url), 'utf8'))
const en = JSON.parse(readFileSync(new URL('../i18n/locales/en.json', import.meta.url), 'utf8'))
const primaryPaths = ['/insights', '/insights/sites', '/insights/games', '/insights/changes']
const domainPaths = {
  site: ['/insights/sites', '/insights/sites/certificates', '/insights/sites/compare'],
  game: ['/insights/games', '/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare'],
}
for (const domain of ['site', 'game']) {
  assert(overviewExploreGroups[domain].map(item => item.path).join('|') === domainPaths[domain].join('|'), `Overview ${domain} destinations changed`)
  for (const item of overviewExploreGroups[domain]) {
    for (const messages of [zh, en]) assert(messages.insights.editorial.links[item.key]?.description, `Overview ${item.path} lost its localized description`)
  }
}
assert(overviewChangesPath === '/insights/changes', 'Overview Changes destination changed')
const overviewSource = readFileSync(new URL('../app/pages/insights/index.vue', import.meta.url), 'utf8')
assert(overviewSource.includes("<h1>{{ $t('insights.overview.title') }}</h1>"), 'Overview lost its visible localized H1')
assert(overviewSource.includes('EcosystemNavigation') && !overviewSource.includes('InsightsStats'), 'Overview navigation or typography statistics regressed')
assert(overviewSource.includes('Promise.allSettled') && overviewSource.includes('getGameHomePanel(locale.value)'), 'Overview lost independent sources or reused Panel request')
const sitePage = readFileSync(new URL('../app/pages/insights/sites/index.vue', import.meta.url), 'utf8')
const gamePage = readFileSync(new URL('../app/pages/insights/games/index.vue', import.meta.url), 'utf8')
for (const [page, domain, metrics, dimensions, defaultMetric, defaultDimension] of [
  [sitePage, 'site', ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'], ['country', 'group', 'nsfw', 'public_interest'], 'ipv6', 'country'],
  [gamePage, 'game', ['free', 'windows', 'mac', 'linux'], ['primary_tag', 'tag'], 'free', 'primary_tag'],
]) {
  const keys = name => [...(page.match(new RegExp(`const ${name} = \\[([^\\]]+)\\]`))?.[1] || '').matchAll(/'([^']+)'/g)].map(match => match[1])
  assert(keys(domain === 'site' ? 'navMetrics' : 'gameMetrics').join('|') === metrics.join('|'), `${domain} metric keys changed`)
  assert(keys(domain === 'site' ? 'siteDimensions' : 'gameDimensions').join('|') === dimensions.join('|'), `${domain} dimension keys changed`)
  assert(page.includes(`defaultMetric: '${defaultMetric}'`) && page.includes(`defaultDimension: '${defaultDimension}'`), `${domain} query defaults changed`)
  assert(page.includes(`<EcosystemNavigation context="${domain}"`) && page.includes('useInsightsDomain(') && page.includes('useInsightsDimensions('), `${domain} bypassed navigation or query owners`)
  assert(page.includes('localePath(item.path)'), `${domain} deep links lost locale awareness`)
}
assert(!sitePage.includes('getGameHomePanel') && gamePage.includes('getGameHomePanel(locale.value)') && gamePage.includes('Promise.allSettled'), 'Game Panel source lost its independent Game-only boundary')
const domainHeader = readFileSync(new URL('../app/components/insights/domain/InsightsDomainHeader.vue', import.meta.url), 'utf8')
assert(/<h1>\{\{ \$t\(/.test(domainHeader) && domainHeader.includes('insights.sites.title') && domainHeader.includes('insights.games.title'), 'Domain visible H1 lost existing locale semantics')
for (const file of readdirSync(new URL('../app/components/insights/domain/', import.meta.url)).filter(name => name.endsWith('.vue'))) {
  const source = readFileSync(new URL(`../app/components/insights/domain/${file}`, import.meta.url), 'utf8')
  assert(!/\b(?:fetch|useFetch|useAsyncData)\s*\(|from ['"]@\/services\//.test(source), `${file} added presentation-level API requests`)
}
for (const file of ['InsightMetricTrend', 'InsightSliceTrend']) {
  const source = readFileSync(new URL(`../app/components/insights/domain/${file}.vue`, import.meta.url), 'utf8')
  assert(source.includes("await import('echarts')") && source.includes("renderer: 'canvas'") && source.includes('ResizeObserver'), `${file} lost its lazy canvas lifecycle`)
  assert(source.includes('connectNulls: false') && source.includes("trigger: 'axis'") && source.includes('<= 31'), `${file} changed gap, tooltip or symbol semantics`)
}
const slices = { items: Array.from({ length: 10 }, (_, i) => ({ value: String(i), metric_value: i === 1 ? null : i === 0 ? 0 : 2, population: 10 - i })) }
const siteBars = insightDimensionBars(slices, 'site')
const gameBars = insightDimensionBars(slices, 'game')
assert(siteBars.map(bar => bar.item.value).join(',') === '0,1,2,3,4,5,6,7', 'dimension bars reordered or failed to bound the authoritative items')
assert(siteBars[0].signal === 0 && siteBars[1].signal === null && siteBars[2].signal === 1, 'Site bars confused missing/zero or failed to clamp ratios')
assert(gameBars[0].value === 10 && gameBars[0].maximum === 10 && gameBars[1].value === 9, 'Game bars used metric ratios instead of population')
const mediaSource = readFileSync(new URL('../app/components/insights/entity/InsightEntityMedia.vue', import.meta.url), 'utf8')
assert(mediaSource.includes(':alt="entity.name"') && mediaSource.includes(':aria-label="entity.name"') && mediaSource.includes('@error="onError"'), 'entity media lost accessible identity or error fallback')
assert(mediaSource.includes('useManagedAsset(') && mediaSource.includes("'lazy'"), 'entity media lost managed Site resolution or native lazy loading')
for (const path of ['activity/InsightActivityItem.vue', 'overview/InsightsOverviewSites.vue', 'domain/InsightsGamePulse.vue']) {
  const source = readFileSync(new URL(`../app/components/insights/${path}`, import.meta.url), 'utf8')
  assert(source.includes('localePath('), `${path} lost localized entity links`)
  assert(!/\b(?:fetch|useFetch|useAsyncData|onMounted)\s*\(|from ['"]@\/services\//.test(source), `${path} added per-entity fetching`)
}
assert(!/\b(?:fetch|useFetch|useAsyncData)\s*\(|from ['"]@\/services\//.test(mediaSource), 'media added N+1 fetching')
assert(insightsPrimaryItems.map(item => item.path).join('|') === primaryPaths.join('|'), 'Primary navigation URL contract changed')
for (const domain of ['site', 'game']) {
  assert(insightsDomainItems[domain].map(item => item.path).join('|') === domainPaths[domain].join('|'), `${domain} navigation URL contract changed`)
}
for (const item of [...insightsPrimaryItems, ...insightsDomainItems.site, ...insightsDomainItems.game]) {
  for (const messages of [zh, en]) {
    assert(typeof item.label.split('.').reduce((value, key) => value?.[key], messages) === 'string', `${item.path} lost its localized label`)
  }
}
for (const prefix of ['', '/en']) {
  for (const path of primaryPaths) {
    const active = insightsPrimaryItems.filter(item => isInsightsPrimaryActive(`${prefix}${path}`, item.path))
    assert(active.length === 1 && active[0].path === path, `${prefix}${path} did not activate only its own primary tab`)
  }
  for (const domain of ['site', 'game']) {
    const parent = domainPaths[domain][0]
    for (const path of domainPaths[domain]) {
      const primary = insightsPrimaryItems.filter(item => isInsightsPrimaryActive(`${prefix}${path}`, item.path))
      assert(primary.length === 1 && primary[0].path === parent, `${prefix}${path} lost its parent primary tab`)
      const secondary = insightsDomainItems[domain].filter(item => isInsightsDomainActive(`${prefix}${path}`, item.path))
      assert(secondary.length === 1 && secondary[0].path === path, `${prefix}${path} did not activate only its exact domain tab`)
      assert(isInsightsPrimaryActive(`${prefix}${path}/detail`, parent), `${prefix}${path}/detail lost its parent primary tab`)
      assert(!isInsightsDomainActive(`${prefix}${path}/detail`, path), 'Domain navigation stopped using exact matching')
    }
    assert(!isInsightsPrimaryActive(`${prefix}${parent}-other`, parent), 'Primary navigation matched a sibling URL prefix')
  }
  assert(!isInsightsPrimaryActive(`${prefix}/insights/other`, '/insights'), 'Overview navigation stopped using exact matching')
}
for (const name of ['InsightsPrimaryNav', 'InsightsDomainNav']) {
  const source = readFileSync(new URL(`../app/components/insights/navigation/${name}.vue`, import.meta.url), 'utf8')
  assert(source.includes('useRoute()') && source.includes('useLocalePath()') && source.includes('localePath(item.path)'), `${name} lost reactive localized navigation`)
  assert(source.includes(`is${name.replace('Nav', '')}Active(route.path, item.path)`), `${name} bypassed the shared active route contract`)
  assert(source.includes('<nav') && source.includes(':aria-label=') && source.includes(':aria-current='), `${name} lost accessible navigation semantics`)
}
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
const globalStyles = readFileSync(new URL('../app/assets/styles/tokens.less', import.meta.url), 'utf8')
const shellStyles = readFileSync(new URL('../app/assets/styles/components/shell.less', import.meta.url), 'utf8')
assert(backgroundSource.includes('ref(defaultBackgroundPreference())') && backgroundSource.includes('mask-image: var(--gf-page-pattern)'), 'default layout lost its SSR default mask-based public pattern')
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
const gameTabSource = readFileSync(new URL('../app/components/game/detail/insights/GameTabInsights.vue', import.meta.url), 'utf8')
const playerTrendSource = readFileSync(new URL('../app/components/game/detail/insights/GamePlayerTrend.vue', import.meta.url), 'utf8')
const priceHistorySource = readFileSync(new URL('../app/components/game/detail/insights/GamePriceHistory.vue', import.meta.url), 'utf8')
const gameTimelineSource = readFileSync(new URL('../app/components/game/detail/insights/GameInsightsTimeline.vue', import.meta.url), 'utf8')
const insightsStyles = readFileSync(new URL('../app/assets/styles/pages/insights.less', import.meta.url), 'utf8')
assert(gameTabSource.includes('<GameInsightsOverview') && gameTabSource.includes('<GameInsightsTimeline'), 'Game detail lost its dedicated overview or timeline')
assert(!gameTabSource.includes('gameEyebrow') && !gameTabSource.includes('gameDescription'), 'retired Game ecosystem explanatory chrome returned')
assert(playerTrendSource.includes("dailyPeak") && playerTrendSource.includes("dailyAverage") && playerTrendSource.includes('point.avg'), 'player history lost its peak/average dual-series contract')
assert(priceHistorySource.includes('entries.find(item => item.data?.point)'), 'segmented price tooltip stopped selecting the real point entry')
assert(playerTrendSource.includes("trigger: 'axis'") && priceHistorySource.includes("trigger: 'axis'"), 'long Game history lost axis hover tooltips')
assert(playerTrendSource.includes('showSymbol: props.points.length <= 31') && priceHistorySource.includes('showSymbol: props.points.length <= 31'), 'long Game history restored persistent point symbols')
assert(gameTimelineSource.includes('v-for="(item, index) in orderedItems"') && !gameTimelineSource.includes('.reverse('), 'Game timeline changed DOM order to create its visual path')
assert(gameTimelineSource.includes('class="insights-ranges"') && gameTimelineSource.includes("'insights-ranges__button--active': mode === option"), 'Game timeline must reuse shared Insights selectors')
assert(!insightsStyles.includes('.game-insights-timeline__modes'), 'Game timeline restored duplicate selector styles')
assert(insightsStyles.includes('grid-template-columns: repeat(3, minmax(0, 1fr))') && insightsStyles.includes("[data-connector='left']"), 'Game compact timeline lost its three-column serpentine layout')
assert(existsSync(new URL('../app/components/experimental/ambient/GoFurryGridBackground.vue', import.meta.url))
  && existsSync(new URL('../app/components/experimental/ambient/FallingLeavesCanvas.vue', import.meta.url)), 'retired ambient effects were deleted instead of preserved experimentally')
assert(zh.insights.entity.latestPlayerObservation === '最近一次玩家观测' && en.insights.entity.latestPlayerObservation === 'Latest player observation', 'entity player observation was presented as realtime')
assert(zh.insights.entity.observedLowValue.startsWith('GoFurry 观测低价') && en.insights.entity.observedLowValue.startsWith('GoFurry Observed Low'), 'bounded observed-low naming drifted')
for (const forbidden of ['历史最低', 'Historical Low', 'All-time Low', 'Steam Historical Low']) {
  assert(!JSON.stringify([zh.insights, en.insights]).includes(forbidden), `forbidden observed-low wording leaked: ${forbidden}`)
}
for (const forbidden of ['winner', 'score', 'ranking', 'recommendation', '胜出', '评分', '排名', '推荐']) {
  assert(!JSON.stringify([zh.insights.siteCompare, zh.insights.gameCompare, en.insights.siteCompare, en.insights.gameCompare]).toLowerCase().includes(forbidden.toLowerCase()), `judgement wording leaked into Compare: ${forbidden}`)
}

// B4 preserves data/query owners while changing the workspace presentation.
const workspaceSource = path => readFileSync(new URL('../app/' + path, import.meta.url), 'utf8')
const workspacePages = Object.fromEntries(['players', 'prices', 'languages', 'certificates'].map(name => [name, workspaceSource('pages/insights/' + (name === 'certificates' ? 'sites/' : 'games/') + name + '.vue')]))
for (const [name, source] of Object.entries(workspacePages)) {
  const key = { players: 'playerIntelligence', prices: 'priceIntelligence', languages: 'languageIntelligence', certificates: 'certificateIntelligence' }[name]
  assert(source.includes('InsightsWorkspaceHeader') && source.includes("$t('insights." + key + ".title')"), name + ' lost its visible localized H1')
  assert(source.includes('<EcosystemNavigation context="' + (name === 'certificates' ? 'site' : 'game') + '"'), name + ' lost Domain navigation')
  assert(source.includes('InsightWorkspaceDisclosure'), name + ' lost data disclosure')
  assert(!source.includes('intelligence-panel') && !source.includes('intelligence-stats'), name + ' restored KPI cards')
  assert(!/get(?:GameDetail|SiteDetail|GameInfo|NavSite)|fetch\(/.test(source), name + ' added entity lookups')
}
const workspaceHeader = workspaceSource('components/insights/workspace/InsightsWorkspaceHeader.vue')
assert(workspaceHeader.includes('<h1>{{ title }}</h1>'), 'Workspace header hid its H1')
for (const file of readdirSync(new URL('../app/components/insights/workspace/', import.meta.url))) {
  const source = workspaceSource('components/insights/workspace/' + file)
  assert(!/\b(?:fetch|useFetch|useAsyncData)\s*\(|from ['"]@\/services\//.test(source), file + ' added per-entity API requests')
}
const { players, prices, languages, certificates } = workspacePages
assert(players.includes("['latest_observed', 'peak_30d', 'average_30d']") && players.includes(": 'latest_observed'") && players.includes('query: { metric }') && players.includes('getGamePlayerRanking(selectedMetric.value)'), 'player metric/query contract changed')
assert(prices.includes("['CN', 'US', 'HK']") && prices.includes(": 'CN'") && prices.includes('query: { region }') && prices.includes('formatMinorAmount(value, currency, locale.value)'), 'price region/query or amount semantics changed')
assert(prices.includes('overviewError') && prices.includes('discountError'), 'price sources lost independent failures')
assert(languages.includes('insights.languageIntelligence.overlap') && languages.includes('explicit_full_audio_games') && languages.includes('explicit_full_audio_share') && languages.includes('supported_games') && languages.includes('<table>') && languages.includes('<details'), 'language distribution lost overlapping/full-audio/raw data semantics')
for (const key of ['verified', 'failed', 'known', 'coverage', 'expired', 'expires_within_7d', 'expires_in_8_30d', 'later', 'not_applicable', 'stale', 'not_probed', 'probe_failed', 'unknown']) assert(certificates.includes(key), 'certificate lost ' + key)
for (const [source, path] of [[workspaceSource('components/insights/workspace/InsightRankingList.vue'), '/games/'], [prices, '/games/'], [workspaceSource('components/insights/workspace/InsightRiskList.vue'), '/site/']]) {
  assert(source.includes('InsightEntityMedia') && source.includes("localePath('" + path), 'Workspace identity media or localized entity link lost')
}

// B5 preserves share URLs, evidence groups, and optional identity presentation.
for (const input of ['0,1', '-1,2', '1,bad', '1,2,3,4,5', ['1','2']]) assert(parseInsightCompareIDs(input) === null, 'invalid comparison IDs accepted')
assert(parseInsightCompareIDs('3,1,2,3')?.join(',') === '3,1,2', 'Compare URL order changed')
assert(insightCompareReady([1,2,3,4]) && !insightCompareReady([]), 'Compare readiness boundary changed')
const pickerItems = [{ id: 1, name: 'Some Fox', subtitle: 'a.fox.example' }, { id: 2, name: 'Fox', subtitle: 'b.test' }, { id: 3, name: 'Fox Den', subtitle: 'c.test' }, { id: 4, name: 'Fox Bay', subtitle: 'd.test' }]
assert(filterCompareSites(pickerItems, 'FOX', []).map(item => item.id).join(',') === '2,3,4,1', 'Site filter lost exact/prefix/contains/stable order')
assert(filterCompareSites(pickerItems, 'a.fox.example', [2]).map(item => item.id).join(',') === '1', 'Site domain search failed')
assert(filterCompareSites(pickerItems, '', [1,2]).map(item => item.id).join(',') === '3,4', 'selected Sites returned as addable results')
const resolved = compareSelectedEntities([3,2,9], [{ id: 3, name: 'Authoritative' }], pickerItems)
assert(resolved.map(item => item.name).join('|') === 'Authoritative|Fox|#9', 'selected identity priority/order changed')
for (const [domain, folder, required] of [['site','sites',['capabilities','certificate']], ['game','games',['basic','platform','activity','price','language']]]) {
  const source = workspaceSource('pages/insights/' + folder + '/compare.vue')
  assert(source.includes("robots: 'noindex, follow'") && source.includes('InsightsWorkspaceHeader'), domain + ' Compare SEO/H1 lost')
  assert(source.includes('parseInsightCompareIDs(route.query.ids)') && source.includes('insightCompareReady(selectedIDs.value)'), domain + ' bypassed URL contract')
  assert(source.includes('selectedIDs.value.flatMap') && source.includes('item.' + domain + '.id === id'), domain + ' matrix no longer follows URL order')
  assert(source.includes('InsightComparePicker') && !source.includes('inputmode="numeric"') && !source.includes('applySelection'), domain + ' restored raw ID input UX')
  for (const key of required) assert(source.includes("key: '" + key + "'"), domain + ' lost matrix group ' + key)
  assert(!/get(?:GameDetail|SiteDetail|GameInfo|NavSite)/.test(source), domain + ' added per-entity details')
}
const picker = workspaceSource('components/insights/compare/InsightComparePicker.vue')
assert(picker.includes('role="combobox"') && picker.includes('role="listbox"') && picker.includes('aria-activedescendant') && picker.includes('ArrowDown') && picker.includes('ArrowUp') && picker.includes('Enter') && picker.includes('Escape'), 'picker keyboard semantics missing')
const pickerData = workspaceSource('composables/useInsightComparePicker.ts')
assert(pickerData.includes('getNavSiteDirectory(locale.value)') && pickerData.includes('onMounted(') && pickerData.includes('getSearchSimple(locale.value, value.trim(), { signal: controller.signal })'), 'picker lost client directory or authoritative Game simple search')
assert(pickerData.includes('token !== generation') && pickerData.includes('controller?.abort()') && pickerData.includes('350'), 'Game search lost debounce/abort/stale guard')
const matrix = workspaceSource('components/insights/compare/InsightCompareMatrix.vue')
assert(matrix.includes('InsightEntityMedia') && matrix.includes('localePath(') && matrix.includes('scope="row"') && matrix.includes('scope="col"'), 'matrix lost identity links or table semantics')
const sitemapSource = readFileSync(new URL('../server/routes/sitemap.xml.ts', import.meta.url), 'utf8')
assert(!/['"]\/(?:en\/)?insights\/(?:sites|games)\/compare/.test(sitemapSource), 'Compare entered sitemap inventory')
for (const messages of [zh, en]) {
  const copy = JSON.stringify([messages.insights.siteCompare, messages.insights.gameCompare, messages.insights.comparePicker])
  for (const forbidden of ['winner', 'score', 'ranking', 'recommendation', '胜出', '评分', '排名', '推荐', '领先', '更安全', '性价比更高']) assert(!copy.toLowerCase().includes(forbidden.toLowerCase()), 'Compare judgement wording: ' + forbidden)
  assert(!/\b(?:Best|Better|Live Players)\b/.test(copy), 'Compare introduced evaluative/realtime wording')
}

// B6 keeps source order, date precision and the existing filter/cursor owner.
const timelineItems = [
  { ...exactChange, domain: 'site', entity: { id: 1 }, date: '2026-09-09' },
  { ...dayChange, domain: 'site', entity: { id: 1 }, date: '2026-09-09' },
  { ...dayChange, domain: 'game', entity: { id: 2 }, date: '2026-09-08' },
]
const timelineGroups = groupInsightChangeDates(timelineItems)
assert(timelineGroups.length === 2 && timelineGroups[0].items.length === 2, 'timeline deduplicated repeated entity events')
assert(timelineGroups.flatMap(group => group.items).every((item, i) => item === timelineItems[i]), 'timeline reordered or mutated source events')
assert(groupInsightChangeDates([...timelineItems, { ...timelineItems[2] }])[1].items.length === 2, 'pagination split the same date')
assert(groupInsightChangeDates([...timelineItems, timelineItems[0]]).length === 3, 'timeline moved a non-adjacent event')
assert(groupInsightChangeDates([]).length === 0 && formatInsightChangeDate('bad', 'en') === 'bad', 'timeline empty/invalid date handling changed')
assert(formatInsightChangeDate('2026-09-09', 'en') === 'Sep 9, 2026', 'date grouping shifted its UTC day')
const changesPage = workspaceSource('pages/insights/changes.vue')
assert(changesPage.includes('<EcosystemNavigation />') && changesPage.includes('InsightsWorkspaceHeader') && changesPage.includes("$t('insights.changeExplorer.title')"), 'Changes lost global navigation or visible localized H1')
assert(changesPage.includes("buildInsightsSeo('changes', locale.value)"), 'Changes SEO contract changed')
for (const values of ["['site', 'game']", "['7d', '30d', '90d', 'all']", "['capability', 'target', 'certificate']", "['pricing_model', 'platform', 'release', 'price', 'discount']"]) assert(changesPage.includes(values), 'Changes filter keys changed')
assert(changesPage.includes('sequence === requestSequence') && changesPage.includes('delete query.cursor') && changesPage.includes('items.value = [...items.value, ...page.items]'), 'Changes lost stale-response/opaque pagination contract')
const changeFeed = workspaceSource('components/insights/InsightsChangeExplorerFeed.vue')
assert(changeFeed.includes('InsightEntityMedia') && changeFeed.includes('localePath(entityPath(item))') && changeFeed.includes('formatInsightChangeWhen(item, locale)'), 'Changes lost identity, localized links or timestamp precision')
assert(changeFeed.includes(':aria-busy=') && changeFeed.includes('data-load-more') && changeFeed.includes('role="status"'), 'Changes lost loading/error/pagination accessibility')
assert(!/\b(?:fetch|useFetch|useAsyncData)\s*\(|from ['"]@\/services\//.test(changeFeed), 'timeline added entity requests')
assert(!/get(?:GameDetail|SiteDetail|GameInfo|NavSite)/.test(changesPage), 'Changes added per-entity requests')
for (const messages of [zh,en]) {
  for (const forbidden of ['重大变化', '热门变化', '趋势上涨', '实时', 'realtime', 'Live feed']) assert(!JSON.stringify(messages.insights.changeExplorer).includes(forbidden), 'Changes fabricated event significance/time')
}

console.log('[insights] navigation, public price, regional identity, timeline, dimension, and Compare semantics passed')

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
