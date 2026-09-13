import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { readFileSync } from 'node:fs'
import { stripTypeScriptTypes } from 'node:module'
import { launchPerfBrowser } from './perf/shared.mjs'

const source = stripTypeScriptTypes(readFileSync(new URL('../app/utils/backgroundPreferences.ts', import.meta.url), 'utf8'))
const moduleURL = `data:text/javascript;base64,${Buffer.from(source).toString('base64')}`
const png = readFileSync(new URL('../public/logo-mini.png', import.meta.url)).toString('base64')
const server = createServer((_req, res) => { res.setHeader('Content-Type', 'text/html'); res.end('<!doctype html><title>Local background acceptance</title>') })
await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve))
const browser = await launchPerfBrowser()
try {
  const page = await browser.newPage()
  await page.goto(`http://127.0.0.1:${server.address().port}`)
  let mutations = 0
  page.on('request', (request) => { if (request.method() !== 'GET') mutations++ })
  const svg = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M1 1h3v3H1z"/></svg>'
  const first = await page.evaluate(async ({ moduleURL, svg, png }) => {
    const bg = await import(moduleURL)
    for (const body of ['<svg onload="alert(1)"/>', '<svg><use href="https://remote.invalid/secret"/></svg>', '<svg><image href="#x"/></svg>']) {
      let rejected = false
      try { await bg.prepareLocalBackground(new File([body], 'bad.svg', { type: 'image/svg+xml' })) } catch { rejected = true }
      if (!rejected) throw new Error('active SVG was accepted')
    }
    const bitmap = Uint8Array.from(atob(png), (char) => char.charCodeAt(0))
    const raster = await bg.prepareLocalBackground(new File([bitmap], 'local.png', { type: 'image/png' }))
    if (raster.kind !== 'raster') throw new Error('raster was treated as a color mask')
    const local = await bg.prepareLocalBackground(new File([svg], 'local.svg', { type: 'image/svg+xml' }))
    // Reactive UI records are proxies; persist a plain record containing the Blob.
    await bg.saveBackgroundPreference({ version: 1, source: 'local', overrides: { color: '#123456', opacity: 0, size_px: 200 } }, new Proxy(local, {}))
    return { raw: localStorage.getItem(bg.BACKGROUND_PREFERENCE_KEY), text: await (await bg.readLocalBackground()).blob.text() }
  }, { moduleURL, svg, png })
  assert.equal(first.text, svg)
  assert.ok(!first.raw.includes('<svg') && !first.raw.includes('blob:') && !first.raw.includes('data:'))
  await page.reload()
  const persisted = await page.evaluate(async (moduleURL) => {
    const bg = await import(moduleURL)
    const item = await bg.readLocalBackground()
    const preference = bg.readBackgroundPreference()
    await bg.saveBackgroundPreference({ version: 1, source: 'server', pattern_id: '12', overrides: { opacity: 0.1 } })
    const serverPreference = localStorage.getItem(bg.BACKGROUND_PREFERENCE_KEY)
    await bg.saveBackgroundPreference(bg.defaultBackgroundPreference(), undefined, true)
    return { text: await item.blob.text(), preference, serverPreference, cleared: await bg.readLocalBackground() }
  }, moduleURL)
  assert.equal(persisted.text, svg)
  assert.equal(persisted.preference.overrides.opacity, 0)
  assert.deepEqual(JSON.parse(persisted.serverPreference), { version: 1, source: 'server', pattern_id: '12', overrides: { opacity: 0.1 } })
  assert.equal(persisted.cleared, null)
  assert.equal(mutations, 0, 'local file operations must not send HTTP uploads')
  console.log('Browser background acceptance passed: SVG safety, raster decoding, IndexedDB persistence/reload/clear, explicit overrides, and zero uploads')
} finally { await browser.close(); await new Promise((resolve) => server.close(resolve)) }
