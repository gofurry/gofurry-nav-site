import { execFileSync } from 'node:child_process'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { parse } from '@vue/compiler-sfc'
import ts from 'typescript'
import { extractCssFacts } from './css.mjs'

// Same source boundary as P0, also including new, not-yet-staged source files.
export function isPolicySource(file) {
  return file.startsWith('app/') && /\.(vue|ts|js|css|less)$/.test(file)
    && file !== 'app/assets/js/china.js'
    && !/(?:^|\/)(?:tests|__tests__|fixtures|generated)(?:\/|$)|\.(?:test|spec|d)\.(?:ts|js)$/.test(file)
}

export function discoverSources(root) {
  const tracked = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '-z', '--', 'app'], { cwd: root, encoding: 'utf8' })
  const deleted = new Set(execFileSync('git', ['ls-files', '--deleted', '-z', '--', 'app'], { cwd: root, encoding: 'utf8' }).split('\0'))
  return [...new Set(tracked.split('\0'))].filter(file => isPolicySource(file) && !deleted.has(file)).sort()
}

const visualProperty = /^(?:color|background(?:Color|Image)?|border(?:Top|Right|Bottom|Left)?Color|shadowColor|fill|stroke|(?:light|dark)_color|stop-color|flood-color|lighting-color)$/
const visualGroup = /^(?:colors|palette|textStyle|lineStyle|areaStyle|itemStyle)$/
const wrappers = new Set(['computed', 'ref', 'shallowRef', 'reactive', 'shallowReactive', 'readonly', 'unref', 'toValue'])
const propertyName = node => node && (ts.isIdentifier(node) || ts.isStringLiteral(node) || ts.isNumericLiteral(node)) ? node.text : undefined

// Local, bounded dataflow. Only class/style sinks and named visual objects are
// roots; unrelated messages/imported functions are never interpreted as styles.
class LocalFacts {
  constructor(file, source) {
    this.file = file
    this.source = source
    this.facts = []
    this.seen = new Set()
    this.scopes = new WeakMap()
    this.metadata = new WeakMap()
    this.top = { values: new Map(), parent: null }
    this.unresolved = new Set()
  }

  sourceFile(code, offset, scope = this.top) {
    const ast = ts.createSourceFile(this.file + '.ts', code, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
    if (ast.parseDiagnostics.length) {
      const error = ast.parseDiagnostics[0]
      throw new Error(`${this.file}:${this.line(offset + (error.start || 0))}: ${ts.flattenDiagnosticMessageText(error.messageText, ' ')}`)
    }
    this.metadata.set(ast, { offset })
    const visit = (node, current) => {
      if (ts.isFunctionDeclaration(node) && node.name) current.values.set(node.name.text, node)
      if (ts.isFunctionLike(node) || ts.isBlock(node)) current = { values: new Map(), parent: current }
      if (ts.isParameter(node) && ts.isIdentifier(node.name)) current.values.set(node.name.text, null)
      if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)) current.values.set(node.name.text, node.initializer || null)
      this.scopes.set(node, current)
      ts.forEachChild(node, child => visit(child, current))
    }
    visit(ast, scope)
    return ast
  }

  expression(code, offset, scope = this.top) {
    const ast = this.sourceFile(`(${code})`, offset - 1, scope)
    return ast.statements[0]?.expression
  }

  line(offset) { return this.source.slice(0, offset).split('\n').length }
  offset(node) { return this.metadata.get(node.getSourceFile()).offset + node.getStart() }
  lookup(name, node) {
    for (let scope = this.scopes.get(node) || this.top; scope; scope = scope.parent) {
      if (scope.values.has(name)) return scope.values.get(name)
    }
    return undefined
  }

  emit(kind, value, offset) {
    const key = `${kind}:${offset}:${value}`
    if (this.seen.has(key)) return
    this.seen.add(key)
    this.facts.push({ kind, file: this.file, value, line: this.line(offset), offset })
  }

  returns(fn) {
    if (!fn.body) return []
    if (!ts.isBlock(fn.body)) return [fn.body]
    const out = []
    const visit = node => {
      if (ts.isReturnStatement(node) && node.expression) out.push(node.expression)
      else if (!ts.isFunctionLike(node)) ts.forEachChild(node, visit)
    }
    visit(fn.body)
    return out
  }

  isRef(node, trail = new Set()) {
    if (!node || trail.has(node) || trail.size > 32) return false
    const next = new Set(trail).add(node)
    if (ts.isIdentifier(node)) return this.isRef(this.lookup(node.text, node), next)
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isNonNullExpression(node)) return this.isRef(node.expression, next)
    return ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ['ref', 'shallowRef', 'computed'].includes(node.expression.text)
  }

  values(node, trail = new Set()) {
    if (!node) return []
    if (trail.has(node) || trail.size > 32) {
      this.unresolved.add('recursive/deep local dataflow')
      return []
    }
    const next = new Set(trail).add(node)
    const resolve = value => this.values(value, next)
    if (node.elementOf) return resolve(node.elementOf).flatMap(value => ts.isArrayLiteralExpression(value) ? value.elements.flatMap(resolve) : [value])
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isSatisfiesExpression(node) || ts.isNonNullExpression(node) || ts.isTypeAssertionExpression(node)) return resolve(node.expression)
    if (ts.isIdentifier(node)) return resolve(this.lookup(node.text, node))
    if (ts.isConditionalExpression(node)) return [...resolve(node.whenTrue), ...resolve(node.whenFalse)]
    // A pure interpolation aliases its source; it does not author another class.
    if (ts.isTemplateExpression(node) && node.head.text === '' && node.templateSpans.length === 1 && node.templateSpans[0].literal.text === '') return resolve(node.templateSpans[0].expression)
    if (ts.isBinaryExpression(node)) {
      const op = node.operatorToken.kind
      if ([ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(op)) {
        return op === ts.SyntaxKind.AmpersandAmpersandToken ? resolve(node.right) : [...resolve(node.left), ...resolve(node.right)]
      }
    }
    if (ts.isArrowFunction(node) || ts.isFunctionExpression(node) || ts.isFunctionDeclaration(node)) return this.returns(node).flatMap(resolve)
    if (ts.isCallExpression(node)) {
      if (ts.isIdentifier(node.expression) && wrappers.has(node.expression.text)) return node.arguments.slice(0, 1).flatMap(resolve)
      const callee = ts.isIdentifier(node.expression) ? this.lookup(node.expression.text, node.expression) : undefined
      if (callee) return resolve(callee)
      this.unresolved.add('external/dynamic call')
      return []
    }
    if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
      const keys = ts.isPropertyAccessExpression(node) ? [node.name.text]
        : resolve(node.argumentExpression).map(value => ts.isStringLiteral(value) || ts.isNumericLiteral(value) ? value.text : undefined).filter(value => value !== undefined)
      const key = keys.length === 1 ? keys[0] : undefined
      const bases = resolve(node.expression)
      return bases.flatMap(base => {
        if (key === 'value' && this.isRef(node.expression)) return [base]
        if (ts.isArrayLiteralExpression(base)) {
          const index = key !== undefined && /^\d+$/.test(key) ? Number(key) : undefined
          return (index === undefined ? base.elements : [base.elements[index]]).flatMap(resolve)
        }
        if (ts.isObjectLiteralExpression(base)) return base.properties.flatMap(prop => {
          if (ts.isSpreadAssignment(prop)) return resolve(prop.expression).flatMap(value => this.objectValues(value, key, resolve))
          return key === undefined || propertyName(prop.name) === key ? resolve(prop.initializer || (ts.isShorthandPropertyAssignment(prop) ? prop.name : undefined)) : []
        })
        return []
      })
    }
    return [node]
  }

  objectValues(node, key, resolve) {
    if (!ts.isObjectLiteralExpression(node)) return []
    return node.properties.flatMap(prop => key === undefined || propertyName(prop.name) === key ? resolve(prop.initializer) : [])
  }

  literal(node, trail = new Set()) {
    if (trail.has(node) || trail.size > 32) { this.unresolved.add('recursive/deep template literal'); return undefined }
    const next = new Set(trail).add(node)
    if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text
    if (ts.isTemplateExpression(node)) {
      let result = node.head.text
      for (const span of node.templateSpans) {
        const values = this.values(span.expression)
        const texts = [...new Set(values.map(value => ts.isNumericLiteral(value) ? value.text : this.literal(value, next)).filter(value => value !== undefined))]
        if (texts.length !== 1) { this.unresolved.add('dynamic template literal'); return undefined }
        result += texts[0] + span.literal.text
      }
      return result
    }
    return undefined
  }

  extract(node, kind, trail = new Set()) {
    for (const value of this.values(node)) {
      if (trail.has(value)) continue
      const next = new Set(trail).add(value)
      const literal = this.literal(value)
      if (literal !== undefined) this.emit(kind, literal, this.offset(value) + 1)
      else if (ts.isArrayLiteralExpression(value)) value.elements.forEach(item => this.extract(ts.isSpreadElement(item) ? item.expression : item, kind, next))
      else if (ts.isObjectLiteralExpression(value)) {
        for (const prop of value.properties) {
          if (ts.isSpreadAssignment(prop)) this.extract(prop.expression, kind, next)
          else if (kind === 'class') {
            if (ts.isComputedPropertyName(prop.name)) this.extract(prop.name.expression, kind, next)
            else if (propertyName(prop.name)) this.emit(kind, propertyName(prop.name), this.offset(prop.name) + (ts.isStringLiteral(prop.name) ? 1 : 0))
          } else this.extract(prop.initializer || (ts.isShorthandPropertyAssignment(prop) ? prop.name : undefined), kind, next)
        }
      }
    }
  }

  scriptVisuals(ast) {
    const visit = node => {
      if (ts.isPropertyAssignment(node) && (visualProperty.test(propertyName(node.name)) || visualGroup.test(propertyName(node.name)))) this.extract(node.initializer, 'visual-value')
      if (ts.isPropertyAssignment(node) && propertyName(node.name) === 'innerHTML') this.embeddedStyles(node.initializer)
      if (ts.isVariableDeclaration(node) && visualGroup.test(propertyName(node.name))) this.extract(node.initializer, 'visual-value')
      if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(node.left) && visualProperty.test(node.left.name.text)) this.extract(node.right, 'visual-value')
      if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(node.left) && node.left.name.text === 'className') this.extract(node.right, 'class')
      if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(node.left) && node.left.name.text === 'innerHTML') this.embeddedStyles(node.right)
      if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(node.left) && node.left.name.text === 'cssText') this.styleBinding(node.right)
      if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression)) {
        const method = node.expression.name.text
        const receiver = node.expression.expression
        if (['add', 'remove', 'toggle', 'replace'].includes(method) && ts.isPropertyAccessExpression(receiver) && receiver.name.text === 'classList') node.arguments.forEach(arg => this.extract(arg, 'class'))
        if (method === 'insertAdjacentHTML') this.embeddedStyles(node.arguments[1])
        if (method === 'setAttribute' && node.arguments[0] && ts.isStringLiteral(node.arguments[0])) {
          const name = node.arguments[0].text
          if (name === 'class') this.extract(node.arguments[1], 'class')
          if (name === 'style') this.styleBinding(node.arguments[1])
          if (['fill', 'stroke'].includes(name)) this.extract(node.arguments[1], 'visual-value')
        }
      }
      ts.forEachChild(node, visit)
    }
    visit(ast)
  }

  embeddedStyles(node) {
    for (const value of this.values(node)) {
      const literal = this.literal(value)
      if (literal === undefined) continue
      for (const match of literal.matchAll(/<style(?:\s[^>]*)?>([\s\S]*?)<\/style>/gi)) {
        const offset = this.offset(value) + 1 + match.index + match[0].indexOf(match[1])
        const key = `embedded:${offset}`
        if (this.seen.has(key)) continue
        this.seen.add(key)
        this.facts.push(...extractCssFacts(match[1], { file: this.file, line: this.line(offset) }))
      }
    }
  }

  styleBinding(node, trail = new Set()) {
    for (const value of this.values(node)) {
      if (trail.has(value)) { this.unresolved.add('recursive style array'); continue }
      const next = new Set(trail).add(value)
      if (ts.isArrayLiteralExpression(value)) {
        value.elements.forEach(element => this.styleBinding(element, next))
        continue
      }
      const literal = this.literal(value)
      if (literal === undefined) { this.extract(value, 'visual-value'); continue }
      const offset = this.offset(value) + 1
      const key = `inline:${offset}`
      if (this.seen.has(key)) continue
      this.seen.add(key)
      this.facts.push(...extractCssFacts(`.inline{${literal}}`, { file: this.file, line: this.line(offset) }))
    }
  }

  template(ast) {
    const visit = (node, parentScope) => {
      let scope = parentScope
      const loop = node.props?.find(prop => prop.type === 7 && prop.name === 'for' && prop.exp)
      if (loop) {
        const match = loop.exp.content.match(/^\s*(?:\(([^)]+)\)|([\w$]+))\s+(?:in|of)\s+([\s\S]+)$/)
        if (match) {
          const expression = match[3]
          const collection = this.expression(expression, loop.exp.loc.start.offset + loop.exp.content.lastIndexOf(expression), scope)
          scope = { values: new Map(), parent: scope }
          const names = (match[1] || match[2]).split(',').map(name => name.trim())
          scope.values.set(names[0], { elementOf: collection })
          for (const name of names.slice(1)) scope.values.set(name, null)
        } else this.unresolved.add('dynamic v-for binding')
      }
      for (const prop of node.props || []) {
        if (prop.type === 7 && prop.name === 'html' && prop.exp) this.embeddedStyles(this.expression(prop.exp.content, prop.exp.loc.start.offset, scope))
        const name = prop.type === 6 ? prop.name : prop.type === 7 && prop.name === 'bind' ? prop.arg?.content : undefined
        const kind = /(?:^|-)class$/.test(name || '') ? 'class' : /^(?:style|fill|stroke|stop-color|flood-color|lighting-color)$/.test(name || '') ? 'visual-value' : undefined
        if (!kind) continue
        if (prop.type === 6 && prop.value) {
          if (name === 'style') this.facts.push(...extractCssFacts(`.inline{${prop.value.content}}`, { file: this.file, line: prop.value.loc.start.line }))
          else this.emit(kind, prop.value.content, prop.value.loc.start.offset + 1)
        }
        if (prop.type === 7 && prop.exp) {
          const expression = this.expression(prop.exp.content, prop.exp.loc.start.offset, scope)
          if (name === 'style') this.styleBinding(expression)
          else this.extract(expression, kind)
        }
      }
      for (const child of node.children || []) visit(child, scope)
    }
    visit(ast, this.top)
  }
}

export function extractSource(source, file) {
  const local = new LocalFacts(file, source)
  if (/\.(css|less)$/.test(file)) return { facts: extractCssFacts(source, { file, less: file.endsWith('.less') }), unresolved: [] }
  if (file.endsWith('.vue')) {
    const { descriptor, errors } = parse(source, { filename: file })
    if (errors.length) throw new Error(`${file}: ${errors.map(error => error.message || error).join('; ')}`)
    const scripts = [descriptor.script, descriptor.scriptSetup].filter(Boolean).map(block => {
      if (block.src || (block.lang && !['ts', 'js'].includes(block.lang))) throw new Error(`${file}: unsupported external/script language`)
      return local.sourceFile(block.content, block.loc.start.offset)
    })
    for (const script of scripts) local.scriptVisuals(script)
    if (descriptor.template) {
      if (descriptor.template.src || (descriptor.template.lang && descriptor.template.lang !== 'html')) throw new Error(`${file}: unsupported external/template language`)
      local.template(descriptor.template.ast)
    }
    for (const block of descriptor.styles) {
      if (block.src || (block.lang && !['css', 'less'].includes(block.lang))) throw new Error(`${file}: unsupported external/style language`)
      local.facts.push(...extractCssFacts(block.content, { file, line: block.loc.start.line, less: block.lang === 'less' }))
    }
  } else local.scriptVisuals(local.sourceFile(source, 0))
  return { facts: local.facts, unresolved: [...local.unresolved] }
}

export async function readSourceFacts(root, files = discoverSources(root)) {
  const facts = []
  const unresolved = []
  for (const file of files) {
    const source = await readFile(path.join(root, file), 'utf8')
    const result = extractSource(source, file)
    facts.push(...result.facts)
    unresolved.push(...result.unresolved.map(message => ({ file, message })))
  }
  return { files, facts, unresolved }
}
