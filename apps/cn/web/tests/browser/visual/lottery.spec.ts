import { test, expect, assertLotteryAppearance, assertLotteryModal, settleLottery } from '../fixtures/lottery'

for (const theme of ['light', 'dark'] as const) for (const device of ['desktop', 'mobile'] as const) {
  const width = device === 'desktop' ? 1440 : 390
  test(`Lottery page ${theme} ${device}`, async ({ lottery }) => {
    await lottery.open({ theme, width })
    await lottery.page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { visibility: hidden; }' })
    await settleLottery(lottery, lottery.root)
    await assertLotteryAppearance(lottery)
    await expect(lottery.page.locator('.lottery-summary__value')).toHaveText(['2', '9'])
    await expect(lottery.page.locator('.lottery-history__row')).toHaveCount(2)
    lottery.assertQuiet()
    // Bounded two-pool/two-history domain root; neither Nav nor Footer is a Lottery golden.
    await expect(lottery.root).toHaveScreenshot(`lottery-page-${theme}-${device}.png`)
    lottery.assertQuiet()
  })
  test(`Lottery join ${theme} ${device}`, async ({ lottery }) => {
    await lottery.open({ theme, width }); await lottery.openModal()
    await expect(lottery.inputs.first()).toBeFocused()
    await expect(lottery.dialog.locator('.lottery-modal__chip')).toHaveCount(5)
    for (const input of await lottery.inputs.all()) await expect(input).toHaveValue('')
    await expect(lottery.dialog.locator('.lottery-modal__message')).toHaveCount(0)
    await settleLottery(lottery, lottery.modal); await assertLotteryModal(lottery)
    for (const control of [...await lottery.inputs.all(), lottery.submit]) await expect(control).toBeInViewport()
    lottery.assertQuiet()
    await expect(lottery.modal).toHaveScreenshot(`lottery-join-${theme}-${device}.png`)
    lottery.assertQuiet()
  })
}

for (const theme of ['light', 'dark'] as const) for (const status of ['success', 'fail'] as const) {
  const device = status === 'success' ? 'desktop' : 'mobile'
  test(`Lottery activation ${status} ${theme} ${device}`, async ({ lottery }) => {
    await lottery.open({ theme, width: device === 'desktop' ? 1440 : 390, activation: status })
    const root = lottery.page.locator('.lottery-activation-page'), card = lottery.page.locator('.activation-card')
    await settleLottery(lottery, root)
    await expect(root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
    await expect(card).toHaveCSS('border-radius', '16px'); await expect(card).toHaveCSS('backdrop-filter', 'blur(18px) saturate(1.12)')
    await expect(card).toHaveCSS('background-color', 'rgba(15, 23, 42, 0.62)'); await expect(card).toHaveCSS('box-shadow', 'rgba(2, 6, 23, 0.34) 0px 24px 70px 0px')
    await expect(lottery.page.locator('.activation-card__title')).toHaveCSS('font-size', status === 'success' ? '36px' : '30px')
    await expect(lottery.page.locator(`.activation-status--${status}`)).toHaveCSS('color', status === 'success' ? 'rgb(134, 239, 172)' : 'rgb(253, 186, 116)')
    await expect(lottery.page.locator('.activation-card__link')).toHaveCSS('font-size', '14px')
    await expect(lottery.page.locator('.activation-card__countdown span')).toHaveText('15')
    lottery.assertQuiet()
    await expect(root).toHaveScreenshot(`lottery-activation-${status}-${theme}-${device}.png`)
    lottery.assertQuiet()
  })
}
