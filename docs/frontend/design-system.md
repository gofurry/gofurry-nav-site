# Nav Web design system

The [frontend contract](../../contracts/nav-web-frontend.md) owns the rules.
This guide explains how the existing source applies them. P2.1 established token
ownership and annotation; P2.2 gives domain-neutral primitives an explicit owner;
P2.3 separates Modal from Preferences and removes its dead cascade. These phases
preserve rendered appearance and introduce no token scales. Actual values
belong in source, not this guide.

## Visual layers

The [Less entry](../../apps/cn/nav-web/app/assets/styles/index.less) composes
`tokens → mixins → primitives → components → pages`:

| Layer | Responsibility |
| --- | --- |
| `tokens.less` | Shared global foundation semantics |
| `mixins.less` | Reusable Less behavior consuming tokens |
| `primitives/` | Domain-neutral button, card, chip, input, modal, pagination and rating building blocks |
| `components/` | Compound product UI: preferences, shell, navigation and footer |
| `pages/` | Existing page/domain composition and its historical styles |

A primitive has stable appearance semantics and can be composed across domains
without depending on a page root or business tokens. A compound component combines
building blocks into product UI. Navigation's `--gf-nav-*` and Footer's
`--gf-footer-*` therefore remain with their compound owners; reuse across routes
does not make them primitives or global tokens.

[modal.less](../../apps/cn/nav-web/app/assets/styles/primitives/modal.less) owns
generic `.gf-modal*` appearance for both Preferences and NSFW confirmation.
[preferences.less](../../apps/cn/nav-web/app/assets/styles/components/preferences.less)
composes Modal/Input/Button/Chip and shared tabs, sources and carousel controls.
`components/` is an intentional layer, not a directory awaiting wholesale renaming.
`domains/` remains a later page-migration target; do not create it empty.

## Choose the owner by meaning

| Meaning | Owner | Current example |
| --- | --- | --- |
| Product/theme semantics shared across domains and primitives | [tokens.less](../../apps/cn/nav-web/app/assets/styles/tokens.less) | `--gf-surface`, `--gf-text-main`, `--gf-focus-ring` |
| Semantics specific to a reusable primitive | Its stylesheet under `styles/primitives/` | [rating.less](../../apps/cn/nav-web/app/assets/styles/primitives/rating.less) owns `--gf-rating-empty` and `--gf-rating-fill` |
| State semantics of a product composition | Its exact compound stylesheet/root | `preferences.less` owns local `--gf-preferences-*` input/toggle states |
| Business/domain semantics | Its existing page/domain stylesheet | [games.less](../../apps/cn/nav-web/app/assets/styles/pages/games.less) owns `--games-*` |

Component-private geometry, such as a cover ratio or a title line count, can stay
in scoped styles. It does not justify a separate local palette, elevation or
typography theme. A reusable Rating needs its own filled/empty meaning; those
tokens need not become global just because multiple pages use the primitive.
Domain tokens normally alias global meaning. Independent domain values need a
business rationale; historical values are not automatically approved examples.

The layout-owned canvas defaults (`--gf-page-background` and
`--gf-page-pattern*`) now live in `tokens.less`. Their consumer,
[PublicPageBackground.vue](../../apps/cn/nav-web/app/components/common/PublicPageBackground.vue),
keeps its existing behavior. [main.css](../../apps/cn/nav-web/app/assets/css/main.css)
owns Tailwind bootstrap, browser reset and base helpers, not global theme tokens.

## Preserve semantic identity

Equal values do not mean equal tokens. `--gf-accent` describes emphasis;
`--gf-accent-fill` describes action fills. `--gf-surface` describes a general
content surface; `--gf-input-bg` describes a form surface. These pairs already
diverge in dark mode, so replacing one with the other would erase meaning.

The existing [button primitive](../../apps/cn/nav-web/app/assets/styles/primitives/button.less)
uses action fill/contrast tokens for its primary variant, while the
[input primitive](../../apps/cn/nav-web/app/assets/styles/primitives/input.less)
uses form background/focus tokens. Choose a token for the role it serves, not
because its current color happens to match a screenshot.

New global names use `--gf-<semantic-role>[-<variant-or-state>]`, as in
`--gf-surface-hover`, `--gf-border-strong` and `--gf-accent-fill-hover`.
Literal names such as `--gf-orange-500` or vague property-first names such as
`--gf-bg-foo` do not explain product meaning. Historical names are not a reason
to rename every active token in this phase.

## Annotate meaning and theme behavior

`tokens.less` keeps root values in semantic groups and mirrors their order in
`html.dark`. Root group comments explain meaning, scope and why the shared role
exists. A short dark-theme contract explains that meaning stays stable and omitted
tokens inherit the root value; section labels need not repeat full explanations.
Individual comments are useful for exceptions, rather than describing syntax or
naming a color. A domain exception should explain why global meaning does not fit.

[mixins.less](../../apps/cn/nav-web/app/assets/styles/mixins.less) consumes this
foundation through `.gf-surface`, `.gf-focus-ring`, `.gf-interactive-motion` and
`.gf-hover-surface`. Mixins express reusable behavior; their existence does not
create another source of token values. This phase leaves them unchanged.

## Add, retain or remove a token

Before adding a token, search the global, primitive and domain owners. Identify
the real consumer and stable semantic role, then choose the smallest correct
owner. Add a semantic comment and account for light/dark meaning. Promote repeated
real needs when migration demonstrates them; do not prebuild typography, spacing,
control-height, z-index or container scales.

Search for consumers before removing a token. Confirmed unused tokens should be
removed unless an explicit external compatibility contract needs them. P2.1
removed unused `--gf-bg-grid-line` from both themes.

`--gf-bg-page` remains a legacy fallback consumed by
[static.less](../../apps/cn/nav-web/app/assets/styles/pages/static.less). Its value
and consumer stay intact until the static-page migration. New layout backgrounds
use `--gf-page-background`; do not copy the legacy fallback into new code.

Avoid promoting a raw literal solely to make the style-policy check pass, merging
tokens by literal equality, or moving primitive/domain semantics into the global
file for convenience. The policy approves exact declaration owners and still
measures ordinary raw values inside those files. P2.1's owner migration is
debt-neutral; the P0/P1 rule/file budgets remain unchanged.

P2.2 moves six zero-debt primitive files without changing their contents. Rating's
exact policy owner follows its new path, while its tokens stay with Rating. A
future debt-bearing file move needs reviewed policy migration: the old path would
have a stale baseline and the new path would have zero budget. Do not raise
budgets or use a structural move to change selectors, values or behavior.

Run the existing checks listed in the [Agent entry](../../apps/cn/nav-web/AGENTS.md).
For token-owner work, compare declarations by theme selector and token name before
and after, including inherited dark values. A green debt check alone does not
prove preservation of active visual values.

## Modal and Preferences ownership example

Preferences' quick-access switch has one product-specific consumer. Its
`preferences-toggle` / `preferences-toggle--on` API belongs to that compound,
not Generic Modal or a speculative Toggle primitive. P2.3 removed the overridden
generic toggle cascade instead of creating tokens for dead styles.

The compound's local tokens describe input idle/focus border, surface and shadow,
and toggle idle/active surface, text, thumb and thumb shadow. Light values alias
global semantics where they already did; dark values preserve the existing
Preferences treatment. Only `.gf-preferences-modal` and
`html.dark .gf-preferences-modal` in `components/preferences.less` may declare
these tokens. Ordinary raw properties there still count as debt.

Shared composition belongs in `preferences.less`; paging/scroll geometry and
Background/Hero/ResourceRoute private structure remain in their Vue scoped styles.
The header override moved from the Modal SFC into the compound owner. Toggle
events, ARIA and Save/Cancel behavior remain with the unchanged Vue logic.

P2.3 retires 29 raw-color occurrences: 11 overridden values are deleted and 18
effective values keep their meaning in 14 local semantic tokens. The safe updater
removes only the old modal budget (924 → 895), granting neither new file debt.
Use the existing Hero/preferences, resource-routing and game-detail smoke scripts
for themed responsive screenshots and effective input/toggle/NSFW modal checks.
