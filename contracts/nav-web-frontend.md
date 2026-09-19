# Nav Web frontend contract

## Scope and authority

This contract governs `apps/cn/nav-web` under
[#124](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423).
MUST/MUST NOT are requirements; SHOULD permits a reasoned, documented departure;
MAY denotes an allowed choice. The P0 implementation specification narrows the
parent programme to governance and a one-time debt baseline.

For frontend decisions, resolve evidence in this order:

1. Executable code and tests for actual behavior and compatibility.
2. This contract for intended frontend architecture.
3. `apps/cn/nav-web/AGENTS.md` for operational guidance.
4. `docs/frontend/*` for explanation and usage.
5. `frontend-style-debt.json` only for historical migration state.

Contributors MUST report discrepancies. Existing noncompliant code is evidence of
debt, not precedent that overrides the new-code contract. Repository-wide and
[managed asset](assets.md) contracts still apply; this contract does not change
API, SSR, hydration, routing, cookies, local backgrounds or resource fallback.
React Admin follows its separate [frontend contract](admin-frontend.md).

## Styling ownership

**Tailwind owns structure; Less owns appearance.**

Tailwind MUST own structural composition only. Visual appearance MUST be
expressed through semantic classes, Less and design tokens.

| Responsibility | Owner |
| --- | --- |
| Placement, flex/grid, alignment, position, responsive composition, overflow, visibility, page/container sizing, outer spacing | Tailwind |
| Text alignment, truncation and whitespace behavior | Tailwind MAY be used |
| Control height/padding, color, background, border appearance, radius, shadow, ring, typography, opacity, visual hover/focus/motion | Shared primitive or domain Less using semantic tokens |
| Component-private structural geometry not broadly reusable | Scoped style MAY be used |

Allowed examples: `flex`, `grid`, `items-center`, `justify-between`, `relative`,
`absolute`, `fixed`, `overflow-hidden`, `w-full`, `max-w-*`, `px-4`, `gap-4`,
`md:flex`, `lg:grid-cols-3`, `truncate`, `whitespace-nowrap`, `text-center`.
These classes are for composition: `px-4` on a container does not authorize
reconstructing a button's internal padding/height outside its primitive.

New code MUST NOT express appearance through Tailwind utilities such as
`bg-slate-*`, `text-orange-*`, `border-[#…]`, `rounded-xl`, `shadow-*`, `ring-*`,
`font-bold`, `text-sm`, `leading-relaxed`, `tracking-*`, `opacity-70`,
`hover:bg-*`, `focus:ring-*`, `dark:bg-*` or `dark:text-*`.
Responsive/state modifiers do not change the underlying ownership.

New arbitrary visual values (`bg-[#…]`, `text-[#…]`, `border-[#…]`, `shadow-[…]`,
`rounded-[…]`) MUST NOT be introduced. Structural arbitrary values such as
`z-[120]`, `max-w-[2080px]` and `min-h-[calc(…)]` are outside P0's visual counts;
their existing uses MAY remain. Future structural tokens require demonstrated
reuse and a separate scoped change; P0 does not normalize these values.

Typography is visual language: font size, weight, line height, letter spacing
and text color SHOULD migrate from historical Tailwind into semantic/domain
classes. New reusable UI SHOULD consume shared or domain typography rather than
create a local scale. Text alignment/truncation remains structural.

## Current source layout

`nuxt.config.ts` loads `app/assets/css/main.css`, then
`app/assets/styles/index.less`. Contributors MUST retain this working structure
until a separately scoped migration:

- `main.css`: Tailwind bootstrap, reset, base elements/scrollbars and existing
  page-background base variables. This is not an alternate business-style layer.
- `styles/index.less`: style composition/import root.
- `styles/tokens.less`: global semantic token declarations.
- `styles/mixins.less`: reusable Less behavior built on tokens.
- `styles/components/*.less`: current shared visual primitives.
- `styles/pages/*.less` and existing subdirectories: current page/domain styles.
- `app/components/common`, `nav`, `game`, `site`, `insights`: existing Vue owners.

The parent plan's `primitives/`, `domains/` and `tests/` layout is a future target,
not an instruction to create/move directories in P0. Existing raw values and
domain theme islands remain historical debt. Do not copy them into new code.

## Token hierarchy and annotation

| Level | Meaning and ownership | Examples |
| --- | --- | --- |
| Global | Reusable product semantics in `styles/tokens.less`, independent of a business page | `--gf-surface`, `--gf-surface-strong`, `--gf-surface-hover`, `--gf-text-main`, `--gf-text-muted`, `--gf-border`, `--gf-accent`, `--gf-danger`, `--gf-radius-*`, `--gf-shadow-*`, `--gf-motion-*` |
| Domain | Shared visual meaning within a domain, in that domain's existing Less owner | `--games-*`, `--nav-*`, `--updates-*` |
| Component | Local geometry or behavior, not a second theme | Proposed examples: cover ratio or title line count |

Contributors MUST search existing tokens before adding one. Domain tokens SHOULD
alias global semantics by default, for example `--games-border: var(--gf-border)`.
An independent domain value requires a genuine business meaning and a documented
rationale, such as discount emphasis differing from ordinary interaction accent.
Component-local tokens MUST NOT create a separate color, shadow, radius or
typography theme. A token name alone does not make an arbitrary value semantic.

Future feedback, typography, spacing, control-height, layer and container tokens
MAY be added when actual reuse justifies them. Examples in the parent plan are
not a claim that names such as `--gf-success` or `--gf-warning` exist today.

New/changed token groups MUST have concise comments explaining semantic meaning,
scope, why they exist and why they belong at that level. Exceptional tokens MUST
also explain the specific exception. Group comments SHOULD carry shared context;
individual comments are for exceptions, not a description of every CSS literal.
P0 records this requirement without rewriting existing token declarations.

```less
/* Surface: shared content and interaction surfaces across domains.
 * Global ownership keeps light/dark meaning consistent; consumers should
 * reuse these semantics before introducing domain-specific colors.
 */
```

A comment merely saying "red" is not a rationale. A domain exception comment
should explain its business role and why the global meaning does not fit.

## Shared primitives

The existing foundation includes `.gf-button`, `.gf-card`, `.gf-input`,
`.gf-chip`, `.gf-modal`, `.gf-pagination` and `.gf-rating` in `styles/components/`.
Contributors MUST inspect and reuse the appropriate primitive and its variants
before creating a new selector. New consumers MUST NOT redefine its core
appearance through local overrides; extend the owning primitive deliberately
when a reusable variant is needed.

Appearance reuse belongs in a **CSS primitive**. Reused behavior plus
accessibility (state, ARIA, keyboard navigation and focus management) belongs
in a **Vue primitive**: tabs, dialogs, selects, carousels, popovers or segmented
controls when such reuse is demonstrated. Contributors MUST NOT add one-line
wrapper components without behavior/accessibility value.

When a pattern appears in two different business domains, contributors SHOULD
evaluate promotion into the shared foundation. Domain-specific meaning MAY stay
in its domain. Contributors MUST NOT create speculative fields/tabs/notices or
other primitives merely because they appear in the future programme.

## Scoped styles, themes and raw values

`<style scoped>` MAY own component-private structure. It MUST NOT be a place to
invent a new local theme or replace existing primitive appearance. Repeated
appearance SHOULD move upward as reuse warrants it: scoped → domain → primitive.

Theme semantics MUST use CSS variables/semantic tokens, preserving meaning
between light and dark. Contributors MUST reuse the existing `html.dark` theme
entry. They MUST NOT create independent page-dark systems or Tailwind dark
appearance when existing semantics suffice. Legacy theme islands are debt.

Raw colors (`#hex`, `rgb()`, `rgba()`), literal shadows/radii and visual durations
SHOULD live in approved token declarations, not ordinary component/page
selectors, inline styles or script-generated appearance. The same ownership
applies to typography scales. Approved locations are global declarations,
justified domain declarations, and precise documented exceptions. The narrow
P0 color count below does not authorize uncatalogued raw typography or geometry
used as control appearance.

## Debt and exceptions

Historical debt MAY remain during staged migration. New debt MUST NOT increase;
after removal, the baseline MUST decrease in the same change. The manifest is
current state, not a design example or a pool of credits for new violations.

`frontend-style-debt.json` uses `schema_version: 1`, a `baseline` object containing
exactly the six P0 rule families below, and an `exceptions` array. Each baseline
maps Nav Web-relative POSIX paths to positive integer occurrence counts. Omitted
file/rule pairs have budget **0**, including new files. Do not store line numbers.
Contributors MUST NOT offset one file's increase with another file's decrease,
inflate counts to pass checks, or evade measurement by moving/renaming debt.

An exception intentionally retains a specific pattern temporarily; baseline
debt merely records historical noncompliance. Every exception MUST specify all
of `path`, `rule`, `issue`, `reason`, `remove_when`, using an exact path and a
specific rule. Reasons MUST explain why the existing pattern is necessary and
the removal condition MUST be actionable. Unexplained ignore lists and permanent
directory wildcards are prohibited. Existing visual-guard allowlists MUST NOT
be silently promoted to exceptions. P0 grants no exceptions automatically.

#108 Insights and #109 Site Detail are **migration exclusions**, not style
exceptions. P4–P6 MUST NOT proactively migrate those surfaces. Their historical
debt MUST still be measured, and new code there MUST follow this contract.
Any precise exception tied to those redesigns MUST be removed when its stated
replacement lands; issue membership is not a blanket exemption.

`shared-primitive-override` is a normative concern above, but has no numeric P0
baseline: reliable low-false-positive detection belongs to later review.

## P0 audit definition

The initial inventory audits the unchanged application tree at
`f5576f6bfe1981a7c1b3a2d58f0287bf13d2a15a` on 2026-09-19. This section freezes
the one-time counting semantics so P1 can reproduce them; no policy scanner is
installed by P0. Paths below are relative to `apps/cn/nav-web`.

### Source scope and occurrence unit

Inspect Git-tracked `app/**/*.vue`, `*.ts`, `*.js`, `*.css` and `*.less`.
Exclude `tests`, `__tests__`, `fixtures`, `generated` directories, TypeScript
`*.test.ts`, `*.spec.ts`, `*.d.ts`, and the encoded geographic data asset
`app/assets/js/china.js`. The P0 snapshot has 245 in-scope files. Static SVG,
JSON, raster/icon assets, tooling/config, `server/`, dependencies and generated
Nuxt/output directories are not style source for this inventory.
`app/components/experimental/ambient/` IS in scope; it is not the repository's
excluded root `experimental/` tree. #108/#109 receive no scope exclusion.

Count authored source occurrences, not unique spellings, matched lines,
compiled selectors or runtime repetitions. Two copies of a class count twice;
one class in a loop counts once. Ignore comments and non-style content such as
validation messages, slot names and selectors merely mentioning utility classes.
Parse Vue template/style/script regions and CSS/Less declarations separately.

For Tailwind, inspect static `class` and `*-class` attributes (including transition
class props), static choices in bound class expressions, script literals that
actually flow into those bindings, and `@apply` lists. Each branch is authored
source and counts separately. The P0 script-dataflow review found class literals
in `Footer.vue` (`hoverClass`), `site/SiteOverview.vue` (`tagClass`), and
`site/SitePerformance.vue` (`getColor`, `iconColor`), all under `app/components/`.
These are evidence anchors, not permanent scanner allowlists. Do not count CSS
selector references such as `:deep(.rounded-xl)` again as Tailwind use.

### Six rule families

| Rule | P0 count |
| --- | --- |
| `tailwind-appearance` | Each actual Tailwind visual utility occurrence under the classification below |
| `tailwind-arbitrary-appearance` | Each appearance occurrence whose base utility contains an arbitrary `[…]` value/property or `(…)` variable shorthand |
| `raw-visual-value` | Each static hex or numeric `rgb()`/`rgba()` color literal outside the approved declarations below |
| `important` | Each CSS `!important` annotation, including optional whitespace after `!` |
| `deep-selector` | Each authored `:deep(` selector entry, including optional whitespace before `(` |
| `legacy-dark-entry` | Each complete legacy class-name occurrence listed below, plus each style `:global(.dark…)` entry |

Arbitrary visual utility occurrences are a **subset** of appearance and appear
in both rule budgets. They MUST NOT be counted a third time as raw color values.
Totals across rules are rule hits, not disjoint violations.

Tailwind candidates are validated against the lockfile's Tailwind v4 default
theme/utility compiler before classification, so semantic names such as
`blur-wrapper` do not become false positives. Strip variant prefixes at colons
outside brackets/parentheses and leading/trailing utility importance markers;
classify the base utility, keeping the original variant chain as one occurrence.
The P0 appearance families are:

- `rounded`, `shadow`, `inset-shadow`, `drop-shadow`, `text-shadow`, `ring`,
  `inset-ring`, `opacity`, `leading`, `tracking`, `font`, and their utilities.
- `blur`, `brightness`, `contrast`, `grayscale`, `hue-rotate`, `invert`,
  `saturate`, `sepia`, and their `backdrop-*` equivalents, including backdrop opacity.
- `transition`, `duration`, `delay`, `ease`, `animate`; `italic`, `not-italic`,
  `antialiased`, `subpixel-antialiased`, `uppercase`, `lowercase`, `capitalize`,
  `normal-case`, `underline`, `overline`, `line-through`, `no-underline`;
  `decoration`, `underline-offset`, `accent`, `caret`, `fill`, `stroke` utilities.
- `text-*`, excluding alignment (`left`, `center`, `right`, `justify`, `start`,
  `end`) and wrapping/overflow (`wrap`, `nowrap`, `balance`, `pretty`, `ellipsis`, `clip`).
- `bg-*`, excluding attachment (`fixed`, `local`, `scroll`), `clip-*`, `origin-*`,
  sizing (`auto`, `cover`, `contain`), cardinal positioning (`center`, `top`,
  `right`, `bottom`, `left`), `repeat`/`repeat-*` and `no-repeat`.
- `from-*`, `via-*`, `to-*` color stops: standard Tailwind color palette names,
  `black`, `white`, `transparent`, `current`, `inherit`, or arbitrary values.
  Numeric stop positions alone are not P0 color debt.
- `border`, `divide`, `outline` utilities, excluding `border-collapse`,
  `border-separate`, `border-spacing*`, `divide-x-reverse`, `divide-y-reverse`.
- `scale`, `translate`, `rotate`, `skew`, `transform` utilities only under
  `hover`, `focus`, `focus-visible`, `focus-within`, `active`, including group/peer
  variants and named group/peer variants. Ordinary placement transforms are not
  counted here; `hover:scale-[1.05]` is visual in both Tailwind rules.

All other structure is outside these initial counts. Context-sensitive issues
such as a button's internal spacing still require contract review; absence of
a numeric detector is not permission to violate ownership.

Raw colors are `#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`, or a fully static numeric
`rgb()`/`rgba()` call, including comma/space/slash numeric syntax. Each function
counts once; a gradient or shadow with multiple color literals counts each.
Inspect CSS/Less declarations, Vue style blocks, literal inline style/SVG paint
attributes, bound style literals and script-generated visual settings/styles.
P0 manually verified script colors in `BackgroundPreferencesEditor.vue`
(`colors`, `defaults`), `game/detail/insights/GamePlayerTrend.vue` and
`GamePriceHistory.vue` (`renderChart`), and `site/SitePerformance.vue`
(`updateChart`), under `app/components/`. Charts are not a blanket exemption.
Color validation messages such as "#9c846a" are not rendered styles.

P0 does not quantify named/HSL colors, dynamically interpolated colors,
token/channel-derived colors, literal radii, shadow geometry, durations or
typography. Their ownership rules still apply. This is a high-confidence color
baseline, not a claim that all raw visual debt has been measured.

Count `!important` and deep entries in authored styles, including CSS emitted by
script strings: the `<noscript>` style in `ErrorExperience.vue` counts. Tailwind's
`!` marker is handled as a class modifier, not another CSS `!important` annotation.
Existing visual-guard deep allowlists do not remove those occurrences from debt.

The legacy list comes from `scripts/perf/visual-guard.mjs`: `games-page--dark`,
`search-results--dark`, `is-dark-theme`, `spotlight-panels--dark`,
`about-page--dark`, `legal-page--dark`, `updates-page--dark`,
`nav-home-page--dark`, `gf-static-page--dark`, `lottery-page--dark`.
Match complete names in classes, selectors or class-bearing script literals,
not substrings/comments. Include the guard's prohibited `:global(.dark…)` form
in this family. Canonical `html.dark`/`:global(html.dark …)` is not legacy debt.

### Approved existing raw-color declaration locations

Only declarations matching **all three** columns are excluded from
`raw-visual-value`. Match the exact rule selector (normalize whitespace and comma
spacing) and property prefix; do not exempt a file, descendant rule or ordinary
property such as `background`. Grouped lottery roots mean the listed selector
group and its corresponding all-dark group.

| File | Exact declaration selector(s) | Custom-property prefix |
| --- | --- | --- |
| `app/assets/styles/tokens.less` | `:root`, `html.dark` | `--gf-` |
| `app/assets/css/main.css` | `:root`, `html.dark` | `--gf-page-` |
| `app/assets/styles/components/nav.less` | `.gf-nav`, `html.dark .gf-nav` | `--gf-nav-` |
| `app/assets/styles/components/footer.less` | `.gf-footer-shell`, `html.dark .gf-footer-shell` | `--gf-footer-` |
| `app/assets/styles/components/rating.less` | `.gf-rating`, `html.dark .gf-rating` | `--gf-rating-` |
| `app/assets/styles/pages/games.less` | `.games-page`, `html.dark .games-page` | `--games-` |
| `app/assets/styles/pages/games-search.less` | `.games-search-page`, `html.dark .games-search-page` | `--games-search-` |
| `app/assets/styles/pages/nav.less` | `.nav-home-page`, `html.dark .nav-home-page` | `--nav-home-` |
| `app/assets/styles/pages/updates.less` | `.updates-page`, `html.dark .updates-page` | `--updates-` |
| `app/assets/styles/pages/lottery.less` | `.lottery-page, .lottery-activation-page, .lottery-modal`; the same group with `html.dark ` before each root | `--lottery-` |
| `app/assets/styles/pages/static.less` | `.about-page`, `html.dark .about-page` | `--about-` |
| `app/assets/styles/pages/static.less` | `.legal-page`, `html.dark .legal-page` | `--legal-` |
| `app/assets/styles/pages/insights/foundation.less` | `.insights-page` | `--insights-` |

This recognizes existing declaration locations only. It does not approve all
current token hierarchy, aliasing or annotation, nor establish a fourth token
level; consolidation belongs to P2. `foundation.less` currently has structural
values/aliases and excludes no raw colors. Ordinary selectors in every listed
file remain measured. Game-detail-specific text overrides, datepicker `--dp-*`
overrides, `ErrorExperience.vue`'s local theme and `SiteDetailPage.vue`'s local
theme are not approved declaration layers and remain raw-color debt.

P1 MUST compare its initial detector results against this same source snapshot,
investigate discrepancies and document corrections rather than raising budgets
to hide differences. It MUST NOT turn these P0 evidence anchors into exclusions
for newly added files or script-generated styles.

## Verification responsibilities and phase boundary

| Responsibility | Owner / state at P0 |
| --- | --- |
| TypeScript/Vue typing | Existing Nuxt typecheck |
| JS/TS/Vue engineering rules | Planned ESLint, P1 |
| CSS/Less correctness and hygiene | Planned Stylelint, P1 |
| GoFurry-specific ownership and per-file debt budget | Planned `style:policy`, P1 |
| Unit/component logic | Planned Vitest, P3; current focused Node harnesses remain |
| Browser behavior, SSR, hydration and historical regressions | Current Playwright scripts; Playwright Test migration in P3 |
| Screenshot baseline comparison | Planned Playwright Visual, P3; current screenshots alone are not a visual regression gate |
| Performance budget | Existing `perf:guard`, separate from visual correctness |
| Real external services | Explicit development acceptance, not a default PR gate |

P0 MUST NOT add runners, dependencies, CI gates or nonexistent package commands.
Run from `apps/cn/nav-web`: `npm ci`, `npm run typecheck`, `npm run build`.
For later runtime changes, run relevant existing focused scripts from
`package.json` as well. A skipped external acceptance test is not a pass.

Keep `scripts/perf/visual-guard.mjs` and current regression harnesses intact.
P1 implements the policy detector and MUST first reproduce this baseline with
zero regressions on the unchanged app tree. P2 owns token/primitive organization;
P3 owns the testing foundation; P4+ owns staged, visually equivalent migrations.
Each phase MUST finish in a stable, independently deployable state.

P0 MUST NOT change production Vue, CSS/Less, runtime behavior, package/lockfiles,
CI, dependency versions or style directories, and MUST NOT implement #108/#109.
It neither replaces Nuxt/Tailwind/Less nor introduces a new UI framework.
