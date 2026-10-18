import { test, type SiteVisualScene } from '../fixtures/site-detail-visual'

// Eight representative compositions only; behavior/error matrices remain Functional.
for (const [scene, theme, device] of [
  ['performance', 'light', 'desktop'],
  ['performance', 'dark', 'desktop'],
  ['overview', 'light', 'desktop'],
  ['security-tls', 'light', 'desktop'],
  ['insights', 'light', 'desktop'],
  ['http', 'light', 'desktop'],
  ['performance', 'light', 'mobile'],
  ['similar', 'light', 'mobile'],
] as const satisfies ReadonlyArray<readonly [SiteVisualScene, 'light' | 'dark', 'desktop' | 'mobile']>) {
  test(`Site Detail ${scene} ${theme} ${device}`, async ({ siteVisual }) => {
    await siteVisual.open({ scene, theme, width: device === 'desktop' ? 1440 : 390 })
    await siteVisual.capture(`site-detail-${scene}-${theme}-${device}.png`)
  })
}
