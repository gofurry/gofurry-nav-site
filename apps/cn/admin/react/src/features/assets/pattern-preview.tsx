import type { CSSProperties } from 'react'

export type PatternAppearance = { light_color: string; dark_color: string; light_opacity: number; dark_opacity: number; default_size_px: number }

export function patternMaskStyle(url: string, color: string, opacity: number, size: number): CSSProperties {
  return { backgroundColor: color, opacity, maskImage: `url("${url}")`, WebkitMaskImage: `url("${url}")`, maskRepeat: 'repeat', WebkitMaskRepeat: 'repeat', maskSize: `${size}px`, WebkitMaskSize: `${size}px` }
}

export function PatternPreview({ url, appearance, fill = false }: { url: string; appearance: PatternAppearance; fill?: boolean }) {
  return <div className={`grid grid-cols-2 overflow-hidden ${fill ? 'h-full w-full' : 'aspect-video rounded-md border'}`} aria-label="背景亮暗预览">{(['light', 'dark'] as const).map((theme) => <div key={theme} className="relative min-h-0 overflow-hidden" style={{ background: theme === 'light' ? '#f6eddf' : '#211c19' }}>
    <div data-pattern-mask={theme} className="absolute inset-0" style={patternMaskStyle(url, appearance[`${theme}_color`], appearance[`${theme}_opacity`], appearance.default_size_px)} />
    <span className={`absolute bottom-2 left-2 rounded px-1.5 py-0.5 text-[10px] ${theme === 'light' ? 'bg-white/75 text-stone-800' : 'bg-black/40 text-stone-100'}`}>{theme === 'light' ? '浅色' : '深色'}</span>
  </div>)}</div>
}
