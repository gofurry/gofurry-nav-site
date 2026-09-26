import type { IdeaInput, Kind } from './types'

// Visible separators only. Canonical source keys and duplicate checks belong to Go.
function fields(line: string): string[] {
  const values: string[] = []
  let value = ''
  for (let i = 0; i < line.length; i++) {
    const char = line[i]
    if (char === '\\' && (line[i + 1] === '|' || line[i + 1] === '\\')) value += line[++i]
    else if (char === '|') { values.push(value.trim()); value = '' }
    else value += char
  }
  values.push(value.trim())
  return values
}

function sourceOnly(value: string, kind: Kind) {
  return !/\s/.test(value) && (/^https?:\/\//i.test(value) || /^(?:www\.)?[a-z0-9-]+(?:\.[a-z0-9-]+)+(?:[:/].*)?$/i.test(value) || (kind === 'game' && /^(?:steam:)?\d+$/i.test(value)))
}

export function parsePaste(text: string, kind: Kind): IdeaInput[] {
  const lines = text.replace(/\r\n?/g, '\n').split('\n').map((line, index) => ({ line: line.trim(), number: index + 1 })).filter(({ line }) => line)
  if (!lines.length) throw new Error('请先粘贴内容')
  if (lines.length > 500) throw new Error('每批最多 500 条，请分批加入')
  const errors: string[] = []
  const items: IdeaInput[] = []
  for (const { line, number } of lines) {
    const columns = fields(line)
    let error = ''
    if (line.includes('\t')) error = '不支持 Tab / Excel 分列，请用 | 分隔名称、来源、备注。'
    else if (columns.length === 1 && !sourceOnly(columns[0], kind)) error = '请按“名称 | 来源 | 备注”填写；空格不会分列。仅来源可单独一行。'
    else if (columns.length > 3) error = '最多三列；内容中的竖线请写成 \\|。'
    else if (columns.length > 1 && !columns[0] && !columns[1]) error = '名称和来源至少填写一个。'
    if (error) { errors.push(`第 ${number} 行：${error}`); continue }
    items.push({ kind, title: columns.length === 1 ? '' : columns[0], source: columns.length === 1 ? columns[0] : columns[1], note: columns[2] ?? '', priority: 'normal' })
  }
  if (errors.length) throw new Error(errors.slice(0, 5).join('\n') + (errors.length > 5 ? `\n另有 ${errors.length - 5} 行格式错误。` : ''))
  return items
}
