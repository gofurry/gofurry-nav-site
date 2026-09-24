# Nav Web design system

The [frontend contract](../../contracts/nav-web-frontend.md) owns the rules.
This guide explains how the existing source applies them. P2.1 established token
ownership and annotation; P2.2 gives domain-neutral primitives an explicit owner;
P2.3 separates Modal from Preferences and removes its dead cascade. These phases
preserve rendered appearance and introduce no token scales. Actual values
belong in source, not this guide.

P4–P6 migration is complete; the phase names below preserve implementation and
cascade decisions, not pending cleanup authorization. Current verification and
remaining owners are recorded in [#124 closure](../acceptance/issue-124-frontend-engineering-closure.md).

## Visual layers

The [Less entry](../../apps/cn/nav-web/app/assets/styles/index.less) composes
`tokens → mixins → primitives → components → pages`:

| Layer | Responsibility |
| --- | --- |
| `tokens.less` | Shared global foundation semantics |
| `mixins.less` | Reusable Less behavior consuming tokens |
| `primitives/` | Domain-neutral button, card, chip, input, modal, pagination and rating building blocks |
| `components/` | Compound product UI: preferences, shell, navigation, footer, error experience and scroll dock |
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
The existing page/domain files remain canonical; `domains/` is not a mandatory
follow-up or a reason to move files without a concrete product need.

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

## Common directory semantic boundaries

`app/components/common/` is a physical location, not a shared visual owner.
Current consumers establish these boundaries; no file moves accompany P4.1:

| Surface | Evidence and semantic owner |
| --- | --- |
| Preferences family | NavBar/MobileBottomTabBar open ModeSettingModal; its Background, Hero and Resource Routing editors and their child controls belong to P4 Preferences |
| PageScrollDock | Mounted by `app.vue`; global scroll tooling belongs to P4.5 |
| MobileBottomTabBar | Mounted by the default layout; mobile navigation belongs to P5 |
| BlurWrapper / LinkTag | Consumed by GameDetailMain / GameSidebarLinks; Game semantics belong to P6 |
| ManagedAssetImage / SteamAssetImage | Managed/Steam loading and fallback infrastructure used by Nav, Site and/or Game; not appearance migration targets based on directory placement |

P4.1 removed EmptyState, ErrorState and LoadingState only after finding no
production consumers across template names (including Nuxt-prefixed/lazy and
kebab-case forms), imports, dynamic component resolution and source paths.
Insights' `LoadingState` type and empty-state CSS/i18n names are unrelated;
React Admin owns separate state components. The policy updater removes the three
dead files' 11 + 8 + 2 Tailwind appearances, without moving debt or changing live
styles. #108/#109 consumers remain excluded from proactive migration.

## Footer, App Shell and global browser chrome

P4.6.2 completes Footer's compound appearance in
[footer.less](../../apps/cn/nav-web/app/assets/styles/components/footer.less).
Its local `--gf-footer-*` tokens own canvas, text/link hierarchy and brand hover
glows; Dark overrides only theme-sensitive values. Vue retains structural
Tailwind and business link data, with `data-brand` identifying social icons.
Typography includes the original section-title and meta line heights; brand
glows retain the generated Tailwind hover-capability media condition.

Audit correction to the P4.6.2 plan: the old Dark link selector's specificity
kept ordinary and meta links at their resting color on hover. Dark therefore
aliases `--gf-footer-link-hover` to `--gf-footer-link`, preserving that verified
behavior rather than introducing the plan's assumed cross-theme white hover.

[shell.less](../../apps/cn/nav-web/app/assets/styles/components/shell.less) owns
the App Shell's full color-transition contract. The global scrollbar consumes
foundation `--gf-scrollbar-*` browser-chrome semantics in `main.css`, retaining
selectors, geometry and the shared Light/Dark treatment. Accepted Footer/Shell
tests and goldens remain the authority; this migration adds no new visual design.

## Navigation shell appearance

P5.1.2 gives NavBar and MobileBottomTabBar one compound appearance owner in
[nav.less](../../apps/cn/nav-web/app/assets/styles/components/nav.less).
Local `--gf-nav-*` semantics belong to `.gf-nav` and `.mobile-bottom-tabs`, with
their exact `html.dark` counterparts. BottomTab's physical `common/` location
does not create a new visual owner or require a separate stylesheet.

NavBar retains structural Tailwind and explicit appearance hooks; its menu
transition, typography and surface treatment live in Less. BottomTab retains
private placement, breakpoint and runtime geometry in scoped styles, while Less
owns its surface, icon, active/hover and motion treatment. Active border and fill
keep separate semantic identities. The accepted Nav Shell contracts remain
unchanged; Nav Home Header/Content, SearchBox and QuickAccess have separate P5
ownership from this shell migration.

Audit correction to the P5.1.2 plan: button `text-sm` was overridden by the
unlayered `font: inherit` rule in `main.css`. Theme/Mode and Mobile action/language
buttons therefore retain inherited typography (currently 16px/24px), while brand
and navigation links retain their effective 14px/20px typography. Migration
preserves accepted computed appearance, not an overridden utility's intent; no
button typography override or otherwise-unused theme-button hook is needed.

## Nav Home Header appearance

P5.2.2 keeps Header, Search, QuickAccess and Quick Sites in
[pages/nav.less](../../apps/cn/nav-web/app/assets/styles/pages/nav.less), using
`--nav-home-*` on `.nav-home-page`; there is no separate Header stylesheet.
Hero-backed surfaces remain theme-independent. Mobile shadows/backgrounds are
root-declared semantics consumed by media rules, not Dark overrides. Compound
shadows and filters preserve their full values; shared Quick Sites roles reuse
tokens for borders, subtle surfaces and text hierarchy.

Search retains the accepted category 14px/20px and platform 12px/16px typography,
500 weight and 12px radius. Its effective background/box-shadow/color 500ms
transition remains authoritative over the removed `transition-all` utility.
The unitless line-height ratios retain Tailwind's accepted built precision;
rounding them to exact rem heights shifts layout by 1/64px despite equal computed
pixel labels and changes input text rasterization.
`SiteIconStrip` and its complete stylesheet block were removed after confirming
zero consumers, rather than tokenizing dead code. P5.3 owns revealed Content,
Spotlight, ToolDock, TransitionBar, Cards and Popovers.

## Nav revealed Core appearance

P5.3.2 keeps Cards, Groups, Popovers and TransitionBar in `pages/nav.less` with
`--nav-home-*` semantics. Card titles retain 16px/24px/500; descriptions retain
12px/16.2px (`line-height: 1.35`). Existing token names and compound shadow values
remain intact; the redundant Dark card background declaration is removed.

Audited corrections to the page-root-only plan are explicitly approved:
body Teleports declare their own narrow tokens on `.site-popover`,
`.group-popover` and `.nav-transition-bar__author` (and their Dark roots).
SitePopover preserves its actual transparent background and zero border, rather
than activating previously unreachable page-root surface tokens. Cards and group
toggles also serve Site Groups outside `.nav-home-page`, so their precise
`--nav-home-card-hover-*` and `--nav-home-group-toggle-*` states belong on those
shared consumer roots. Policy registration permits only these selector/prefix
pairs, not whole-file exemptions. The existing games-page cascade is unchanged.

P5.3.3/P5.3.4 subsequently completed ToolDock/Spotlight tokens at the same page
roots, and removed the proven-dead content-loading selector. P5.4 decoupled Site
Groups into `.site-group-page` / `--nav-site-group-*`. These completed migrations
do not change the Core/Teleport ownership decisions above.

## P4 exit ownership (historical assignment, now completed through P6)

The remaining manifest was audited by semantic consumer, including Common and
Game Detail insights widgets. No unassigned Stable/Common debt remains.

| Remaining surface | Next owner |
| --- | --- |
| NavBar, Nav domain, site-groups, MobileBottomTabBar | P5 Nav Surface |
| Game, Lottery, BlurWrapper, LinkTag, Game Detail insights widgets | P6 Game Surface |
| Site Detail components | #109 |
| Insights page/domain styles | #108 |
| Experimental ambient effects | Intentionally preserved experimental |

`frontend-style-debt.json` remains the exact machine-owned inventory. These
assignments are phase boundaries, not authorization to clear debt early; ambient
effects must not be deleted merely to lower totals.

## Error and PageScrollDock appearance

P4.5.3 moves Error appearance into
[error.less](../../apps/cn/nav-web/app/assets/styles/components/error.less),
including its local `--gf-error-*` canvas/text/action semantics, artwork mask,
typography and motion. Its scoped SFC retains private layout. No-JS visibility
uses normal declarations with greater specificity; keyboard focus cancels the
staged action animation before revealing controls, without `!important`.

[page-scroll-dock.less](../../apps/cn/nav-web/app/assets/styles/components/page-scroll-dock.less)
owns Dock ring/core/shadow/text and hover appearance through compound local
`--gf-scroll-dock-*` tokens. Placement, grid and mobile unmount/display structure
stay in the SFC. The step-delay and progress properties remain dynamic Vue
channels, preserving their original values and behavior.

Single-consumer visual semantics stay component-local; clearing debt alone does
not justify promotion to global tokens. Both owners load in the component layer
before pages and must preserve the accepted Error/Dock runtime and pixel
contracts without modifying tests or snapshots.

## Updates domain namespace

P4.4.2 normalizes generic Timeline/Year/Entry selectors to `updates-timeline*`,
`updates-year-group*` and `updates-entry*`, retaining declaration values,
specificity and visual behavior. The unused template wrapper class is removed;
its list item remains. `is-*` state modifiers remain valid on domain-owned bases.
Stylelint enforces this namespace only in `pages/updates.less`; it allows
`--updates-*` custom properties and consumption of `--gf-*` global semantics.
Style Policy retains its exact root/theme token ownership, and all 41 existing
Updates tokens keep their identities and values. `--updates-timeline-delay` is
the Vue-supplied per-year animation delay, a dynamic Updates-local property rather
than a global design token. Existing runtime tests and accepted goldens protect
this boundary without snapshot or debt updates.

## Games Home News ownership

P6.1.4 completes direct Home appearance with News in the existing `pages/games.less`.
Its ten `--games-home-news-*` roles belong to `.games-page` and
`html.dark .games-page`: text/muted; card background/shadow/hover background;
control text/hover background/hover text; and progress track/fill backgrounds.
These are independent News semantics, not aliases granted because another
surface currently has equal literals. Progress keeps its complete color-mix
meaning. No global scale, stylesheet or policy owner is added.

Typography stays in the same Less owner; Vue retains structure and behavior.
The accepted computed line-height precision is authority: preserve the existing
ratio and build result rather than rounding the rendered value. Authored alpha
values likewise remain distinct from browser serialization. News card typography,
clipping, motion and hover semantics remain frozen by the P6.1.3 contracts.

Home closure originally excluded shared ReviewDialog and SidebarSearch. P6.3.2
retired the latter's 13 raw occurrences, P6.5.2 retired the 19 Detail/BlurWrapper
occurrences, and P6.6 resolves the final shared Dark root occurrence. Their
separate runtime and visual contracts remain authoritative after Games closure.

## Search and shared SidebarSearch appearance

P6.3.2 keeps Advanced Search, Filter/Jump and result motion in the existing
`pages/games-search.less`. The page and body-overlay roots share `--games-search-*`
semantics, including media-overlay borders and Datepicker feedback adapters.
The layout still owns the transparent Search canvas. SFCs retain track geometry,
container layouts and unchanged timers, inline transforms and interaction logic.

Shared SidebarSearch stays in `pages/games.less`: eight
`--games-sidebar-search-*` roles cover input/focus, panel and card/hover states
on the existing Games roots. Keep its higher-specificity Dark consumer selectors;
merely moving all theme differences into root variables would let later Search
rules change accepted pixels. P6.6 retains the audited Search `games-page` bridge.

Input `text-sm` did not own the rendered typography: Sidebar/Jump inputs inherit
16px/24px through the existing reset. Jump title remains 14px/20px, while Datepicker
keeps 13.76px/24px and simple-result title 13.12px/15.088px. Preserve these measured
cascades and complete transition properties/durations, not overridden utility
intent. Domain focus selectors retain their reset values without `!important`.
The accepted 14 Search goldens and all earlier images remain unchanged.

## Shared Review appearance

P6.2.2 gives the Home/Search/Detail Review Dialog its own compound owner,
`components/game-review-dialog.less`. The `.review-dialog-backdrop` body Teleport
and its Dark root declare `--games-review-*`; no page-root inheritance is assumed.
Vue keeps private geometry and behavior. Global Modal/Text/Form/Action aliases
provide the approved deep slate Dark theme, while focus/feedback and the accepted
Light treatment keep their distinct Review semantics. This does not apply the
Generic Modal markup or its geometry. The four Dark Review goldens are deliberately
revised after P6.2.1 maintainer feedback; Light and the other 64 images are frozen.
Field/button typography follows the actual inherited 16px/24px cascade, and the
title retains its audited line height. Accessibility and request lifecycle changes
remain outside this appearance phase.

## Lottery appearance and closure

After the twelve P6.4.1 images received maintainer acceptance, P6.4.2 moves
Prize/Join/Activation appearance to the existing `pages/lottery.less` owner.
Its full three-root selector group and Dark counterpart declare `--lottery-*`;
the body-mounted modal remains independent of page ancestry. Thirteen new roles
cover complete canvas washes, summary fill, dialog elevation, action contrast/
hover, activation card composition/elevation and status border/fill/text. No
palette redesign, global token or policy expansion accompanies this migration.

The layout still owns the transparent page canvas; existing Lottery background
tokens and theme overrides are retained. Form controls preserve the reset's
inherited 16px/24px/400 instead of activating old `text-sm`/`font-medium` intent.
Preserve built unitless line-height precision, including small summary labels,
the measured 12px backdrop blur and complete transition property sets. Accepted
runtime tests and all 98 images are unchanged. P6.4 is closed; Detail/Common and
Games-wide final cleanup remain separate P6.5/P6.6 work.

## Detail and Common consumer ownership (P6.5.2)

The maintainer accepted P6.5.1's twenty Detail images before appearance migration.
`pages/games.less` owns Detail tokens under `.game-detail-page` / Dark and
`--games-detail-*`. The body-mounted Lightbox declares only its own required
palette on its existing exact roots. BlurWrapper and LinkTag stay in `common/`
but inherit Detail appearance; they do not establish a global Common palette.
Similar's scoped styles retain geometry/containment, while domain Less owns
typography, cover treatment and paging motion.

Preserve accepted computed appearance: native buttons may inherit fonts despite
old Tailwind typography, and Similar's smaller reason text inherits its parent's
unitless line-height ratio. Do not round those ratios or rewrite authored colors
from browser-serialized values. Game chart roles live in the Detail palette;
the Canvas bridge reads resolved element tokens after real theme application,
without a second JS palette, and keeps ECharts instances shallow.

Only Game-specific Insights appearance is included. Site capability and shared
Insights styles keep their #109/#108 owners. All 118 goldens and runtime contracts
remain fixed; P6.6 audits and retains the live Games compatibility bridge.

## Games completion and retained compatibility (P6.6)

Home, Search and Detail retain `.games-page`; it is an audited compatibility
scope, not permission to grow a new shared palette. `pages/games.less` and its
existing Light/Dark roots retain these roles (names below omit `--games-`):

| Roles | Actual consumers |
| --- | --- |
| `item-hover-bg`, `item-border`, `item-hover-border`, `text-main`, `focus-ring` | Shared SidebarSearch |
| `text-title`, `accent`, `accent-strong`, `badge-bg`, `shadow-soft` | Detail headings, links, tags, media and cards |
| `text-muted`, `text-soft`, `accent-muted`, `shadow-popover` | Detail and shared SidebarSearch; supporting Detail overlays |
| `text-body` | Root text, preserving the distinct Dark body contrast |

Light Detail's text aliases and Dark Detail's reverse reads remain deliberately
asymmetric. Keep the explicit Dark root color consumer and its specificity;
muted text is not the root body role. The layout owns the transparent canvas and
the existing root color/background transitions remain unchanged.

P6.6 removes 26 unreferenced legacy roles plus `shell-bg`, whose only dependent
was the dead `sidebar-bg`. It also removes the shell-overridden `page-bg` and its
two consumers: 28 names / 56 theme declarations in total. Legacy ToolDock panel
transition rules and its unused active branch are removed; real hover and the
dynamic InfoGroup/Sidebar transitions remain. Chart tokens are live dynamic
readers, not dead just because they lack literal CSS `var()` consumers.

`primitives/pagination.less` owns the explicit `.gf-pagination--plain` variant's
transparent border/fill and square corners. Search selects that variant instead
of resetting `.gf-pagination__button*`. Its existing domain adapter still owns
text emphasis, active underline, geometry, typography and motion. Default
Pagination and the accepted Search computed appearance remain unchanged.

The six remaining Games SFC scoped blocks retain private layout, containment,
spacers and runtime geometry. Review/Filter/Jump/Lightbox remain body-owned;
NSFW and Rating retain their existing primitives. No global Common palette or
file relocation is needed to close P6.

Remaining baseline ownership is explicit: Site #109 owns Tailwind 632, arbitrary
5, deep 33 and raw 388 (382 in Site SFCs, 6 Site capability values in
`insights.less`). Ambient's 75 raw values remain experimental. The three shared
Insights foundation `!important` declarations and two Insights domain declarations
remain #108/shared-boundary work; shared consumers do not authorize Game-only
removal. P6 closure does not authorize #108/#109 changes. P7 subsequently retires
the legacy runners after mapping their assertions into the formal gates.

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

P4.3.2 retired `--gf-bg-page` after the real Static/Legal visual contract proved
the roots transparent under the shell. Do not reintroduce that fallback or a
replacement canvas token: the layout and `PublicPageBackground` own the canvas.
[static.less](../../apps/cn/nav-web/app/assets/styles/pages/static.less) owns the
full migrated color transition, retaining its duration and easing. Shared
`--gf-static-panel-shadow` was promoted to global Elevation & Blur only after
About and Legal demonstrated real shared usage; Dark inherits the same value.

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

P4.2 keeps Background Preferences' private layout, scrolling and dynamic preview
state in its Vue component. Control, palette, slider and preview appearance lives
under `.gf-preferences-modal` in `components/preferences.less`. Its ordered
`--gf-preferences-background-preset-*` palette and swatch ring are local product
semantics, inherited unchanged by both themes, not global theme tokens. Vue reads
the palette through its editor root's CSSOM; pattern defaults continue to consume
the real `--gf-page-pattern-*` tokens without a duplicate fallback palette.

P2.3 retires 29 raw-color occurrences: 11 overridden values are deleted and 18
effective values keep their meaning in 14 local semantic tokens. The safe updater
removes only the old modal budget (924 → 895), granting neither new file debt.
Use the formal Hero/preferences, resource-routing and game-detail Browser/Visual
specs for themed responsive and effective input/toggle/NSFW modal checks. Their
old smoke scripts are retired; see the current commands in [testing](testing.md).
