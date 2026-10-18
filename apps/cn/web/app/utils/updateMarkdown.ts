import MarkdownIt from 'markdown-it'
import sanitizeHtml from 'sanitize-html'

const markdown = new MarkdownIt({ html: false, breaks: true, linkify: true, typographer: false }).disable('table')
markdown.renderer.rules.s_open = () => '<del>'
markdown.renderer.rules.s_close = () => '</del>'
markdown.core.ruler.push('release_heading_levels', state => {
  for (const token of state.tokens) {
    if (token.type === 'heading_open' || token.type === 'heading_close') token.tag = `h${Math.min(6, Number(token.tag.slice(1)) + 1)}`
  }
})

function allowedURL(value: string | undefined, image = false): value is string {
  if (!value || [...value].some(char => char.charCodeAt(0) <= 32 || char.charCodeAt(0) === 127 || char === '\\' || /\s/.test(char))) return false
  if (value.startsWith('/')) return !value.startsWith('//')
  if (!image && value.startsWith('#')) return true
  if (!image && /^mailto:[^/]/i.test(value)) return true
  if (!(image ? /^https:\/\//i : /^https?:\/\//i).test(value)) return false
  try { return Boolean(new URL(value).hostname) } catch { return false }
}

// Final boundary shared by production render and the security fixture.
// Exported for adversarial sanitizer tests; preview components use renderUpdateMarkdown only.
export function sanitizeUpdateMarkdown(html: string): string {
  return sanitizeHtml(html, {
    allowedTags: ['p', 'h2', 'h3', 'h4', 'h5', 'h6', 'strong', 'em', 'del', 'ul', 'ol', 'li', 'blockquote', 'a', 'img', 'code', 'pre', 'hr', 'br'],
    allowedAttributes: { a: ['href', 'target', 'rel'], img: ['src', 'alt', 'title', 'loading', 'decoding', 'referrerpolicy'], code: ['class'] },
    allowedClasses: { code: [/^language-[a-zA-Z0-9_-]+$/] },
    allowedSchemes: ['http', 'https', 'mailto'], allowedSchemesByTag: { img: ['https'] },
    allowProtocolRelative: false, parseStyleAttributes: false,
    transformTags: {
      a: (_tag, attrs): sanitizeHtml.Tag => {
        const href = attrs.href
        if (!allowedURL(href)) return { tagName: 'a', attribs: {} }
        return { tagName: 'a', attribs: /^https?:\/\//i.test(href) ? { href, target: '_blank', rel: 'noopener noreferrer' } : { href } }
      },
      img: (_tag, attrs) => ({ tagName: 'img', attribs: allowedURL(attrs.src, true) ? {
        src: attrs.src, alt: attrs.alt ?? '', ...(attrs.title ? { title: attrs.title } : {}),
        loading: 'lazy', decoding: 'async', referrerpolicy: 'no-referrer',
      } : {} }),
    },
    exclusiveFilter: frame => frame.tag === 'img' && !frame.attribs.src,
  })
}

export function renderUpdateMarkdown(source: string): string {
  return sanitizeUpdateMarkdown(markdown.render(source))
}
