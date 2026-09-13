import type { CSSProperties } from 'react'

export type PatternAppearance = { light_color: string; dark_color: string; light_opacity: number; dark_opacity: number; default_size_px: number }

export function patternMaskStyle(url: string, color: string, opacity: number, size: number): CSSProperties {
  return { backgroundColor: color, opacity, maskImage: `url("${url}")`, WebkitMaskImage: `url("${url}")`, maskRepeat: 'repeat', WebkitMaskRepeat: 'repeat', maskSize: `${size}px`, WebkitMaskSize: `${size}px` }
}

export function PatternPreview({ url, appearance }: { url: string; appearance: PatternAppearance }) {
  return <div className="grid gap-3 sm:grid-cols-2" aria-label="背景亮暗预览">{(['light', 'dark'] as const).map((theme) => <div key={theme} className="relative h-44 overflow-hidden rounded-md border" style={{ background: theme === 'light' ? '#f6eddf' : '#211c19' }}>
    <div data-pattern-mask={theme} className="absolute inset-0" style={patternMaskStyle(url, appearance[`${theme}_color`], appearance[`${theme}_opacity`], appearance.default_size_px)} />
    <span className="absolute left-3 top-3 rounded bg-background px-2 py-1 text-xs">{theme === 'light' ? '浅色主题' : '深色主题'}</span>
  </div>)}</div>
}
