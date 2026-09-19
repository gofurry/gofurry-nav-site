# Frontend engineering

- [Contract](../../contracts/nav-web-frontend.md): normative Nav Web rules,
  ownership boundaries and the P0 audit counting definitions.
- [Agent entry](../../apps/cn/nav-web/AGENTS.md): short operational guidance.
- [Design system](design-system.md): token ownership, naming and usage examples;
  actual values remain in the source owners.
- [Style debt manifest](../../apps/cn/nav-web/frontend-style-debt.json): current
  implementation debt by rule/file; it never overrides the contract.
- [#124 phase plan](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423):
  the staged engineering programme; P0 establishes governance without changing UI.

Contracts state requirements. Documentation explains usage. The debt manifest
records historical noncompliance, not examples or permission for new code.

P2.1 formalizes global token ownership and annotation without changing active
visual values. Primitive boundary/directory work remains P2.2, preference/modal
cleanup P2.3, testing guidance P3, and staged style migrations P4 onward.
The existing source layout and regression harnesses remain in place.

P1 now enforces the contract through ESLint, Stylelint and `style:policy` in the
existing Nav Web CI job. The [Agent entry](../../apps/cn/nav-web/AGENTS.md) lists
the same commands for local validation. Historical ESLint findings use official
bulk suppressions; prune them after fixes. Style policy requires exact rule/file
budgets; use `npm run style:policy:update` only to record reductions. Neither
mechanism permits adding debt. See the contract's P1 maintenance section for
the static-analysis boundaries and fail-closed behavior.
