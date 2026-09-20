# Nav Web frontend contract

## Scope and authority

This contract governs `apps/cn/nav-web` under
[#124](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423).
MUST/MUST NOT are requirements; SHOULD permits a reasoned, documented departure;
MAY denotes an allowed choice. P0 established governance and a one-time debt
baseline; P1 added static enforcement. P2.1 formalized token ownership and
annotation; P2.2 separates visual primitives from compound product styles without
changing their contents or visual behavior. P2.3 separates Generic Modal from
Preferences composition and retires its dead cascade/raw-color debt while
preserving effective rendered states.

For frontend decisions, resolve evidence in this order:

1. Executable code and tests for actual behavior and compatibility.
2. This contract for intended frontend architecture.
3. `apps/cn/nav-web/AGENTS.md` for operational guidance.
4. `docs/frontend/*` for explanation and usage.
5. Historical `apps/cn/nav-web/docs/*` for past migration context only.

`frontend-style-debt.json` records migration state; it is not design authority.
Historical completion claims such as the `v2.2.x` style migration MUST NOT be
interpreted as completion of #124's engineering foundation.

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
| Control height/padding, color, background, border appearance, radius, shadow, ring, typography, opacity, visual hover/focus/motion | Owning primitive, compound component or domain Less using semantic tokens |
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

- `main.css`: Tailwind bootstrap, reset, base elements/scrollbars and generic
  helpers. It MUST NOT own global GoFurry theme/design-token declarations.
- `styles/index.less`: style composition/import root.
- `styles/tokens.less`: canonical owner of global semantic token declarations,
  including the layout-owned `--gf-page-background` and `--gf-page-pattern*` group.
- `styles/mixins.less`: reusable Less behavior built on tokens.
- `styles/primitives/*.less`: shared domain-neutral visual building blocks.
- `styles/components/*.less`: compound product UI, including preferences,
  shell, navigation and footer; this directory is not deprecated.
- `styles/pages/*.less` and existing subdirectories: current page/domain styles.
- `app/components/common`, `nav`, `game`, `site`, `insights`: existing Vue owners.

`index.less` MUST compose `tokens → mixins → primitives → components → pages`.
The primitive order is button, card, chip, input, modal, pagination, rating; retain
the existing relative page-style order. Preferences MUST load after Modal.
`domains/` and the later test layout remain future migration work; contributors
MUST NOT create empty placeholders. Existing raw values and domain theme islands
remain historical debt. Do not copy them into new code.

## Token ownership, naming and lifecycle

| Level | Meaning and ownership | Examples |
| --- | --- | --- |
| Global/Foundation | Reusable product/theme semantics in `styles/tokens.less`, shared across domains and primitives | `--gf-page-background`, `--gf-surface`, `--gf-text-main`, `--gf-border`, `--gf-accent`, `--gf-focus-ring` |
| Primitive-local | Meaning specific to a reusable primitive, declared in its owning stylesheet | `--gf-rating-empty`, `--gf-rating-fill` in `styles/primitives/rating.less` |
| Compound-local | Product composition/state semantics with an exact approved owner, selector and prefix | `--gf-preferences-*` in `styles/components/preferences.less` |
| Domain | Shared visual meaning within a domain, in that domain's existing Less owner | `--games-*`, `--nav-*`, `--updates-*` |

Contributors MUST search existing tokens before adding one. Primitive-specific
meaning SHOULD stay with its primitive rather than being promoted merely because
the primitive is reusable. Domain tokens SHOULD
alias global semantics by default, for example `--games-border: var(--gf-border)`.
An independent domain value requires a genuine business meaning and a documented
rationale, such as discount emphasis differing from ordinary interaction accent.
Component-private geometry, such as cover ratio or title line count, MAY remain
scoped. It is not another theme-token level and MUST NOT create a separate color,
shadow, radius or typography system.

Token identity MUST follow semantic role, not current literal equality.
`--gf-accent` and `--gf-accent-fill`, or `--gf-surface` and `--gf-input-bg`, MUST
remain distinct: emphasis, action fill, general surface and form surface have
different roles, including different dark-theme behavior. Contributors MUST NOT
deduplicate tokens solely because their current values match.

New global names MUST follow `--gf-<semantic-role>[-<variant-or-state>]`, such as
`--gf-surface-hover` or `--gf-accent-fill-hover`. New literal/color/property-first
names such as `--gf-orange-500`, `--gf-bg-foo` or `--gf-color-bar` require an
explicit compatibility rationale. This is not a mandate to rename active tokens.

A new token MUST have a real consumer and stable semantic meaning. Moving a
literal into a token only to satisfy `style:policy` does not meet that requirement.
Contributors SHOULD remove confirmed unused tokens unless an explicit external
compatibility contract requires retention. P2.1 removed unused `--gf-bg-grid-line`;
do not retain dead tokens for hypothetical future use.

Static/Legal roots MUST remain transparent and MUST NOT own the application
canvas; the default layout and `PublicPageBackground` own page background
semantics through `--gf-page-background`. The retired `--gf-bg-page` compatibility
token MUST NOT be reintroduced or replaced with a Static-specific canvas token.
`--gf-static-panel-shadow` owns the shared About/Legal reading-surface elevation;
Dark intentionally inherits the root value. `static.less` owns the full migrated
color-transition property set, its 500ms duration and existing easing.

Updates global selectors MUST use the `updates-*` domain namespace. State
modifiers MAY use `is-*` only when attached to an Updates-owned base selector.
Updates-owned custom properties MUST use `--updates-*`; domain styles MAY consume
`--gf-*` global semantics. P4.4.2 normalizes Timeline/Year/Entry names without
ancestor scoping or specificity changes; declarations and accepted visual behavior
MUST remain equivalent. Stylelint enforces names only: Style Policy token owners
remain `.updates-page` / `html.dark .updates-page` with the `--updates-` prefix.
`--updates-timeline-delay` is a dynamic runtime property, not a global design token.

P2.1 MUST NOT prebuild typography, spacing, control-height, z-index or container
scales. Later promotion requires repeated real needs and a scoped migration.
Examples in the parent plan do not imply that `--gf-success` or `--gf-warning`
exist today. Actual values belong in their source owner, not duplicated in docs.

### Annotation and theme semantics

New/changed token groups MUST have concise comments explaining semantic meaning,
scope, why they exist and why they belong at that level. Exceptional tokens MUST
also explain the specific exception. Group comments SHOULD carry shared context;
individual comments are for exceptions, not a description of every CSS literal.
Global values stay in the single `tokens.less` owner. Its root groups are ordered
Page & Canvas, Surface, Border, Text, Accent & Action, Form Controls, Modal &
Overlay, Feedback, Focus, Elevation & Blur, Shape, Motion.
`html.dark` MUST preserve those meanings and the relative group order; omitted
tokens intentionally inherit root values. Short dark-section labels suffice;
do not repeat the root explanations or invent groups for symmetry.

```less
/* Surface: shared content and interaction surfaces across domains.
 * Global ownership keeps light/dark meaning consistent; consumers should
 * reuse these semantics before introducing domain-specific colors.
 */
```

A comment merely saying "red" is not a rationale. A domain exception comment
should explain its business role and why the global meaning does not fit.
See the [design-system guide](../docs/frontend/design-system.md) for current
repository examples; it is not a second token-value source.

## Shared visual primitives and compound components

Physical placement in `app/components/common/` MUST NOT imply shared visual
ownership. Classify by responsibility and real consumers: Preferences belongs to
P4, PageScrollDock to P4.5, MobileBottomTabBar to P5, and BlurWrapper/LinkTag to
P6 Game. ManagedAssetImage/SteamAssetImage are runtime infrastructure; their
location does not authorize appearance or routing migration. See the current
[ownership map](../docs/frontend/design-system.md#common-directory-semantic-boundaries).

Before deleting a historical common component, contributors MUST prove zero
production consumers by checking PascalCase, Nuxt path-derived/lazy names,
kebab-case, explicit imports, dynamic components/`resolveComponent` and source
paths. Generated registrations are not consumers. Live components MUST remain;
do not rewrite consumers to force deletion. P4.1 makes no file moves or visual
migration. After deletion, inspect `style:policy` for stale-only debt before
running its downward-only updater; never transfer or raise budgets.

A reusable, domain-neutral visual building block MUST be treated as a primitive.
It has stable appearance semantics, composes into larger UI, and SHOULD NOT
depend on a page root or business token namespace. The seven current owners in
`styles/primitives/` are button, card, chip, input, modal, pagination and rating,
exposing `.gf-button`, `.gf-card`, `.gf-chip`, `.gf-input`, `.gf-modal`,
`.gf-pagination` and `.gf-rating`.
Contributors MUST inspect and reuse the appropriate primitive and its variants
before creating a new selector. New consumers MUST NOT redefine its core
appearance through local overrides; extend the owning primitive deliberately
when a reusable variant is needed.

A product-level UI composition SHOULD remain in `styles/components/`, even when
reused across routes. Navigation, footer and shell are compound owners, not
primitives. Their product-local semantics such as `--gf-nav-*` and `--gf-footer-*`
MAY stay with those owners; reuse alone does not promote them to global tokens.
`primitives/modal.less` owns generic `.gf-modal*` appearance, shared by Preferences
and NSFW confirmation. `components/preferences.less` owns Preferences overrides,
tabs, source selectors, carousel/arrows and the product-only `preferences-toggle`.
Contributors MUST NOT reintroduce `gf-modal__toggle` or promote a control with
one product-specific consumer into a speculative generic primitive.

Preferences input/toggle idle, focus and active semantics MUST stay local under
`--gf-preferences-*`, declared only at `.gf-preferences-modal` and
`html.dark .gf-preferences-modal` in `components/preferences.less`. Theme selects
token values; control selectors select state. Compound-local approval MUST match
the exact file, root selectors and prefix; it is never a whole-file exemption.
Private `.preferences-pages`/`.preferences-page` and Background/Hero/ResourceRoute
editor structure MUST stay scoped, not accumulate in the shared compound owner.

Moving a primitive MUST preserve its selectors, declarations, token values and
variants. A structural move is not authorization to clean up typography,
hover/focus behavior or other historical appearance.

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
primitive-local declarations, approved compound-local declarations, justified
domain declarations, and precise documented exceptions. The narrow
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

Moving a debt-bearing source file is a **policy migration**, not a mechanical
rename: the old path becomes stale and the new path has budget zero. Future moves
MUST explicitly migrate debt ownership under review as a policy change, proving
no new debt is granted. Neither a rename nor `style:policy:update` transfers a
budget automatically. P2.2's six selected files have no per-file style debt to
transfer, so their move MUST leave the manifest and all six totals unchanged.
P2.3 retires the old `components/modal.less` raw-color budget of 29 through dead
cascade removal and semantic ownership: raw debt decreases from 924 to 895.
Neither new stylesheet receives a debt budget; other rule/file counts stay intact.

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
Bracket-only CSS properties follow the same ownership: `[color:#fff]` is visual,
`[width:10px]` is structural, and an arbitrary transform is visual under the
interaction variants below. None occurred in the initial P0 snapshot.

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

### Current approved raw-color declaration locations

The P0/P1 snapshot also approved `app/assets/css/main.css` at `:root` and
`html.dark` for the `--gf-page-` prefix. P2.1 moved those declarations unchanged
into `tokens.less` and removed that obsolete approval. The table below is the
current policy; this ownership migration and removal of the approved dead
`--gf-bg-grid-line` declarations leave all six debt budgets unchanged.

P2.2 moved Rating's approved owner from
`app/assets/styles/components/rating.less` to
`app/assets/styles/primitives/rating.less`, retaining the exact selectors and
`--gf-rating-` prefix. The old path is historical only and is no longer approved.

Only declarations matching **all three** columns are excluded from
`raw-visual-value`. Match the exact rule selector (normalize whitespace and comma
spacing) and property prefix; do not exempt a file, descendant rule or ordinary
property such as `background`. Grouped lottery roots mean the listed selector
group and its corresponding all-dark group.

| File | Exact declaration selector(s) | Custom-property prefix |
| --- | --- | --- |
| `app/assets/styles/tokens.less` | `:root`, `html.dark` | `--gf-` |
| `app/assets/styles/components/nav.less` | `.gf-nav`, `html.dark .gf-nav` | `--gf-nav-` |
| `app/assets/styles/components/footer.less` | `.gf-footer-shell`, `html.dark .gf-footer-shell` | `--gf-footer-` |
| `app/assets/styles/components/preferences.less` | `.gf-preferences-modal`, `html.dark .gf-preferences-modal` | `--gf-preferences-` |
| `app/assets/styles/primitives/rating.less` | `.gf-rating`, `html.dark .gf-rating` | `--gf-rating-` |
| `app/assets/styles/pages/games.less` | `.games-page`, `html.dark .games-page` | `--games-` |
| `app/assets/styles/pages/games-search.less` | `.games-search-page`, `html.dark .games-search-page` | `--games-search-` |
| `app/assets/styles/pages/nav.less` | `.nav-home-page`, `html.dark .nav-home-page` | `--nav-home-` |
| `app/assets/styles/pages/updates.less` | `.updates-page`, `html.dark .updates-page` | `--updates-` |
| `app/assets/styles/pages/lottery.less` | `.lottery-page, .lottery-activation-page, .lottery-modal`; the same group with `html.dark ` before each root | `--lottery-` |
| `app/assets/styles/pages/static.less` | `.about-page`, `html.dark .about-page` | `--about-` |
| `app/assets/styles/pages/static.less` | `.legal-page`, `html.dark .legal-page` | `--legal-` |
| `app/assets/styles/pages/insights/foundation.less` | `.insights-page` | `--insights-` |

This recognizes existing declaration locations only. It does not approve all
current token hierarchy, aliasing or annotation, nor establish another token
level; further consolidation belongs to later phases. `foundation.less` has structural
values/aliases and excludes no raw colors. Ordinary selectors in every listed
file remain measured. Game-detail-specific text overrides, datepicker `--dp-*`
overrides, `ErrorExperience.vue`'s local theme and `SiteDetailPage.vue`'s local
theme are not approved declaration layers and remain raw-color debt.

P1 MUST compare its initial detector results against this same source snapshot,
investigate discrepancies and document corrections rather than raising budgets
to hide differences. It MUST NOT turn these P0 evidence anchors into exclusions
for newly added files or script-generated styles.

## Verification responsibilities and phase boundary

| Responsibility | Current owner / phase |
| --- | --- |
| TypeScript/Vue typing | Existing Nuxt typecheck |
| JS/TS/Vue engineering rules | Nuxt-compatible ESLint flat config; official bulk suppressions track historical findings |
| CSS/Less correctness and hygiene | Conservative Stylelint recommended rules with CSS/Less/Vue parsers |
| GoFurry-specific ownership and per-file debt budget | Parser-backed `style:policy`; Node built-in tests verify the guard |
| Pure TS/domain/utility logic | Vitest `unit` project, Node environment, `tests/unit/*.test.ts` |
| Nuxt runtime/composables | Vitest `nuxt` project with `@nuxt/test-utils`, happy-dom, `tests/nuxt/*.nuxt.test.ts` |
| Repository/source/config/semantic contracts | Existing Node Insights/SEO Contract Guards, outside Vitest |
| Migrated browser behavior, SSR, hydration and historical regressions | Playwright Test in `tests/browser`: Game Detail, Resource Routing/Managed/Steam, Hero lifecycle/Local, Preferences foundation, Fixed/BigInt, Catalog and handoff |
| Stable pixel appearance | `playwright.visual.config.ts` / `test:visual`; P3.3.1 establishes only the pinned environment sentinel, no golden baselines |
| Legacy broad page visual/report checks | Retained `scripts/perf/visual-guard.mjs` / `visual:guard`; independent of Visual snapshots and static `style:policy` |
| Performance budget | Existing `perf:guard`, separate from visual correctness |
| Real external services | Explicit development acceptance, not a default PR gate |

P0 introduced no runners, dependencies or CI gates. P1's local/CI sequence is
`npm ci`, `npm run lint`, `npm run stylelint`, `npm run style:policy:test`,
`npm run style:policy`, `npm run typecheck`, `npm run insights:semantics`,
`npm run seo:recovery:test`, `npm run build`, from `apps/cn/nav-web`.
For runtime changes, run relevant existing focused scripts from `package.json`
as well. A skipped external acceptance test is not a pass.

Keep `scripts/perf/visual-guard.mjs` and current regression harnesses intact.
P1's detector reproduces the P0 baseline on the unchanged app tree. P2 owns token/primitive organization;
P3 owns the testing foundation; P4+ owns staged, visually equivalent migrations.
Each phase MUST finish in a stable, independently deployable state.

P3.1 uses Vitest `projects` with separate `test:unit` and `test:nuxt` commands;
`npm test` runs both. CI MUST execute separate Unit tests and Nuxt tests steps
before typecheck, retained Insights/SEO guards and build. Choose the lowest-cost
environment that faithfully represents the tested behavior. Browser-like globals
alone do not require Nuxt; pure tests SHOULD use Vitest-controlled stubs/cleanup.

Nuxt-dependent cases MUST use the real Nuxt runtime and reset cookies/useState
between cases using supported APIs. Mock only business injection boundaries,
not Nuxt's state/cookie/app context. Source-transpile/data-URL import hacks and
handwritten fake Nuxt runtimes MUST NOT replace ordinary imports/runtime tests.
Keep `@nuxt/test-utils/module` out of production Nuxt config. Do not refactor
production behavior merely to accommodate tests. The first four legacy suites
are replaced, not duplicated; [testing guidance](../docs/frontend/testing.md)
records their assertion ownership. Style-policy's Node runner, Contract Guards,
browser smoke, `visual:guard` and external acceptance retain independent scope.

P3.2.1 establishes a Chromium-only Playwright Test gate against the production
Nitro build, never `nuxt dev`. CI MUST build successfully before installing
Chromium and running `test:browser`; retries are zero and CI has one worker.
Domain fixtures own worker-scoped production servers, close them at teardown,
and reset mutable state for every case. Playwright owns per-test browser contexts
and pages; do not introduce a global `webServer` or order-dependent serial suites.
Capture/assert browser exceptions and hydration errors explicitly. Retain traces
and screenshots only on failure, with video off; CI uploads failure diagnostics.
Game Detail and Resource Routing have migrated, including Preferences, Managed
and Steam resource snapshots/fallback, and legacy Steam recommendation-only
migration. Routing probe latency/failure/block/gate state MUST be test-scoped;
the worker owns only stable production Nitro/local API resources. Managed/Steam
requests MUST use deterministic local responses, never real CDN services.
P3.2.3 also migrates Hero lifecycle/Local, Preferences foundation, Fixed/BigInt,
Catalog and staged handoff. The two Hero domain fixtures MUST remain separate;
each test owns fresh catalogs, request evidence, failures and independent gates.
Teardown MUST release API/image/IndexedDB gates and clear the active scenario;
the worker's API resolver MUST reject access without an active scenario.
Only the Hero-specific assertion may acknowledge the existing mobile homepage
Footer hydration debt: width below 768, exactly one mismatch, no SSR footer,
exactly one client footer, retained SSR Hero node, and no other browser errors.
Generic error capture MUST NOT ignore hydration. Same-run clipped paint Buffer
and computed-style equality MAY remain runtime invariants; success screenshots,
computed audit JSON and golden baselines MUST NOT be introduced in P3.2.3.
Insights and other legacy browser suites keep their runners. `npm test` remains
Vitest-only; visual baselines belong to P3.3.

P0 MUST NOT change production Vue, CSS/Less, runtime behavior, package/lockfiles,
CI, dependency versions or style directories, and MUST NOT implement #108/#109.
It neither replaces Nuxt/Tailwind/Less nor introduces a new UI framework.

### P3.3 Visual environment and baseline governance

Functional `playwright.config.ts` MUST exclude `tests/browser/visual/**` and keep
the existing Smoke/Regression gate unchanged. `playwright.visual.config.ts` owns
Visual tests with Chromium only, one worker, zero retries and a 60-second timeout.
The environment MUST fix headless mode, 1440×900, zh-CN, UTC, DPR 1, light default
and reduced motion. Failure-only trace/screenshots and no video use independent
`playwright-visual-report/` and `visual-test-results/` directories.

Authoritative Visual CI MUST use the official Playwright image matching the
package version, pinned by a verified immutable digest, with `--ipc=host` and
Node 24. The exact identity and local commands live in
[testing guidance](../docs/frontend/testing.md#visual-runner-and-pinned-environment-p331).
The separate `nav-web-visual` job MUST depend on successful `nav-web` and a Nav Web
change, install packages and rebuild inside that container, then run `test:visual`.
It MUST NOT install browsers, reuse the functional job's `.output`, or update
snapshots. Failure artifacts retain the distinct Visual report/results for seven days.

Baseline update is an explicit visual-change review action, not a test-fix
command. Agents MUST NOT update baselines just to make CI green. Updates require
explicit maintainer/user approval or a task explicitly authorizing the visual
migration, and MUST use `test:visual:update` in the pinned environment. Its guard
requires Linux, Node 24 and `GOFURRY_VISUAL_ENV=pinned`; the marker is an operator
assertion, not approval or an image fingerprint. Unpinned `test:visual` runs are
diagnostic only. Unexpected differences require investigation and implementation
fixes followed by comparison, not automatic acceptance.

Baselines MUST live in tracked `tests/browser/visual/__snapshots__/`,
using deterministic test-file and explicit snapshot-name paths. No global
`maxDiffPixels` / `maxDiffPixelRatio` or broad threshold widening is allowed;
any future tolerance needs a small, local, justified exception. Playwright
packages, Docker tag/digest, browser revision and visual baselines MUST be reviewed
as one atomic upgrade unit. P3.3.1 MUST create no golden PNG, screenshot assertion,
production test route or UI fixture. The sentinel launches real Chromium and
checks the environment without taking a screenshot; real visual contracts start
in P3.3.2. Keep `visual:guard` and direct `playwright` until their scoped retirement.

P3.3.2's `ui-foundation.spec.ts` owns stable shared primitive appearance through
exactly four locator baselines (light/dark × desktop/mobile). Its test-only
markup MUST consume production Nitro CSS loaded from `/about`, preserve the
production head and exclude product JavaScript. Fixture CSS MAY own structure;
it MUST NOT redefine control appearance. Only the canvas MAY set background/text
using production page/text tokens. Generic Modal MUST remain outside the isolated
Preferences Toggle token scope. Verify tokens, scope, fonts, focus and overflow
before capture; do not mask instability or widen tolerances. Approved baseline
creation MUST pass two consecutive pinned comparisons and receive maintainer
review of all four images. Real Preferences/backdrop composition belongs to
P3.3.3 and business surfaces to P4+, not this Foundation fixture.

P3.3.3's real Preferences visual contract MUST exercise production Nuxt/Vue,
NavBar, Teleport, backdrop and child components through the existing Hero
Preferences fixture. Optional visual seeds MUST preserve functional defaults.
Tests MAY hide unrelated underlying page content and use the semantic canvas;
they MUST NOT fake or restyle Preferences. Theme MUST use the production
localStorage/Theme Store path. Routing time and fresh diagnostics MUST be fixed
only in the test context, with no automatic probe traffic before capture.
Real tab selection, finite animations/fonts, backdrop/filter/containment and
active-page overflow MUST settle semantically, without arbitrary sleeps, masks
or tolerance widening. The eight-case matrix is Home light/dark desktop/mobile,
Background light desktop/mobile and Routing light desktop/mobile. It MUST remain
separate from functional regressions and business surfaces; expansion requires
explicit review. Creation requires pinned generation, two consecutive comparisons
and maintainer review of the eight new images; the four Foundation images remain
unchanged. Real product defects MUST be reported without expanding this phase
into production UI changes.

P4.3.1's Static/Legal Visual baselines MUST exercise real production Nitro,
Nuxt SSR/hydration, Theme Store, default layout and `PublicPageBackground`.
Layout owns the canvas; the Static root MUST remain computed-transparent.
Fixtures MAY hide only PageScrollDock and MobileBottomTabBar, never restyle
Static/Legal content. The six viewport baselines cover About light/dark at
desktop/mobile and Terms light/dark on mobile. Readiness MUST use load plus
semantic image/animation/font waits, with zero external/upstream requests and
the current computed color-transition contract. Creation requires pinned
generation, two consecutive comparisons and maintainer review. Appearance
migrations MUST pass these accepted images without snapshot updates or changes
to layout-owned canvas semantics. All eighteen baselines MUST stay intact.

Updates business-surface baselines MUST use the real production SSR/hydration
path with deterministic ready-state `/nav/updates` data. Functional coverage
owns year expansion and load-more persistence; the shared fixture MUST prove
exactly one SSR API call, payload reuse and zero external/unrelated requests.
The four viewport baselines retain Theme Store, PublicPageBackground and real
reduced motion. Creation requires pinned generation, two consecutive comparisons
and maintainer review. Updates selector normalization MUST pass accepted goldens
without snapshot updates; no selector or token cleanup accompanies P4.4.1.

Error Experience baselines MUST exercise the real Nuxt missing-route error
boundary and its own canvas. Hydrated coverage owns production theme/artwork
readiness and normal-motion keyboard focus visibility; a genuinely JS-disabled,
normal-motion regression MUST preserve the readable default Light fallback.
Appearance migration MUST pass the four accepted Error locator goldens and both
behavior contracts without snapshot updates. Initial creation requires pinned
generation, two consecutive comparisons and maintainer review.

### P1 enforcement and maintenance

ESLint uses the official Nuxt static flat-config factory without a runtime module
or formatting policy. `eslint-suppressions.json` is the one-time historical
bootstrap, not a license to suppress new code. Contributors MUST fix new lint
findings, MUST NOT rerun bulk suppress-all to grant debt, and MUST NOT use
`--pass-on-unpruned-suppressions`. Use `npm run lint:prune` when removing debt;
unused suppressions fail the normal lint command.

Stylelint owns correctness only. Its config documents narrow Less/Tailwind/Vue
compatibility decisions and existing cascade patterns; it does not enforce
formatting or duplicate the six architecture-debt rules. The configured current
source must have zero Stylelint violations; no Stylelint debt manifest exists.

`scripts/style-policy.mjs` discovers tracked and non-ignored new application
source, parses SFC/TS/CSS/Less, extracts shared style facts, applies the six
detectors and exact exceptions, then compares every rule/file budget:

- `actual > baseline`: regression, fail.
- `actual < baseline`: stale budget, fail.
- `actual == baseline`: pass.

`npm run style:policy:update` can only lower budgets or remove zero entries. If
any pair increased, it refuses every write. It preserves exceptions and refuses
to overwrite a manifest changed during the scan. New files default to zero;
moving debt never transfers its budget automatically.

Class extraction works backward from template/DOM class sinks through bounded
local constants, arrays/objects, refs/computed values and simple function returns.
Script visual settings/palettes and actual `innerHTML`/`v-html` embedded styles
are supported without interpreting arbitrary runtime code. Authored literals
are deduplicated across references. Unresolved runtime expressions are an
explicit static-analysis boundary, not permission for new architecture debt.
Cross-file execution, imported function evaluation and general dynamic string
solving are outside P1. Parse/read/manifest/compiler failures fail closed.

Tests verify extraction, classification, exact exceptions, all three comparison
outcomes, update safety, fail-closed CLI behavior and current per-file parity.
The existing Nav Web CI job runs each guard separately before the preserved
typecheck, Insights semantics, SEO recovery and build steps. P1 adds no Vitest,
Playwright Test migration, production style cleanup or UI behavior change.
