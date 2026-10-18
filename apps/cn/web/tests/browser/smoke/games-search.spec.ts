import { test, expect } from '../fixtures/games-search'

test('Games Search hydrates and renders one basic search', async ({ search }) => {
  await search.open({ locale: 'en' }) // Shared owner asserts SSR 200, hydration and initial results.
  expect(search.rendered).toContain('games-search-page')
  await search.openFilter()
  await search.keyword.fill('Smoke search')
  await search.apply()
  await search.waitResults('Smoke search')
  await expect(search.page).toHaveURL(/content=Smoke(?:\+|%20)search/)
  search.assertQuiet()
})
