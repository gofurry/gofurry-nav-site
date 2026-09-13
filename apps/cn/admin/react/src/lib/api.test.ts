import { afterEach, describe, expect, it, vi } from 'vitest'
import { resetCsrf, sendJSON, sendForm } from './api'

describe('authenticated mutation transport', () => {
  afterEach(() => { resetCsrf(); vi.unstubAllGlobals() })

  it('sends multipart with CSRF and lets the browser supply the boundary', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, data: { token: 'csrf', header_name: 'X-CSRF' } })))
      .mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, data: { primary: 'ready' } })))
    vi.stubGlobal('fetch', fetchMock)
    const body = new FormData(); body.set('file', new Blob(['icon']), 'icon.ico')
    await sendForm('/api/v1/nav/sites/1/icon', body)
    expect(fetchMock.mock.calls[1][1]).toEqual(expect.objectContaining({ credentials: 'include', method: 'POST', headers: { 'X-CSRF': 'csrf' }, body }))
  })

  it('fetches CSRF state and sends the server-provided header', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, message: '', data: { token: 'csrf-token', header_name: 'X-CSRF-Token' } }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, message: '', data: { id: 9 } }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    vi.stubGlobal('fetch', fetchMock)
    await sendJSON('/api/v1/nav/sayings', 'POST', { language: 'zh', saying: '测试' })
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(fetchMock.mock.calls[1]?.[1]).toEqual(expect.objectContaining({ method: 'POST', headers: { 'X-CSRF-Token': 'csrf-token' } }))
  })
})
