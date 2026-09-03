import { describe, expect, it, vi } from 'vitest'
import { sendJSON } from '../../lib/api'
import { changeOwnPassword, changeOwnUsername, selfPasswordSchema, selfUsernameSchema } from './self-service-dialogs'

vi.mock('../../lib/api', async (importOriginal) => ({ ...await importOriginal<typeof import('../../lib/api')>(), sendJSON: vi.fn() }))

describe('Admin self-service credentials', () => {
  it('uses the shared username rule and current-password endpoint', async () => {
    expect(selfUsernameSchema.safeParse({ username: 'Invalid Name', current_password: 'current' }).success).toBe(false)
    vi.mocked(sendJSON).mockResolvedValue({ initialized: true, authenticated: true })
    await changeOwnUsername({ username: 'new.name', current_password: 'current' })
    expect(sendJSON).toHaveBeenCalledWith('/api/v1/auth/self/username', 'PUT', { username: 'new.name', current_password: 'current' })
  })

  it('rejects mismatched confirmation and never sends it to the password API', async () => {
    expect(selfPasswordSchema.safeParse({ current_password: 'current', new_password: 'next', confirm_password: 'different' }).success).toBe(false)
    vi.mocked(sendJSON).mockResolvedValue(undefined)
    await changeOwnPassword({ current_password: 'current', new_password: 'next', confirm_password: 'next' })
    expect(sendJSON).toHaveBeenCalledWith('/api/v1/auth/self/password', 'POST', { current_password: 'current', new_password: 'next' })
  })
})
