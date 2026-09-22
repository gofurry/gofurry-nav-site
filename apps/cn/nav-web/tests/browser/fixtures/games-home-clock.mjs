// Loaded only by the isolated Games Home fixture's Nitro child process.
// Keep real timers/performance time: only the product's relative-date input is fixed.
const now = Date.parse(process.env.GOFURRY_TEST_GAMES_HOME_NOW ?? '')
if (!Number.isFinite(now)) throw new Error('Games Home fixture requires a valid fixed date')
Date.now = () => now
