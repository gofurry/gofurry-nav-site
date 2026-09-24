# Frontend engineering

- [Contract](../../contracts/nav-web-frontend.md): normative ownership, debt and verification rules.
- [Agent entry](../../apps/cn/nav-web/AGENTS.md): concise working instructions.
- [Design system](design-system.md): implemented token/primitive/domain ownership and accepted cascade decisions; literals stay in source.
- [Testing](testing.md): current Vitest, Browser, pinned Visual and Contract Guard commands, isolation and migration mapping.
- [Style debt](../../apps/cn/nav-web/frontend-style-debt.json): exact rule/file budgets; never overrides the contract.
- [#124 closure record](../acceptance/issue-124-frontend-engineering-closure.md): final evidence, remaining owners, runtime follow-up and manual/remote sign-off status.

P4–P6 appearance migration is complete. P7 replaces the old ad hoc browser and
visual-report runners with the existing formal gates. Performance tools, explicit
cloud acceptance and source/tooling Contract Guards retain their separate scope.
Site Detail #109, Insights #108 and preserved ambient effects keep measured debt;
new code in those areas follows the same contract.

The original [#124 plan](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423)
and older Nav Web style-system/roadmap/handoff documents are historical context.
They do not override the executable code, accepted contract or current commands.
Do not restore retired runners or mechanically rearrange directories to match
an early illustrative plan.
