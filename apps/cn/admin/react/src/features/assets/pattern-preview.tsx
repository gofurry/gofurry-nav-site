import type { CSSProperties } from 'react'

export type PatternAppearance = { light_color: string; dark_color: string; light_opacity: number; dark_opacity: number; default_size_px: number }

// Browser preflight prevents remote references even before an SVG is uploaded.
// The server independently validates the original bytes with its XML parser.
export function validatePatternSVG(text: string) {
  if (/<!|<\?(?!xml\s)/i.test(text)) throw new Error('SVG 不允许包含声明、实体或处理指令')
  const doc = new DOMParser().parseFromString(text, 'image/svg+xml')
  const allowed = new Set(['svg', 'g', 'path', 'rect', 'circle', 'ellipse', 'line', 'polyline', 'polygon', 'defs', 'use', 'symbol', 'pattern', 'mask', 'clipPath', 'title', 'desc'])
  if (doc.querySelector('parsererror') || doc.documentElement.localName !== 'svg') throw new Error('请选择有效的 SVG 文件')
  for (const node of doc.querySelectorAll('*')) {
    if (!allowed.has(node.localName) || (node.namespaceURI && node.namespaceURI !== 'http://www.w3.org/2000/svg')) throw new Error('SVG 含有不支持或不安全的元素')
    for (const attr of node.attributes) {
      const name = attr.localName.toLowerCase(), value = attr.value.trim()
      if (attr.name === 'xmlns' || attr.name === 'xmlns:xlink') continue
      if (name.startsWith('on') || name === 'style' || name === 'base' || /[\\]|@import/i.test(value) || (/href$/i.test(name) && !/^#[\w-]+$/.test(value)) || /url\s*\(\s*(?!#[\w-]+\s*\))/i.test(value) || value.includes(':')) throw new Error('SVG 不允许脚本、样式或外部资源引用')
    }
  }
}

export function patternMaskStyle(url: string, color: string, opacity: number, size: number): CSSProperties {
  return { backgroundColor: color, opacity, maskImage: `url("${url}")`, WebkitMaskImage: `url("${url}")`, maskRepeat: 'repeat', WebkitMaskRepeat: 'repeat', maskSize: `${size}px`, WebkitMaskSize: `${size}px` }
}

export function PatternPreview({ url, appearance }: { url: string; appearance: PatternAppearance }) {
  return <div className="grid gap-3 sm:grid-cols-2" aria-label="背景亮暗预览">{(['light', 'dark'] as const).map((theme) => <div key={theme} className="relative h-44 overflow-hidden rounded-md border" style={{ background: theme === 'light' ? '#f6eddf' : '#211c19' }}>
    <div data-pattern-mask={theme} className="absolute inset-0" style={patternMaskStyle(url, appearance[`${theme}_color`], appearance[`${theme}_opacity`], appearance.default_size_px)} />
    <span className="absolute left-3 top-3 rounded bg-background px-2 py-1 text-xs">{theme === 'light' ? '浅色主题' : '深色主题'}</span>
  </div>)}</div>
}
