import { afterEach, describe, expect, it, vi } from 'vitest'
import { resetCsrf, sendJSON } from './api'

describe('authenticated mutation transport', () => {
  afterEach(() => { resetCsrf(); vi.unstubAllGlobals() })

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
