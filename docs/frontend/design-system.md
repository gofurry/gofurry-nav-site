# Nav Web design system

The [frontend contract](../../contracts/nav-web-frontend.md) owns the rules.
This guide explains how the existing source applies them. P2.1 establishes token
ownership and annotation; it does not redesign the palette, migrate consumers or
introduce token scales. Actual values belong in source, not this guide.

## Choose the owner by meaning

| Meaning | Owner | Current example |
| --- | --- | --- |
| Product/theme semantics shared across domains and primitives | [tokens.less](../../apps/cn/nav-web/app/assets/styles/tokens.less) | `--gf-surface`, `--gf-text-main`, `--gf-focus-ring` |
| Semantics specific to a reusable primitive | Its stylesheet under `styles/components/` | [rating.less](../../apps/cn/nav-web/app/assets/styles/components/rating.less) owns `--gf-rating-empty` and `--gf-rating-fill` |
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

The existing [button primitive](../../apps/cn/nav-web/app/assets/styles/components/button.less)
uses action fill/contrast tokens for its primary variant, while the
[input primitive](../../apps/cn/nav-web/app/assets/styles/components/input.less)
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
removes unused `--gf-bg-grid-line` from both themes.

`--gf-bg-page` remains a legacy fallback consumed by
[static.less](../../apps/cn/nav-web/app/assets/styles/pages/static.less). Its value
and consumer stay intact until the static-page migration. New layout backgrounds
use `--gf-page-background`; do not copy the legacy fallback into new code.

Avoid promoting a raw literal solely to make the style-policy check pass, merging
tokens by literal equality, or moving primitive/domain semantics into the global
file for convenience. The policy approves exact declaration owners and still
measures ordinary raw values inside those files. P2.1's owner migration is
debt-neutral; the P0/P1 rule/file budgets remain unchanged.

Run the existing checks listed in the [Agent entry](../../apps/cn/nav-web/AGENTS.md).
For token-owner work, compare declarations by theme selector and token name before
and after, including inherited dark values. A green debt check alone does not
prove preservation of active visual values.
