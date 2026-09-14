import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { spawn } from 'node:child_process'
import { once } from 'node:events'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { rootDir } from '../perf/shared.mjs'

// Isolated production server with loopback-only API origins; no runtime .env loading.
export async function startInsightsFixtureApp(resolveResponse, runtimeOverrides = {}) {
  assert(existsSync(join(rootDir, '.output/server/index.mjs')), 'Build Nav Web before running the production fixture smoke')
  let upstreamUrl = ''
  const requests = []
  const upstream = createServer(async (request, response) => {
    const url = new URL(request.url, 'http://localhost')
    const path = url.pathname
    requests.push(url)
    if (path.startsWith('/media/') || path.startsWith('/nav/sites/')) {
      // Small local artwork for layout verification only; never shipped as product assets.
      const site = path.includes('site') || path.includes('default')
      response.writeHead(200, { 'Content-Type': 'image/svg+xml' })
      response.end(site
        ? '<svg xmlns="http://www.w3.org/2000/svg" width="72" height="72"><rect width="72" height="72" rx="14" fill="#82654f"/><path d="M16 54V18l20 18 20-18v36" fill="none" stroke="#f4e9dd" stroke-width="5"/></svg>'
        : '<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215" viewBox="0 0 460 215"><rect width="460" height="215" fill="#394b44"/><circle cx="350" cy="62" r="28" fill="#c8b398"/><path d="M0 215V155L110 58l120 157M190 215l125-123 145 110v13" fill="#778978"/><text x="28" y="180" fill="#fff" font-family="sans-serif" font-size="18">GAME · FIXTURE</text></svg>')
      return
    }
    try {
      let body = ''
      for await (const chunk of request) body += chunk.toString()
      const result = await resolveResponse(url, upstreamUrl + '/media', body ? JSON.parse(body) : undefined)
      response.writeHead(result?.status || 200, { 'Content-Type': 'application/json' })
      response.end(JSON.stringify(result?.status ? { code: 0, message: 'Local fixture unavailable' } : { code: 1, data: result?.data }))
    } catch {
      response.writeHead(500)
      response.end('fixture failure')
    }
  })
  await listen(upstream)
  upstreamUrl = `http://127.0.0.1:${upstream.address().port}`
  const reservation = createServer()
  await listen(reservation)
  const port = reservation.address().port
  await new Promise(resolve => reservation.close(resolve))
  const base = `http://127.0.0.1:${port}`
  const environment = {
    ...process.env, NITRO_HOST: '127.0.0.1', NITRO_PORT: String(port),
    NUXT_PUBLIC_NAV_API_BASE: '/api/v1', NUXT_PUBLIC_NAV_V2_API_BASE: '/api/v2',
    NUXT_PUBLIC_GAME_API_BASE: '/api/v1', NUXT_PUBLIC_GAME_V2_API_BASE: '/api/v2',
    NUXT_PUBLIC_ASSET_PRIMARY_BASE: upstreamUrl,
    NUXT_PUBLIC_ASSET_MIRROR_BASE: upstreamUrl,
    ...runtimeOverrides,
  }
  for (const name of ['NAV_API', 'NAV_V2_API', 'GAME_API', 'GAME_V2_API']) environment[`NUXT_${name}_INTERNAL_BASE`] = `${upstreamUrl}/api/${name.includes('V2') ? 'v2' : 'v1'}`
  const preview = spawn(process.execPath, ['.output/server/index.mjs'], { cwd: rootDir, env: environment, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] })
  let logs = ''
  const record = data => { logs = (logs + data.toString()).slice(-8000) }
  preview.stdout.on('data', record)
  preview.stderr.on('data', record)
  const close = async () => {
    if (preview.exitCode === null) { preview.kill(); await once(preview, 'exit') }
    await new Promise(resolve => upstream.close(resolve))
  }
  try { await ready(base + '/web/background/gofurry-pattern.svg', preview) }
  catch (error) { await close(); throw error }
  return { base, upstreamUrl, requests, close, logs: () => logs }
}
async function listen(server) { server.listen(0, '127.0.0.1'); await once(server, 'listening') }
async function ready(url, child) {
  for (let attempt = 0; attempt < 100; attempt++) {
    assert(child.exitCode === null, 'production server exited during startup')
    try { if ((await fetch(url)).ok) return } catch {}
    await delay(200)
  }
  throw new Error('production server did not become ready')
}
