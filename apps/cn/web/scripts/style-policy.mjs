import { checkPolicy, formatReport } from './style-policy/policy.mjs'

try {
  const args = process.argv.slice(2)
  if (args.some(arg => arg !== '--update-baseline') || args.length > 1) throw new Error('Usage: node scripts/style-policy.mjs [--update-baseline]')
  const result = await checkPolicy(process.cwd(), { update: args.includes('--update-baseline') })
  console.log(formatReport(result))
  if (!result.ok) process.exitCode = 1
} catch (error) {
  console.error(`Style policy tooling failure: ${error.stack || error}`)
  process.exitCode = 1
}
