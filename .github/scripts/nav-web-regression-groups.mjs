import { readdirSync } from 'node:fs'
import { fileURLToPath, pathToFileURL } from 'node:url'

// Explicit membership: new specs must be reviewed and assigned, never auto-binned.
export const regressionGroups = Object.freeze({
  games: Object.freeze([
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
  ]),
  'sites-other': Object.freeze([
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
  ]),
  'insights-core': Object.freeze([
    'insights-changes.spec.ts',
    'insights-compare.spec.ts',
    'insights-domain.spec.ts',
    'insights-entity.spec.ts',
    'insights-navigation.spec.ts',
  ]),
  'insights-nav': Object.freeze([
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
  ]),
})

export function regressionInventory() {
  const directory = fileURLToPath(new URL('../../apps/cn/web/tests/browser/regression/', import.meta.url))
  return readdirSync(directory, { withFileTypes: true })
    .filter(entry => entry.isFile() && entry.name.endsWith('.spec.ts'))
    .map(entry => entry.name).sort()
}

export function validateRegressionGroups(groups = regressionGroups, inventory = regressionInventory()) {
  const owners = new Map(), errors = [], files = new Set(inventory)
  for (const [group, members] of Object.entries(groups)) {
    if (!members.length) errors.push(`Empty regression group: ${group}`)
    for (const file of members) {
      if (!files.has(file)) errors.push(`Missing regression spec: ${file} (${group})`)
      if (owners.has(file)) errors.push(`Duplicate regression spec: ${file} (${owners.get(file)}, ${group})`)
      owners.set(file, group)
    }
  }
  for (const file of inventory) if (!owners.has(file)) errors.push(`Unclassified regression spec: ${file}`)
  if (errors.length) throw new Error(errors.join('\n'))
}

export function regressionFiles(group) {
  validateRegressionGroups()
  if (!Object.hasOwn(regressionGroups, group)) throw new Error(`Unknown regression group: ${group}`)
  return regressionGroups[group].map(file => `tests/browser/regression/${file}`)
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    if (process.argv.length !== 3) throw new Error('Usage: node nav-web-regression-groups.mjs <games|sites-other|insights-core|insights-nav>')
    console.log(regressionFiles(process.argv[2]).join('\n'))
  } catch (error) {
    console.error(error.message)
    process.exitCode = 1
  }
}
