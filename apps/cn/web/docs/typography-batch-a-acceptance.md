# #126 Batch A — Typography family acceptance

Date: 2026-09-29. Scope: Nav Web family activation only. This is a batch evidence
record, not a new typography governance rule or approval of later typography work.

## Implementation and boundaries

- `@fontsource-variable/inter` and `@fontsource-variable/noto-sans-sc` are both
  exactly `5.3.0` in the package and frozen lockfile. No other dependency changes.
- Nuxt loads only their normal `wght.css` entries before the existing global CSS.
  Vite emits 105 local WOFF2 assets; none are italic. Fontsource's existing
  unicode-range declarations are used without a custom subset pipeline.
- The existing `tokens.less` root owns `--gf-font-sans` and `--gf-font-mono`.
  `html/body` consume Sans, identically across locales/themes. Native controls
  retain `font: inherit`.
- Only the four existing Site Detail and two Updates mono declarations now consume
  the Mono token. Mono uses the specified system stack, without another font download.
- Rating's Arial glyph implementation and every existing `font-weight` are unchanged.
  No font size, line height, tracking, spacing, dimensions, transform or negative
  margin was changed. No component has a locale-specific family override.
- Style debt manifest and all 130 accepted PNG hashes are unchanged. Existing
  ambient raw debt (75) and Insights important debt (5) remain intact; deep/legacy
  dark remain zero. No new style-policy/Stylelint typography rule was added.

## Engineering verification

Passed: frozen install, typecheck, lint, stylelint, style policy, policy tests
(75), Unit (357), Nuxt (25), Insights semantics, SEO recovery guard, production
build and all seven existing Browser Smoke owners (30 cases).

Browser/Visual use the existing production Nitro/upstream fixtures and pinned
Linux Playwright 1.60.0 image, Node 24.15.0, Chromium 148.0.7778.96, one worker,
zero retries. No global runner, error-capture or font-readiness changes were made.

## Cold-cache font delivery

The one-time audit ran 14 samples: the seven routes below at 1440 and 390 pixels.
Each used a fresh BrowserContext, disabled cache and bypassed service workers via
CDP **before navigation**. Capture awaited `document.fonts.ready` and two frames.
Resource Timing `transferSize` includes browser-reported header overhead;
`encodedBodySize` is the encoded response-body size. All measured font requests
have positive transfer/body sizes, rather than warm-cache zero-byte entries.

| Fixture route | Desktop requests | Transfer bytes | Encoded bytes |
| --- | ---: | ---: | ---: |
| `/` | 12 | 696,276 | 692,676 |
| `/en` | 4 | 224,912 | 223,712 |
| `/site/41` | 14 | 826,788 | 822,588 |
| `/en/site/41` | 1 | 48,556 | 48,256 |
| `/games` | 14 | 820,240 | 816,040 |
| `/insights` | 13 | 708,644 | 704,744 |
| `/updates` | 12 | 699,356 | 695,756 |

Mobile results match Desktop except `/games`: 13 requests, 758,112 transfer bytes,
754,212 encoded bytes. All sampled English pages are below the 400 KB reference;
all sampled Chinese pages are below 1.5 MB. These are fixture measurements, not a
guarantee for arbitrary production content or a new CI performance budget.
The Home shell fixture uses an empty resource inventory; other owners retain
their existing fixed content. The English Home's existing Chinese search
placeholder also requests CJK chunks; no copy change is part of this batch.

All font URLs use the page's own `127.0.0.1:<fixture-port>` host and `/_nuxt/` asset
path; **third-party font requests = 0**, font preloads = 0. Examples:

```text
http://127.0.0.1:33347/_nuxt/inter-latin-wght-normal.Dx4kXJAl.woff2
http://127.0.0.1:33347/_nuxt/noto-sans-sc-119-wght-normal.BfzSEbFz.woff2
http://127.0.0.1:33347/_nuxt/noto-sans-sc-118-wght-normal.BPEb0gM9.woff2
```

The [static delivery evidence](typography-batch-a-font-delivery.json) records every
URL, resource type, request host, transfer/encoded bytes and sample totals.
Loopback ports identify temporary fixtures, not deployed infrastructure. The
one-time audit script is not committed and creates no permanent font subsystem.

## Visual and layout review

The complete existing Visual compare ran with
`pnpm exec playwright test --config=playwright.visual.config.ts`: **5 passed,
126 failed**, zero skipped/flaky, in 8.6 minutes. All 126 failures are
`toHaveScreenshot` pixel mismatches; there are no other reported test errors or
global runner errors. This is not a Visual compare pass. The requested family
activation changes glyphs and natural text layout against the existing system-font
baselines. No update mode was used and all 130 accepted PNG hashes remain unchanged.
The passing cases are the two Game gallery captures, two Home active BottomTab
captures and the pinned Chromium environment assertion.

The 14 delivery samples have no body horizontal overflow or clipped-control
candidates. Screenshot spot checks also cover mobile Updates, English Site Detail,
Insights, Home, navigation, Preferences and review dialogs, plus desktop Games
and Site Detail Performance/TLS in representative Light/Dark states. No true
clipping, collision or overflow regression was identified in these checks, and
no layout robustness fix was necessary. Natural wrapping, changed glyph widths
and stronger appearance of existing weights are expected family changes, not
compensation targets. The HTML compare report and temporary review captures are
retained locally for maintainer review, outside the committed golden inventory.

Maintainer approval is pending. Review Chinese/English text, the appearance of
existing high weights, mixed-script baselines, mobile labels/tabs and technical
Mono values. This batch does not authorize weight convergence, a typography guard
or a later batch. No remote CI or production rollout is claimed.

## Gate A final family decision (2026-09-29)

After manual comparison, the maintainer rejected Candidate B (static Lato 400/700)
and selected Candidate A: **Manrope Variable + Noto Sans SC Variable**. This
supersedes the initial Inter family choice above; the Batch A measurements remain
historical Inter evidence, not Manrope delivery or Visual acceptance.

The final runtime pins `@fontsource-variable/manrope` to `5.3.0` and loads only
its normal `wght.css`. Noto Sans SC Variable remains at `5.3.0`. Lato's dependency
and CSS imports are removed. Mono, Rating's Arial exception, authored weights,
typography metrics, layout, style policy and Visual goldens remain unchanged.
This decision freezes Gate A's family selection only; it does not authorize a
later typography batch or approve/update Visual baselines.

Restoration verification passed: frozen install, typecheck, lint, stylelint and
production build. The rebuilt output contains five normal Manrope WOFF2 assets
and no Lato font assets. Candidate B had not been committed; the restored runtime
and lockfile exactly match the earlier Candidate A local commit. No Browser or
Visual comparison was rerun for this focused restoration.
