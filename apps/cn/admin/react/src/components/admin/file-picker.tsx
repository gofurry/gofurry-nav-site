import { FolderOpen } from '@phosphor-icons/react'
import { useId, useRef } from 'react'
import { Button } from '../ui/button'

export function FilePicker({ label, file, help, accept, disabled, onSelect }: {
  label: string; file: File | null; help?: string; accept?: string; disabled?: boolean; onSelect: (file: File | null) => void
}) {
  const labelId = useId()
  const input = useRef<HTMLInputElement>(null)
  return <div className="min-w-0 space-y-1" data-file-picker>
    <div className="flex min-w-0 items-center gap-2">
      <span id={labelId} className="shrink-0 text-sm font-medium">{label}</span>
      <input ref={input} type="file" className="hidden" aria-labelledby={labelId} accept={accept} disabled={disabled} onChange={(event) => {
        const selected = event.target.files?.[0]
        if (selected) onSelect(selected)
        event.target.value = '' // Allow selecting the same file after a failed upload.
      }} />
      <Button type="button" variant="secondary" size="icon" className="size-8 shrink-0" aria-label="浏览文件" title="选择文件" disabled={disabled} onClick={() => input.current?.click()}><FolderOpen className="size-4" /></Button>
      <span className="min-w-0 truncate text-xs text-muted-foreground" title={file?.name}>{file?.name ?? '未选择文件'}</span>
    </div>
    {help && <p className="text-xs text-muted-foreground">{help}</p>}
  </div>
}
