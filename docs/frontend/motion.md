# GoFurry Motion Contract V1

This is the single motion policy for Nav Web. The [frontend contract](../../contracts/nav-web-frontend.md)
still owns styling, SSR, testing and debt boundaries. #139-A establishes this
policy and repairs the audited Reduced Motion gaps; it does not authorize a
site-wide redesign, library migration or changes to accepted Visual goldens.

## Meaning and ownership

| Role | Guidance for future work | Use |
| --- | --- | --- |
| Immediate feedback | 120–180ms, normally 160ms | Hover, focus and selected-state decoration; never delay input |
| Enter / exit | 180–260ms, normally 220ms | Suggestions, overlays and small content changes |
| Spatial transition | 350–500ms, around 420ms | User-directed paging or a meaningful position change |
| Continuous | Only while needed, with a way to stop | Loading and limited brand decoration |

These ranges are guidance, not permission to retime existing consumers. Global
values remain in `tokens.less` (`--gf-motion-fast`, `--gf-motion-base`), and
`mixins.less` retains `.gf-interactive-motion()`. Spotlight's 520ms animation /
560ms cleanup, Search's 420ms animation / 380ms + 80ms cleanup and existing 500ms
theme transitions retain their Normal behavior in A.

Prefer opacity and, when meaningful, 4–12px of displacement for entry. Existing
`cubic-bezier(0.22, 1, 0.36, 1)` deceleration can be reused. Exit direction must
match the action; business state must not wait for decoration. Hover normally
changes background, border or text rather than lifting/scaling every card.
Keep focus outlines and selected controls immediately discernible.

Tailwind owns structure and responsiveness; the existing Less owner owns motion
appearance. Do not add a second token system, global motion plugin/store, generic
carousel framework, default page fly-ins, typewriter text, exaggerated springs,
global route slides or scroll reveal that hides primary first-paint content.

## Reduced Motion is a live setting

Support initial `reduce`, `no-preference → reduce` during work, and restoration
to `no-preference` without reload. CSS alone cannot settle a JS state machine.

| Consumer | Reduced behavior |
| --- | --- |
| Hover / focus | Immediate or very short static feedback, retaining the focus outline |
| Manual paging | One correct selected target, no residual tracks or artificial timer wait; controls remain usable |
| Autoplay | Pause automatic advancement; manual navigation remains available |
| Loading / skeleton / spinner | Visible, identifiable static waiting state until the real work finishes |
| Nav Home scrolling | Native user-controlled scrolling in both modes; content and Footer are SSR-rendered, with no Reveal lock or automatic positioning (#145-B) |
| Decorative loops | Stop motion without removing necessary information |
| Charts | Preserve real data, missing/unknown and loading semantics; no invented transition data |

Cancel visual timers/RAFs and invalidate deferred completion work when settling
or unmounting. Use the latest selected target. Search must distinguish visual
completion from pending network data: retain static loading while pending, then
show only the result accepted by its existing request owner. Returning to Normal
affects the next interaction and must not revive old work. Changing the preference
must not initiate or reposition Home scrolling.

Use narrowly scoped Reduced rules in the existing stylesheet. Do not use a
global `*` animation reset, hide a spinner without waiting feedback, or set zero
duration while leaving the original completion timer in charge. A small helper
is justified only by multiple real consumers: `useReducedMotion` owns a mounted,
disposable media-query subscription for Spotlight and Search, not their paging.

## Performance, SSR and accessibility

Prefer transform/opacity over per-frame layout, scrolling or large blur/filter
work. Do not permanently promote large collections with `will-change`.
SSR must present real accessible content or deterministic loading; motion cannot
replace hydration, image decode or lazy-load eligibility. Every media listener,
timeout, RAF, observer and scroll listener needs an owner and teardown. Deferred
callbacks must not mutate a replacement page, filter or unmounted component.

Preserve real links, keyboard operation, focus, `aria-hidden` and `inert`
semantics. Visual hiding alone does not make offscreen content inaccessible.
Tests use the existing Unit/Nuxt/Browser owners, real component lifecycles and
same-page `emulateMedia` changes. Fake timers belong in isolated unit/Nuxt cases;
Browser evidence uses observable readiness, strict request/error accounting and
no fixed sleeps. Local fixtures and review screenshots are not production or
maintainer visual acceptance. See the [A audit and evidence](../acceptance/139-a-motion-foundation.md).

## User-triggered theme reveal (#150)

NavBar alone wraps its existing Theme Store update with a same-document native
View Transition. The new root snapshot expands from the actual button center
to the farthest viewport corner in 420ms, using `cubic-bezier(0.22, 1, 0.36, 1)`
in both directions. This does not enable cross-page transitions. Initialization,
`html.dark`, persistence and theme ownership remain unchanged.

`useThemeReveal` / `themeReveal.ts` own the disposable visual lifecycle, ignore
overlapping clicks and apply the selected theme once even if capture is skipped.
Reduced Motion, hidden pages, unsupported APIs and invalid origins use the direct
path; Reduced also finishes theme-color transitions immediately. Live Reduced,
resize, route change, page exit and unmount settle active work.
Native `ready` / `finished` and WAAPI completion own cleanup, with no timer.
Before the new capture, only color/surface CSS transitions are finished; transform,
opacity and CSS animation contracts remain intact. Error artwork alone suppresses
its theme-layer fade while the temporary marker exists. Normal direct fallback
retains the existing 500ms theme transitions.

The lifecycle unit owner is `themeReveal.test.ts`; native Browser evidence stays
with Nav Shell, Error Experience, Games Home and Insights Navigation. Their strict
network/error ledgers remain in force. Fixture evidence is not human visual approval.
