import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PageHeader, PageLayout } from './page'

describe('PageHeader', () => {
  it('renders only the title and actions in the shared Admin header', () => {
    render(<PageLayout><PageHeader title="评论" eyebrow="game.comments" description="开发者说明" actions={<button type="button">新增评论</button>} /><div>正文</div></PageLayout>)

    expect(screen.getByRole('heading', { name: '评论' }).closest('[data-admin-page]')).toHaveClass('gap-4')
    expect(screen.getByRole('heading', { name: '评论' }).closest('[data-admin-page-header]')).not.toHaveClass('-mb-2')
    expect(screen.getByRole('button', { name: '新增评论' })).toBeInTheDocument()
    expect(screen.queryByText('game.comments')).not.toBeInTheDocument()
    expect(screen.queryByText('开发者说明')).not.toBeInTheDocument()
  })
})
