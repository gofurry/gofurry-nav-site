import { isValidElement } from 'react'
import { matchRoutes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { DATAOPS_READ_CAPABILITY } from '../lib/capabilities'
import { router } from './router'

describe('content workspace routing', () => {
  it.each([
    ['/nav/sites', 'nav/sites'], ['/nav/sites/42', 'nav/sites/:id'],
    ['/game/games', 'game/games'], ['/game/games/17', 'game/games/:id'],
  ])('matches %s to the dedicated route', (pathname, expected) => {
    const matches = matchRoutes(router.routes, pathname)
    expect(matches?.at(-1)?.route.path).toBe(expected)
  })

  it.each([
    ['/collection', 'collection'], ['/metrics', 'metrics'], ['/changes', 'changes'],
    ['/system/data-operations', 'system/data-operations'], ['/system/audit', 'system/audit'], ['/system/accounts', 'system/accounts'],
  ])('matches operational route %s natively', (pathname, expected) => {
    const matches = matchRoutes(router.routes, pathname)
    expect(matches?.at(-1)?.route.path).toBe(expected)
  })

  it('guards Data Operations with the canonical backend capability', () => {
    const matches = matchRoutes(router.routes, '/system/data-operations')
    const guard = matches
      ?.map((match) => match.route.element)
      .find((element) => isValidElement<{ capability?: string }>(element) && element.props.capability)
    const capability = isValidElement<{ capability?: string }>(guard) ? guard.props.capability : undefined

    expect(capability).toBe(DATAOPS_READ_CAPABILITY)
    expect(capability).not.toBe('data_ops.read')
  })
})
