import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { DatePicker, DateTimePicker, calendarMonthDays } from './date-picker'
import { Select } from './select'

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

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

  it('initializes an empty date-time draft from local time without committing on open', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<DateTimePicker value="" onValueChange={onChange} ariaLabel="执行时间" now={() => new Date(2026, 8, 5, 10, 47)} />)

    await user.click(screen.getByRole('button', { name: '执行时间' }))
    expect(screen.getByRole('combobox', { name: '小时' })).toHaveTextContent('10')
    expect(screen.getByRole('combobox', { name: '分钟' })).toHaveTextContent('47')
    expect(onChange).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: '确定' }))
    expect(onChange).toHaveBeenCalledOnce()
    expect(onChange).toHaveBeenCalledWith('2026-09-05T10:47')
  })

  it('keeps an existing date-time value when the picker opens', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<DateTimePicker value="2026-02-03T08:30" onValueChange={onChange} ariaLabel="执行时间" now={() => new Date(2026, 8, 5, 10, 47)} />)

    await user.click(screen.getByRole('button', { name: '执行时间' }))
    expect(screen.getByRole('combobox', { name: '小时' })).toHaveTextContent('08')
    expect(screen.getByRole('combobox', { name: '分钟' })).toHaveTextContent('30')
    await user.click(screen.getByRole('button', { name: '确定' }))
    expect(onChange).toHaveBeenCalledWith('2026-02-03T08:30')
  })

  it('adjusts hours and minutes by one', async () => {
    const onChange = vi.fn()
    render(<DateTimePicker value="2026-09-03T08:30" onValueChange={onChange} ariaLabel="执行时间" />)

    await userEvent.click(screen.getByRole('button', { name: '执行时间' }))
    await userEvent.click(screen.getByRole('button', { name: '小时加一' }))
    await userEvent.click(screen.getByRole('button', { name: '分钟减一' }))
    expect(screen.getByRole('combobox', { name: '小时' })).toHaveTextContent('09')
    expect(screen.getByRole('combobox', { name: '分钟' })).toHaveTextContent('29')
    await userEvent.click(screen.getByRole('button', { name: '小时减一' }))
    await userEvent.click(screen.getByRole('button', { name: '分钟加一' }))
    await userEvent.click(screen.getByRole('button', { name: '确定' }))
    expect(onChange).toHaveBeenCalledWith('2026-09-03T08:30')
  })

  it('clamps hour and minute adjustments at their lower and upper bounds', async () => {
    const lowerChange = vi.fn()
    render(<DateTimePicker value="2026-09-03T00:00" onValueChange={lowerChange} ariaLabel="下界时间" />)
    await userEvent.click(screen.getByRole('button', { name: '下界时间' }))
    await userEvent.click(screen.getByRole('button', { name: '小时减一' }))
    await userEvent.click(screen.getByRole('button', { name: '分钟减一' }))
    await userEvent.click(screen.getByRole('button', { name: '确定' }))
    expect(lowerChange).toHaveBeenCalledWith('2026-09-03T00:00')

    cleanup()
    const upperChange = vi.fn()
    render(<DateTimePicker value="2026-09-03T23:59" onValueChange={upperChange} ariaLabel="上界时间" />)
    await userEvent.click(screen.getByRole('button', { name: '上界时间' }))
    await userEvent.click(screen.getByRole('button', { name: '小时加一' }))
    await userEvent.click(screen.getByRole('button', { name: '分钟加一' }))
    await userEvent.click(screen.getByRole('button', { name: '确定' }))
    expect(upperChange).toHaveBeenCalledWith('2026-09-03T23:59')
  })
})
