import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'
import { regressionGroups, regressionInventory, regressionFiles, validateRegressionGroups } from './nav-web-regression-groups.mjs'

const copyGroups = () => Object.fromEntries(Object.entries(regressionGroups).map(([group, files]) => [group, [...files]]))

test('every current regression spec has exactly one explicit group', () => {
  assert.deepEqual(Object.keys(regressionGroups), ['games', 'sites-other', 'insights-core', 'insights-nav'])
  validateRegressionGroups()
  assert.deepEqual(Object.values(regressionGroups).flat().sort(), regressionInventory())
})

test('C3-R1 moves only the five Insights core owners and preserves other membership and order', () => {
  const core = [
    'insights-changes.spec.ts',
    'insights-compare.spec.ts',
    'insights-domain.spec.ts',
    'insights-entity.spec.ts',
    'insights-navigation.spec.ts',
  ]
  assert.deepEqual(regressionGroups['insights-core'], core)
  for (const file of core) assert.equal(regressionGroups['insights-nav'].includes(file), false)
  assert.deepEqual(regressionGroups['insights-nav'], [
    'background-storage.spec.ts',
    'hero-catalog.spec.ts',
    'hero-handoff.spec.ts',
    'hero-lifecycle.spec.ts',
    'hero-local-background.spec.ts',
    'hero-preference-modes.spec.ts',
    'insights-overview.spec.ts',
    'insights-workspace.spec.ts',
    'managed-asset-routing.spec.ts',
    'nav-home-header.spec.ts',
    'nav-revealed-content.spec.ts',
    'nav-shell.spec.ts',
    'resource-routing-preferences.spec.ts',
  ])
  assert.deepEqual(regressionGroups.games, [
    'game-collections.spec.ts',
    'game-detail-content.spec.ts',
    'game-detail-insights.spec.ts',
    'game-detail-interactions.spec.ts',
    'game-detail.spec.ts',
    'game-review-dialog.spec.ts',
    'games-home-closure.spec.ts',
    'games-home-showcase.spec.ts',
    'games-home.spec.ts',
    'games-search-contract.spec.ts',
    'games-search-interactions.spec.ts',
    'games-search-lifecycle.spec.ts',
    'games-search-states.spec.ts',
    'steam-asset-routing.spec.ts',
  ])
  assert.deepEqual(regressionGroups['sites-other'], [
    'error-experience.spec.ts',
    'lottery.spec.ts',
    'page-scroll-dock.spec.ts',
    'preferences-foundation.spec.ts',
    'seo-recovery.spec.ts',
    'site-detail-contract.spec.ts',
    'site-detail-refinement.spec.ts',
    'site-detail-shell.spec.ts',
    'site-groups.spec.ts',
    'site-insights.spec.ts',
    'site-observation.spec.ts',
    'site-overview.spec.ts',
    'site-recommendations.spec.ts',
    'site-security.spec.ts',
    'updates.spec.ts',
  ])
})

test('an existing spec cannot lose its owner', () => {
  const groups = copyGroups(), omitted = groups.games.pop()
  assert.throws(() => validateRegressionGroups(groups), error => error.message.includes(`Unclassified regression spec: ${omitted}`))
})

test('new specs require deliberate classification', () => {
  assert.throws(() => validateRegressionGroups(regressionGroups, [...regressionInventory(), 'future.spec.ts']), /Unclassified regression spec: future/)
})

test('deleted or renamed listed files fail closed', () => {
  const removed = regressionGroups.games[0]
  assert.throws(() => validateRegressionGroups(regressionGroups, regressionInventory().filter(file => file !== removed)), /Missing regression spec/)
})

test('duplicates within or across groups fail', () => {
  for (const group of ['games', 'sites-other']) {
    const groups = copyGroups()
    groups[group].push(groups.games[0])
    assert.throws(() => validateRegressionGroups(groups), /Duplicate regression spec/)
  }
})

test('empty and unknown groups cannot silently run the whole suite', () => {
  assert.throws(() => validateRegressionGroups({ ...regressionGroups, games: [] }), /Empty regression group/)
  assert.throws(() => regressionFiles('future'), /Unknown regression group/)
})

test('CLI emits only explicit Playwright paths and rejects absent/unknown selectors', () => {
  const script = fileURLToPath(new URL('./nav-web-regression-groups.mjs', import.meta.url))
  for (const group of Object.keys(regressionGroups)) {
    const result = spawnSync(process.execPath, [script, group], { encoding: 'utf8' })
    assert.equal(result.status, 0, result.stderr)
    assert.deepEqual(result.stdout.trim().split(/\r?\n/), regressionFiles(group))
  }
  for (const args of [[], ['future'], ['games', 'sites-other']]) {
    const result = spawnSync(process.execPath, [script, ...args], { encoding: 'utf8' })
    assert.equal(result.status, 1)
    assert.equal(result.stdout, '')
  }
})
