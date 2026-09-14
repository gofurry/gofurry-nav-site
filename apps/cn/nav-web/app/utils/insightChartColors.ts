export function insightChartColors(element: HTMLElement) {
  const style = getComputedStyle(element)
  const token = (name: string) => style.getPropertyValue(name).trim()
  return {
    line: token('--gf-accent'), axis: token('--gf-text-muted'),
    split: token('--gf-border'), tooltip: token('--gf-page-background'),
    tooltipBorder: token('--gf-border-strong'), text: token('--gf-text-main'),
  }
}
