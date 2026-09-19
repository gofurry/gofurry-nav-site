# Nav Web frontend contract

## Scope and authority

This contract governs `apps/cn/nav-web` under
[#124](https://github.com/gofurry/gofurry-nav-site/issues/124#issuecomment-5740012423).
MUST/MUST NOT are requirements; SHOULD permits a reasoned, documented departure;
MAY denotes an allowed choice. P0 established governance and a one-time debt
baseline; P1 added static enforcement. P2.1 formalized token ownership and
annotation; P2.2 separates visual primitives from compound product styles without
changing their contents or visual behavior.

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
- `styles/components/*.less`: compound product UI, including modal/preferences,
  shell, navigation and footer; this directory is not deprecated.
- `styles/pages/*.less` and existing subdirectories: current page/domain styles.
- `app/components/common`, `nav`, `game`, `site`, `insights`: existing Vue owners.

`index.less` MUST compose `tokens → mixins → primitives → components → pages`.
The primitive order is button, card, chip, input, pagination, rating; retain the
existing relative page-style order. P2.2 creates only the approved primitive layer.
`domains/` and the later test layout remain future migration work; contributors
MUST NOT create empty placeholders. Existing raw values and domain theme islands
remain historical debt. Do not copy them into new code.

## Token ownership, naming and lifecycle

| Level | Meaning and ownership | Examples |
| --- | --- | --- |
| Global/Foundation | Reusable product/theme semantics in `styles/tokens.less`, shared across domains and primitives | `--gf-page-background`, `--gf-surface`, `--gf-text-main`, `--gf-border`, `--gf-accent`, `--gf-focus-ring` |
| Primitive-local | Meaning specific to a reusable primitive, declared in its owning stylesheet | `--gf-rating-empty`, `--gf-rating-fill` in `styles/primitives/rating.less` |
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

`--gf-bg-page` remains the legacy static-page fallback with its existing consumer
and value. New layout-owned backgrounds MUST use `--gf-page-background`;
contributors MUST NOT use `--gf-bg-page` in new code. Remove that fallback with
the later static-page migration, not by changing its consumer in P2.1.

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
Overlay, Feedback, Focus, Elevation & Blur, Shape, Motion, Legacy Compatibility.
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

A reusable, domain-neutral visual building block MUST be treated as a primitive.
It has stable appearance semantics, composes into larger UI, and SHOULD NOT
depend on a page root or business token namespace. The six current owners in
`styles/primitives/` are button, card, chip, input, pagination and rating, exposing
`.gf-button`, `.gf-card`, `.gf-chip`, `.gf-input`, `.gf-pagination` and `.gf-rating`.
Contributors MUST inspect and reuse the appropriate primitive and its variants
before creating a new selector. New consumers MUST NOT redefine its core
appearance through local overrides; extend the owning primitive deliberately
when a reusable variant is needed.

A product-level UI composition SHOULD remain in `styles/components/`, even when
reused across routes. Navigation, footer and shell are compound owners, not
primitives. Their product-local semantics such as `--gf-nav-*` and `--gf-footer-*`
MAY stay with those owners; reuse alone does not promote them to global tokens.
`modal.less` currently combines generic `.gf-modal` behavior and preferences UI.
Contributors MUST continue to reuse it in place; P2.3 owns its eventual split.
P2.2 MUST NOT move or partially split modal, nav, footer or shell styles.

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
primitive-local declarations, justified domain declarations, and precise
documented exceptions. The narrow
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
| Unit/component logic | Planned Vitest, P3; current focused Node harnesses remain |
| Browser behavior, SSR, hydration and historical regressions | Current Playwright scripts; Playwright Test migration in P3 |
| Screenshot baseline comparison | Planned Playwright Visual, P3; current screenshots alone are not a visual regression gate |
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

P0 MUST NOT change production Vue, CSS/Less, runtime behavior, package/lockfiles,
CI, dependency versions or style directories, and MUST NOT implement #108/#109.
It neither replaces Nuxt/Tailwind/Less nor introduces a new UI framework.

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
