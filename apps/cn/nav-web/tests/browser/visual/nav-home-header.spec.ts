import { test, expect } from '../fixtures/nav-home-header'

test('Nav Home Header / desktop', async ({ page, header }) => {
  await header.open()
  await expect(header.quick.locator('.quick-site-tile')).toHaveCount(16)
  await expect(header.quick.locator('.quick-site-manage')).toBeVisible()
  await expect(header.root.locator('.nav-header__scroll-hint')).toBeVisible()
  await expect(header.root).not.toHaveCSS('box-shadow', 'none')
  await expect(page).toHaveScreenshot('nav-home-header-desktop.png', { clip: await header.headerClip() })
})

test('Nav Home suggestions / desktop', async ({ page, header }) => {
  await header.open()
  await header.showSuggestions()
  await expect(page).toHaveScreenshot('nav-home-search-suggestions-desktop.png', { clip: await header.searchClip() })
})

test('Nav Home suggestions / mobile', async ({ page, header }) => {
  await header.open(390)
  await header.showSuggestions()
  await expect(header.quick).toBeHidden()
  await expect(header.input).toHaveCSS('background-color', 'rgba(255, 255, 255, 0.9)')
  await expect(header.input).not.toHaveCSS('box-shadow', 'none')
  await expect(page).toHaveScreenshot('nav-home-search-suggestions-mobile.png', { clip: await header.searchClip() })
})

test('Nav Home Quick Sites validation / desktop', async ({ page, header }) => {
  await header.open()
  await header.quick.getByRole('button', { name: '管理快捷站点', exact: true }).click()
  await expect(header.modal).toBeVisible()
  // Vue's outer opacity transition and inner panel transform finish separately.
  // Reach the settled real modal before inserting validation/focusing its input.
  await expect(page.locator('.quick-modal-backdrop')).not.toHaveClass(/quick-modal-enter/)
  await expect(header.modal).toHaveCSS('transform', 'none')
  await header.settle(page.locator('.quick-modal-backdrop'))
  await expect(header.modal.locator('.quick-modal-item')).toHaveCount(2)
  await header.modal.locator('button[type="submit"]').click()
  await expect(header.modal.locator('.quick-modal-error')).toHaveText('请输入网站名称')
  await header.modal.locator('#custom-site-name').focus()
  await expect(page).toHaveScreenshot('nav-home-quick-sites-modal-desktop.png', { clip: await header.modalClip() })
})
