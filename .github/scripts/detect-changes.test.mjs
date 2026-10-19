import assert from 'node:assert/strict'
import { test } from 'node:test'
import { detectChanges } from './detect-changes.mjs'

test('doctor and standalone policy checks do not start application/database suites', () => {
  for (const path of ['tools/doctor.mjs', 'tools/check-production-policy/taskfile.go', 'tools/check-gofmt/main.go']) {
    const gates = detectChanges([path])
    assert.equal(gates.policy, true)
    for (const [name, value] of Object.entries(gates)) {
      if (name === 'modules') assert.deepEqual(value, [])
      else if (name !== 'policy') assert.equal(value, false, `${path}: ${name}`)
    }
  }
})

test('frontend documentation does not start builds or browser suites', () => {
  for (const path of ['apps/cn/web/README.md', 'apps/cn/web/AGENTS.md', 'apps/cn/web/docs/roadmap.md', 'apps/cn/admin/README_zh.md']) {
    const gates = detectChanges([path])
    assert.equal(gates.nav_web, false)
    assert.deepEqual(gates.modules, [])
    assert.equal(gates.game_integration, false)
    assert.equal(gates.policy, true)
  }
})

test('Nav Web code, locks, tests and Docker select the daily Fast frontend gate', () => {
  for (const path of ['app/pages/index.vue', 'pnpm-lock.yaml', 'tests/browser/smoke/example.spec.ts', 'Dockerfile', 'app/content/example.md']) {
    const gates = detectChanges([`apps/cn/web/${path}`])
    assert.equal(gates.nav_web, true, path)
    assert.equal(gates.policy, true)
    assert.deepEqual(gates.modules, [])
  }
})

test('Admin frontend builds its embed owner without database integrations', () => {
  const gates = detectChanges(['apps/cn/admin/react/src/main.tsx'])
  assert.deepEqual(gates.modules, ['apps/cn/admin'])
  assert.equal(gates.game_integration || gates.nav_integration || gates.admin_integration, false)
  assert.deepEqual(detectChanges(['apps/cn/web/app/data/platform-icons.json']).modules, ['apps/cn/admin'])
})

test('workspace rename delete/add paths retain frontend and shared-icon consumers', () => {
  const gates = detectChanges(['apps/cn/nav-web/app/data/platform-icons.json', 'apps/cn/web/app/data/platform-icons.json'])
  assert.equal(gates.nav_web, true)
  assert.equal(gates.policy, true)
  assert.deepEqual(gates.modules, ['apps/cn/admin'])
  assert.equal(gates.game_integration || gates.nav_integration || gates.admin_integration, false)
})

test('database ownership includes Admin and the correct consumers', () => {
  for (const domain of ['game', 'nav', 'admin']) {
    const gates = detectChanges([`db/${domain}/migrations/deleted.sql`])
    const expected = domain === 'admin' ? ['apps/cn/admin'] : [`apps/cn/${domain}-collector`, `apps/cn/${domain}-backend`, 'apps/cn/admin']
    assert.deepEqual(gates.modules, expected)
    assert.equal(gates.foundation && gates.admin_integration, true)
    assert.equal(gates.game_integration, domain === 'game')
    assert.equal(gates.nav_integration, domain === 'nav')
    assert.equal(gates.nav_web, false)
  }
  const admin = detectChanges(['apps/cn/admin/internal/bootstrap/server.go'])
  assert.equal(admin.game_integration && admin.nav_integration && admin.admin_integration, true)
})

test('Task and CI changes retain broad build coverage; SQL tooling remains conservative', () => {
  for (const path of ['Taskfile.yml', '.github/workflows/checks.yml', '.github/workflows/nav-web-full.yml', '.github/workflows/nav-web-visual.yml', '.github/workflows/future.yaml', '.github/scripts/detect-changes.mjs', '.github/scripts/nav-web-regression-groups.mjs']) {
    const gates = detectChanges([path])
    assert.equal(gates.modules.length, 5)
    assert.equal(gates.policy && gates.nav_web, true)
  }
  for (const path of ['sqlc.yaml', 'tools/check-sqlc/main.go', 'tools/db-baseline/main.go', 'tools/go.mod', 'tools/future-tool/main.go']) {
    const gates = detectChanges([path])
    assert.equal(gates.modules.length, 5)
    assert.equal(gates.foundation && gates.game_integration && gates.nav_integration && gates.admin_integration, true)
  }
})

test('archived and placeholder trees never select production gates', () => {
  const gates = detectChanges(['legacy/nav/package.json', 'legacy/gf-uptime/go.mod', 'legacy/gf-uptime/main.go', 'experimental/monitor/main.go', 'third-party/lib/main.go', 'apps/intl/README.md'])
  for (const [name, value] of Object.entries(gates)) {
    if (name === 'modules') assert.deepEqual(value, [])
    else assert.equal(value, false, name)
  }
})

test('root version declarations select their consumers without database suites', () => {
  const go = detectChanges(['.go-version'])
  assert.equal(go.modules.length, 5)
  assert.equal(go.policy && go.vulnerability, true)
  assert.equal(go.nav_web, false)
  const node = detectChanges(['.node-version'])
  assert.deepEqual(node.modules, ['apps/cn/admin'])
  assert.equal(node.policy && node.nav_web, true)
  assert.equal(node.vulnerability, false)
  for (const gates of [go, node]) {
    assert.equal(gates.foundation || gates.game_integration || gates.nav_integration || gates.admin_integration, false)
  }
})
