// Read-only checks; installation and shell configuration remain explicit choices.
import { spawnSync } from 'node:child_process'

function version(command, args) {
  const windows = process.platform === 'win32'
  // All commands/arguments below are constants, including pnpm.cmd on older installs.
  const result = spawnSync(windows ? [command, ...args].join(' ') : command, windows ? [] : args, {
    encoding: 'utf8', windowsHide: true, shell: windows,
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
  ['pnpm', version('pnpm', ['--version']), text => text === '12.6.0', '12.6.0'],
  ['Git', version('git', ['--version']), text => /^git version /.test(text), 'available'],
]
for (const [name, actual, accepts, required] of checks) {
  const ok = Boolean(accepts(actual))
  console.log(`${ok ? 'OK' : 'FAIL'} ${name}: ${actual || 'not available'} (required: ${required})`)
  if (!ok) process.exitCode = 1
}
console.log('Docker is optional for local checks; use task doctor:docker before image builds.')
