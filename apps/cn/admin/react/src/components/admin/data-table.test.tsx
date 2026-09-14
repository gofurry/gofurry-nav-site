import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { DataTable } from './data-table'

afterEach(cleanup)

describe('DataTable column visibility menu', () => {
  it('shows a check only for visible columns', async () => {
    render(<DataTable
      data={[{ name: 'GoFurry', internal: 'hidden' }]}
      columns={[{ key: 'name', header: '名称' }, { key: 'internal', header: '内部列', hidden: true }]}
      total={1}
      page={1}
      pageSize={20}
      search=""
      onSearchChange={() => undefined}
      onPageChange={() => undefined}
      onPageSizeChange={() => undefined}
      searchable={false}
    />)

    await userEvent.click(screen.getByRole('button', { name: /列/ }))
    const visible = await screen.findByRole('menuitemcheckbox', { name: '名称' })
    const hidden = await screen.findByRole('menuitemcheckbox', { name: '内部列' })
    expect(visible).toHaveAttribute('aria-checked', 'true')
    expect(visible.querySelector('svg')).not.toBeNull()
    expect(hidden).toHaveAttribute('aria-checked', 'false')
    expect(hidden.querySelector('svg')).toBeNull()
  })

  it('does not reserve an invisible label row above the search input', () => {
    render(<DataTable
      data={[]}
      columns={[{ key: 'name', header: '名称' }]}
      total={0}
      page={1}
      pageSize={20}
      search=""
      onSearchChange={() => undefined}
      onPageChange={() => undefined}
      onPageSizeChange={() => undefined}
    />)

    const search = screen.getByRole('textbox', { name: '搜索列表' })
    expect(search).toHaveAttribute('autocomplete', 'off')
    expect(search.closest('label')).toHaveClass('relative', 'block')
    expect(search.closest('label')?.querySelector('span.invisible[aria-hidden="true"]')).toBeNull()
  })

  it('keeps IME composition local until the candidate is committed', () => {
    const onSearchChange = vi.fn()
    render(<DataTable
      data={[]}
      columns={[{ key: 'name', header: '名称' }]}
      total={0}
      page={1}
      pageSize={20}
      search=""
      onSearchChange={onSearchChange}
      onPageChange={() => undefined}
      onPageSizeChange={() => undefined}
    />)

    const search = screen.getByRole('textbox', { name: '搜索列表' })
    fireEvent.compositionStart(search)
    fireEvent.change(search, { target: { value: 'long' } })
    fireEvent.change(search, { target: { value: 'longlong' } })

    expect(search).toHaveValue('longlong')
    expect(onSearchChange).not.toHaveBeenCalled()

    fireEvent.compositionEnd(search, { data: '龙', target: { value: '龙' } })
    fireEvent.change(search, { target: { value: '龙' } })

    expect(search).toHaveValue('龙')
    expect(onSearchChange).toHaveBeenCalledOnce()
    expect(onSearchChange).toHaveBeenCalledWith('龙')
  })
})
