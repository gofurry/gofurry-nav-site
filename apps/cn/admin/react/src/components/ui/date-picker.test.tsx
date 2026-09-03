import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { DatePicker, DateTimePicker, calendarMonthDays } from './date-picker'
import { Select } from './select'

afterEach(cleanup)

describe('Admin form controls', () => {
  it('uses a Base UI select instead of a native select', () => {
    const { container } = render(<Select value="nav" onValueChange={() => undefined} options={[{ value: 'nav', label: 'Nav' }]} ariaLabel="Domain" />)
    expect(screen.getByRole('combobox', { name: 'Domain' })).toBeInTheDocument()
    expect(container.querySelector('select')).toBeNull()
  })

  it('opens the shared select below and left-aligned with its trigger', async () => {
    render(<Select value="nav" onValueChange={() => undefined} options={[{ value: 'nav', label: 'Nav' }, { value: 'game', label: 'Game' }]} ariaLabel="Domain" />)
    await userEvent.click(screen.getByRole('combobox', { name: 'Domain' }))
    const popup = await screen.findByRole('listbox')
    expect(popup.parentElement).toHaveAttribute('data-side', 'bottom')
    expect(popup.parentElement).toHaveAttribute('data-align', 'start')
  })

  it('selects calendar dates without native date inputs', async () => {
    const onChange = vi.fn()
    const { container } = render(<DatePicker value="2026-09-03" onValueChange={onChange} ariaLabel="开始日期" />)
    expect(container.querySelector('input[type="date"]')).toBeNull()
    await userEvent.click(screen.getByRole('button', { name: '开始日期' }))
    await userEvent.click(screen.getByRole('button', { name: '15' }))
    expect(onChange).toHaveBeenCalledWith('2026-09-15')
  })

  it('provides a reusable date-time trigger without datetime-local', () => {
    const { container } = render(<DateTimePicker value="2026-09-03T08:30" onValueChange={() => undefined} ariaLabel="执行时间" />)
    expect(screen.getByRole('button', { name: '执行时间' })).toHaveTextContent('2026-09-03 08:30')
    expect(container.querySelector('input[type="datetime-local"]')).toBeNull()
    expect(calendarMonthDays(new Date(2026, 8, 1))).toHaveLength(42)
  })
})
