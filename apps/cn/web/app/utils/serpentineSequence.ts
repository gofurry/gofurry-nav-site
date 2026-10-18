import type { CSSProperties } from 'vue'

/** Three columns, alternating direction; DOM order always remains chronological. */
export function compactPlacement(index: number): CSSProperties {
  const position = index % 6
  return { '--timeline-row': Math.floor(index / 3) + 1, '--timeline-column': position < 3 ? position + 1 : 6 - position } as CSSProperties
}
export function connector(index: number, length: number) {
  if (index === length - 1) return 'none'
  const position = index % 6
  return position < 2 ? 'right' : position === 3 || position === 4 ? 'left' : 'down'
}
