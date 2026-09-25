// Read-only checks; installation and shell configuration remain explicit choices.
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const repositoryRoot = fileURLToPath(new URL('../', import.meta.url))
const frontends = [
  ['Admin', 'apps/cn/admin/react'],
  ['Nav Web', 'apps/cn/nav-web'],
]

function version(command, args, cwd = repositoryRoot) {
  const windows = process.platform === 'win32'
  // All commands/arguments below are constants, including pnpm.cmd on older installs.
  const result = spawnSync(windows ? [command, ...args].join(' ') : command, windows ? [] : args, {
    encoding: 'utf8', windowsHide: true, shell: windows,
    cwd,
    // Check prepared tools without downloading a manager or changing its defaults.
    env: command === 'pnpm' ? {
      ...process.env,
      COREPACK_ENABLE_NETWORK: '0',
      COREPACK_ENABLE_AUTO_PIN: '0',
      COREPACK_DEFAULT_TO_LATEST: '0',
      npm_config_manage_package_manager_versions: 'false',
    } : process.env,
    timeout: 30_000,
  })
  return result.status === 0 ? result.stdout.trim() : ''
}

const checks = [
  ['Task', version('task', ['--version']), text => {
    const parts = text.match(/(\d+)\.(\d+)\.(\d+)/)?.slice(1).map(Number)
    return parts && (parts[0] > 3 || (parts[0] === 3 && (parts[1] > 45 || (parts[1] === 45 && parts[2] >= 3))))
  }, '>=3.45.3'],
  ['Go', version('go', ['version']), text => /^go version go1\.26\.7\s/.test(text), '1.26.7'],
  ['Node.js', process.version, text => /^v24\./.test(text), '24.x'],
  ...frontends.map(([name, dir]) => [
    `pnpm (${name})`, version('pnpm', ['--version'], new URL(`../${dir}/`, import.meta.url)),
    text => text === '12.6.0', '12.6.0', dir,
  ]),
  ['Git', version('git', ['--version']), text => /^git version /.test(text), 'available'],
]
for (const [name, actual, accepts, required, frontendDir] of checks) {
  const ok = Boolean(accepts(actual))
  console.log(`${ok ? 'OK' : 'FAIL'} ${name}: ${actual || 'not available'} (required: ${required})`)
  if (!ok) process.exitCode = 1
  if (!ok && frontendDir) {
    console.log(`pnpm must resolve to 12.6.0 inside ${frontendDir}; the global default may differ.`)
    console.log(`If using Corepack, run corepack install from ${frontendDir} to prepare the pinned manager. See docs/development.md.`)
  }
}
console.log('Docker is optional for local checks; use task doctor:docker before image builds.')
