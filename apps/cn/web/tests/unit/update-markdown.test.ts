// @vitest-environment happy-dom
import { describe, expect, it } from 'vitest'
import fixture from '../../../../../contracts/fixtures/update-markdown.json'
import { renderUpdateMarkdown, sanitizeUpdateMarkdown } from '../../app/utils/updateMarkdown'

type ContractCase = { id: string; kind: string; input: string; elements: { selector: string; count?: number; text?: string; attributes?: Record<string, string> }[]; absentSelectors: string[]; textContains?: string }

describe('shared update Markdown contract v1', () => {
  for (const item of fixture.cases as ContractCase[]) it(item.id, () => {
    const output = item.kind === 'html' ? sanitizeUpdateMarkdown(item.input) : renderUpdateMarkdown(item.input)
    const container = document.createElement('div')
    container.innerHTML = output
    for (const rule of item.elements) {
      const elements = container.querySelectorAll(rule.selector)
      expect(elements.length, rule.selector).toBe(rule.count ?? 1)
      if (rule.text !== undefined) expect(elements[0].textContent).toBe(rule.text)
      for (const [key, value] of Object.entries(rule.attributes ?? {})) expect(elements[0].getAttribute(key), key).toBe(value)
    }
    for (const selector of item.absentSelectors) expect(container.querySelector(selector), selector).toBeNull()
    if (item.textContains) expect(container.textContent).toContain(item.textContains)
    const attributes: Record<string, string[]> = { a: ['href', 'target', 'rel'], img: ['src', 'alt', 'title', 'loading', 'decoding', 'referrerpolicy'], code: ['class'] }
    for (const element of container.querySelectorAll('*')) {
      for (const attr of element.attributes) expect(attributes[element.localName] ?? []).toContain(attr.name)
    }
  })
})
