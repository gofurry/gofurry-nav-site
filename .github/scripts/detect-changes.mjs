import { appendFileSync, readFileSync } from 'node:fs'
import { pathToFileURL } from 'node:url'

const modules = ['game-collector', 'game-backend', 'nav-collector', 'nav-backend', 'admin', 'uptime']
const documentation = path => /(^|\/)(README(?:_[\w-]+)?\.md|AGENTS\.md)$/i.test(path)
  || /^(docs\/|contracts\/|\.agents\/|CHANGELOG\.md$)/.test(path)
  || /^apps\/cn\/[^/]+\/docs\//.test(path)

// Keep unknown tooling conservative; known standalone checks do not affect SQL/runtime.
export function detectChanges(paths) {
  const changed = paths.filter(path => !documentation(path))
  const has = pattern => changed.some(path => pattern.test(path))
  const ci = has(/^\.github\/(scripts\/|workflows\/(checks|security)\.yml$)/)
  const runtimeTools = changed.some(path => path.startsWith('tools/')
    && !/^tools\/(doctor\.mjs$|check-gofmt\/|check-production-policy\/)/.test(path))
  const sql = ci || runtimeTools || has(/^sqlc\.yaml$/)
  const shared = sql || has(/^Taskfile\.yml$/)
  const adminBackend = changed.some(path => path.startsWith('apps/cn/admin/') && !path.startsWith('apps/cn/admin/react/'))
  const selected = modules.filter(module => shared
    || has(new RegExp(`^apps/cn/${module}/`))
    || (module.startsWith('game-') && has(/^db\/game\//))
    || (module.startsWith('nav-') && has(/^db\/nav\//))
    || (module === 'admin' && has(/^(db\/(admin|game|nav)\/|apps\/cn\/nav-web\/app\/data\/platform-icons\.json$)/)))
    .map(module => `apps/cn/${module}`)
  return {
    modules: selected,
    has_modules: selected.length > 0,
    policy: paths.some(path => /^(Taskfile\.yml$|apps\/cn\/|db\/|tools\/|sqlc\.yaml$|AGENTS\.md$|\.agents\/|contracts\/|\.github\/(scripts\/|workflows\/(checks|security)\.yml$))/.test(path)),
    foundation: sql || has(/^(db\/|apps\/cn\/[^/]+\/internal\/db\/.*\/queries\/)/),
    nav_web: ci || has(/^(Taskfile\.yml$|apps\/cn\/nav-web\/)/),
    game_integration: sql || adminBackend || has(/^(apps\/cn\/game-(backend|collector)\/|db\/game\/)/),
    nav_integration: sql || adminBackend || has(/^(apps\/cn\/nav-(backend|collector)\/|db\/nav\/)/),
    admin_integration: sql || adminBackend || has(/^db\/(admin|game|nav)\//),
    vulnerability: ci || has(/^(apps\/cn\/(admin|game-backend|game-collector|nav-backend|nav-collector|uptime)\/go\.(mod|sum)$|tools\/go\.(mod|sum)$)/),
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  // git -z preserves whitespace/newlines in filenames, including deleted paths.
  const result = detectChanges(readFileSync(0, 'utf8').split('\0').filter(Boolean))
  console.log(JSON.stringify(result, null, 2))
  if (process.env.GITHUB_OUTPUT) {
    appendFileSync(process.env.GITHUB_OUTPUT, Object.entries(result)
      .map(([key, value]) => `${key}=${JSON.stringify(value)}\n`).join(''))
  }
}
