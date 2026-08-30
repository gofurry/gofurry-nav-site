import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { OperationTabs } from '../../components/admin/operations'
import { changesTabs, collectionActionVisibility, metricsTabs } from './capability-ux'

describe('capability-aware operational UX', () => {
  it('lets Operator execute collection but hides schedule controls', () => {
    const capabilities = new Set(['collection.read', 'collection.execute'])
    expect(collectionActionVisibility((capability) => capabilities.has(capability))).toEqual({ runNow: true, manual: true, control: false })
  })

  it('lets Developer control schedules', () => {
    const capabilities = new Set(['collection.execute', 'collection.control'])
    expect(collectionActionVisibility((capability) => capabilities.has(capability))).toEqual({ runNow: true, manual: true, control: true })
  })

  it('hides technical tabs without their native capabilities', () => {
    const { rerender } = render(<OperationTabs tabs={metricsTabs} active="overview" onChange={vi.fn()} can={() => false} />)
    expect(screen.queryByText('技术契约')).not.toBeInTheDocument()
    rerender(<OperationTabs tabs={changesTabs} active="recent" onChange={vi.fn()} can={() => false} />)
    expect(screen.queryByText('技术契约')).not.toBeInTheDocument()
  })

  it('shows technical tabs only through the matching capability', () => {
    render(<OperationTabs tabs={metricsTabs} active="overview" onChange={vi.fn()} can={(capability) => capability === 'metrics.technical'} />)
    expect(screen.getByText('技术契约')).toBeInTheDocument()
  })
})
