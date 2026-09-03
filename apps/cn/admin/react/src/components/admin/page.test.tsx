import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PageHeader } from './page'

describe('PageHeader', () => {
  it('renders only the title and actions in the shared Admin header', () => {
    render(<PageHeader title="评论" eyebrow="game.comments" description="开发者说明" actions={<button type="button">新增评论</button>} />)

    expect(screen.getByRole('heading', { name: '评论' }).closest('[data-admin-page-header]')).toHaveClass('-mb-2')
    expect(screen.getByRole('button', { name: '新增评论' })).toBeInTheDocument()
    expect(screen.queryByText('game.comments')).not.toBeInTheDocument()
    expect(screen.queryByText('开发者说明')).not.toBeInTheDocument()
  })
})
