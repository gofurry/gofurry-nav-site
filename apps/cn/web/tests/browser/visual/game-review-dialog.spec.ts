import { test, expect, reviewDraft, assertReviewAppearance, assertReviewFeedback, assertReviewFocus } from '../fixtures/game-review-dialog'

for (const theme of ['light', 'dark'] as const) {
  for (const device of ['desktop', 'mobile'] as const) {
    test(`Shared review default ${theme} ${device}`, async ({ review }) => {
      const scene = await review.open({ theme, width: device === 'desktop' ? 1440 : 390 })
      await scene.openDialog()
      await assertReviewAppearance(scene)
      await expect(scene.name).toHaveValue(''); await expect(scene.content).toHaveValue('')
      await expect(scene.score).toHaveValue('0'); await expect(scene.feedback).toHaveCount(0)
      const clip = await scene.visualClip()
      await expect(scene.page).toHaveScreenshot(`game-review-dialog-${theme}-${device}.png`, { clip })
      scene.assertQuiet()
    })
  }
}

for (const config of [{ theme: 'light', device: 'desktop', width: 1440 }, { theme: 'dark', device: 'mobile', width: 390 }] as const) {
  test(`Shared review validation ${config.theme} ${config.device}`, async ({ review }) => {
    const scene = await review.open(config)
    await scene.openDialog(); await scene.submit.click()
    await assertReviewFeedback(scene, scene.text.required)
    await assertReviewFocus(scene)
    const clip = await scene.visualClip(true)
    await expect(scene.page).toHaveScreenshot(`game-review-dialog-validation-${config.theme}-${config.device}.png`, { clip })
    scene.assertQuiet()
  })
}

test('Shared review pending dark desktop', async ({ review }) => {
  const scene = await review.open({ theme: 'dark' })
  await scene.openDialog(); await scene.fill()
  const gate = scene.queue()
  await scene.submit.click()
  expect((await gate.waitReceived()).body).toEqual({ id: '91', ...reviewDraft, score: 4.3 })
  await expect(scene.submit).toBeDisabled(); await expect(scene.submit).toHaveText(scene.text.pending)
  await expect(scene.submit).toHaveCSS('opacity', '0.56')
  const clip = await scene.visualClip()
  await expect(scene.page).toHaveScreenshot('game-review-dialog-pending-dark-desktop.png', { clip })
  scene.assertQuiet()
  // The fixture's finally releases this gate even if screenshot comparison fails.
})

test('Shared review success light desktop', async ({ review }) => {
  const scene = await review.open()
  await scene.openDialog(); await scene.fill()
  const gate = scene.queue()
  await scene.submit.click(); await gate.waitReceived(); gate.release()
  await assertReviewFeedback(scene, scene.text.success, true)
  await expect(scene.submit).toBeEnabled()
  await expect(scene.name).toHaveValue(reviewDraft.name)
  await expect(scene.content).toHaveValue(reviewDraft.content)
  const clip = await scene.visualClip()
  await expect(scene.page).toHaveScreenshot('game-review-dialog-success-light-desktop.png', { clip })
  scene.assertQuiet()
})
