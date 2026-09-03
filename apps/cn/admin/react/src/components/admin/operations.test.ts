import { describe, expect, it } from 'vitest'
import { operationLineSeries } from './operations'

describe('OperationsChart hover behavior', () => {
  it('keeps the line visible while the axis tooltip is active', () => {
    const series = operationLineSeries([{ value: 1 }, { value: 2 }], '#3f6fa8')

    expect(series.data).toEqual([1, 2])
    expect(series.lineStyle).toEqual({ color: '#3f6fa8', width: 2 })
    expect(series.emphasis).toEqual({ disabled: true })
  })
})
