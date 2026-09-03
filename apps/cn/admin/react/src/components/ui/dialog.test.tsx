import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ConfirmAction, Dialog } from './dialog'

describe('shared dialog viewport contract', () => {
  it('portals a centered, viewport-safe popup with an internal scroll area', () => {
    render(<Dialog open onOpenChange={() => undefined} title="全局搜索"><div>内容</div></Dialog>)

    const popup = screen.getByRole('dialog')
    expect(document.body).toContainElement(popup)
    expect(popup).toHaveAttribute('data-admin-dialog')
    expect(popup.className).toContain('fixed')
    expect(popup.className).toContain('left-1/2')
    expect(popup.className).toContain('top-1/2')
    expect(popup.className).toContain('max-h-[calc(100dvh-2rem)]')
    expect(popup.querySelector('[data-admin-dialog-scroll]')).toHaveClass('admin-scroll', 'overflow-y-auto')
  })

  it('keeps destructive confirmations on the same shared primitive', () => {
    render(<ConfirmAction open onOpenChange={() => undefined} title="删除记录" description="确认目标" onConfirm={() => undefined} />)
    expect(screen.getByRole('dialog')).toHaveAttribute('data-admin-dialog')
    expect(screen.getByRole('button', { name: '确认删除' })).toBeInTheDocument()
  })
})
