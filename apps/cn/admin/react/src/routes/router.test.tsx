import { matchRoutes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { router } from './router'

describe('content workspace routing', () => {
  it.each([
    ['/nav/sites', 'nav/sites'], ['/nav/sites/42', 'nav/sites/:id'],
    ['/game/games', 'game/games'], ['/game/games/17', 'game/games/:id'],
  ])('matches %s to the dedicated route', (pathname, expected) => {
    const matches = matchRoutes(router.routes, pathname)
    expect(matches?.at(-1)?.route.path).toBe(expected)
  })
})
