# Frontend engineering

- [Contract](../../contracts/nav-web-frontend.md): normative Nav Web rules,
  ownership boundaries and the P0 audit counting definitions.
- [Agent entry](../../apps/cn/nav-web/AGENTS.md): short operational guidance.
- [Design system](design-system.md): token ownership, naming and usage examples;
  actual values remain in the source owners.
- [Testing](testing.md): Vitest Unit/Nuxt, Playwright Game Detail/Resource Routing, isolation and
  retained Contract Guards, style-policy, legacy browser and external acceptance.
- [Style debt manifest](../../apps/cn/nav-web/frontend-style-debt.json): current
  implementation debt by rule/file; it never overrides the contract.
- [#124 phase plan](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423):
  the staged engineering programme; P0 establishes governance without changing UI.

Contracts state requirements. Documentation explains usage. The debt manifest
records historical noncompliance, not examples or permission for new code.

P2.1 formalized global token ownership and annotation. P2.2 separates six shared
visual primitives from compound product styles, preserving their contents.
P2.3 separates Generic Modal from Preferences composition and retires its raw
debt without changing rendered states. P3.1 establishes Vitest Unit/Nuxt and
migrates four legacy suites. P3.2.1 establishes the Playwright Browser Gate and
migrates Game Detail; P3.2.2 adds Resource Routing/Managed/Steam browser regressions.
Hero lifecycle/Preferences, Insights and other domains keep legacy runners. Visual
baselines remain P3.3 and staged style migrations P4 onward. Independent
contract/tooling guards remain in place.

The older [style-system document](../../apps/cn/nav-web/docs/style-system.md) and
[roadmap](../../apps/cn/nav-web/docs/roadmap.md) are historical migration context.
Their completion claims do not describe #124's current status;
the contract, Agent entry and guide above are the current sources.

P1 now enforces the contract through ESLint, Stylelint and `style:policy` in the
existing Nav Web CI job. The [Agent entry](../../apps/cn/nav-web/AGENTS.md) lists
the same commands for local validation. Historical ESLint findings use official
bulk suppressions; prune them after fixes. Style policy requires exact rule/file
budgets; use `npm run style:policy:update` only to record reductions. Neither
mechanism permits adding debt. See the contract's P1 maintenance section for
the static-analysis boundaries and fail-closed behavior.
