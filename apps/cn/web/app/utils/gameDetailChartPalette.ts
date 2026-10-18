/** Resolve inherited Detail tokens before passing colors to the Canvas renderer. */
export function readGameDetailChartPalette(element: HTMLElement) {
  const style = getComputedStyle(element)
  const read = (role: string) => style.getPropertyValue(`--games-detail-chart-${role}`).trim()
  return {
    playerPeak: read('player-peak'),
    playerAverage: read('player-average'),
    playerArea: read('player-area'),
    priceLine: read('price-line'),
    axis: read('axis'),
    split: read('split'),
    tooltipBackground: read('tooltip-bg'),
    tooltipText: read('tooltip-text'),
    playerTooltipBorder: read('player-tooltip-border'),
    priceTooltipBorder: read('price-tooltip-border'),
  }
}
