import { runtimeTest } from './insights-runtime'

// Static pages have no business API. Reuse the production transport/diagnostics,
// not the Visual fixture's hidden floating controls or screenshot preparation.
export const test = runtimeTest(() => null, () => false, () => ({ status: 500 }))
export { expect, openRuntime, settleRuntime } from './insights-runtime'
