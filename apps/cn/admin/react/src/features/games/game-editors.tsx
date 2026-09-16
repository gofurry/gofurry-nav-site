import { useState } from 'react'
import { Check, Plus, Trash2 } from 'lucide-react'
import { z } from 'zod'
import platformIcons from '../../../../../nav-web/app/data/platform-icons.json'
import { FormField } from '../../components/admin/page'
import { Button } from '../../components/ui/button'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { listJSON } from '../../lib/api'
import type { KeyValue, OptionItem } from '../../lib/types'

export const summarySchema = z.string().refine((value) => Array.from(value).length <= 400, '简介最多 400 个字符')

export function SummaryField({ label, value, onChange, error }: { label: string; value: string; onChange: (value: string) => void; error?: string }) {
  const used = Array.from(value).length
  return <FormField label={label} error={error || (used > 400 ? '简介最多 400 个字符' : undefined)}>
    <Textarea aria-label={label} aria-invalid={!!error || used > 400} value={value} onChange={(event) => onChange(event.target.value)} />
    <span className="text-xs text-muted-foreground">已用 {used}/400 字符 · 剩余 {Math.max(0, 400 - used)} 字符</span>
  </FormField>
}

const normalizeKey = (key: string) => key.trim().toLowerCase()
const platforms = Object.keys(platformIcons)

export function KeyValueEditor({ value, onChange, platformKeys = false }: { value: KeyValue[]; onChange: (value: KeyValue[]) => void; platformKeys?: boolean }) {
  const items = value.length ? value : [{ key: '', value: '' }]
  return <div className="grid gap-2">{items.map((item, index) => {
    const available = platforms.filter((key) => !items.some((entry, current) => current !== index && normalizeKey(entry.key) === key))
    const options = [{ value: '', label: '选择平台' }, ...available.map((key) => ({ value: key, label: key }))]
    // Preserve existing spelling and unknown historical values until explicitly changed.
    if (item.key && !options.some((option) => option.value === item.key)) options.push({ value: item.key, label: item.key })
    const setKey = (key: string) => {
      if (platformKeys && key && items.some((entry, current) => current !== index && normalizeKey(entry.key) === normalizeKey(key))) return
      onChange(items.map((entry, current) => current === index ? { ...entry, key } : entry))
    }
    return <div key={index} className="grid grid-cols-[10rem_minmax(0,1fr)_auto] gap-2">
      {platformKeys ? <Select ariaLabel={`平台 ${index + 1}`} value={item.key} onValueChange={setKey} options={options} /> : <Input value={item.key} onChange={(event) => setKey(event.target.value)} placeholder="键" />}
      <Input aria-label={`URL / 值 ${index + 1}`} value={item.value} onChange={(event) => onChange(items.map((entry, current) => current === index ? { ...entry, value: event.target.value } : entry))} placeholder={platformKeys ? 'URL' : '值 / URL'} />
      <Button type="button" variant="ghost" size="icon" aria-label="删除此项" onClick={() => onChange(items.filter((_, current) => current !== index))}><Trash2 className="size-4" /></Button>
    </div>
  })}<Button type="button" variant="secondary" size="sm" className="w-fit" onClick={() => onChange([...items, { key: '', value: '' }])}><Plus className="size-3.5" />新增一项</Button></div>
}

export async function loadAllTagOptions() {
  const options: OptionItem[] = []
  for (let page = 1; ; page++) {
    const result = await listJSON<OptionItem>('/api/v1/options/tags', page, 200)
    options.push(...result.list)
    if (options.length >= result.total || result.list.length === 0) return options
  }
}

export function TagMultiSelect({ options, selected, onChange, loading, error }: { options: OptionItem[]; selected: string[]; onChange: (value: string[]) => void; loading?: boolean; error?: string }) {
  const [search, setSearch] = useState('')
  const keyword = search.trim().toLocaleLowerCase()
  const visible = options.filter((option) => `${option.label} ${option.extra ?? ''} ${option.id}`.toLocaleLowerCase().includes(keyword))
  return <div className="grid gap-2">
    <Input aria-label="搜索全部标签" placeholder="搜索全部标签…" value={search} onChange={(event) => setSearch(event.target.value)} />
    <p className="text-xs text-muted-foreground">已选 {selected.length} 个标签</p>
    {loading ? <p>加载中…</p> : error ? <p className="text-danger">{error}</p> : <div className="grid max-h-64 gap-1 overflow-auto rounded-md border p-2 md:grid-cols-2 xl:grid-cols-3">
      {visible.map((option) => {
        const active = selected.includes(String(option.id))
        return <button key={option.id} type="button" role="checkbox" aria-checked={active} onClick={() => onChange(active ? selected.filter((id) => id !== String(option.id)) : [...selected, String(option.id)])} className="flex items-center gap-2 rounded px-2 py-2 text-sm hover:bg-surface-muted"><span className={active ? 'grid size-4 place-items-center rounded border border-primary bg-primary text-primary-foreground' : 'size-4 rounded border'}>{active && <Check className="size-3" />}</span>{option.label}</button>
      })}
      {!visible.length && <p className="p-2 text-sm text-muted-foreground">无匹配标签</p>}
    </div>}
  </div>
}
