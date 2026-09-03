import { Menu } from '@base-ui/react/menu'
import { flexRender, getCoreRowModel, getSortedRowModel, useReactTable, type ColumnDef, type SortingState, type VisibilityState } from '@tanstack/react-table'
import { Check, ChevronDown, ChevronLeft, ChevronRight, ChevronsUpDown, Columns3, MoreHorizontal, Pencil, Search, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { getAtPath } from '../../lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Select } from '../ui/select'
import { EmptyState, ErrorState, LoadingState } from './states'

export type AdminColumn<T> = { key: string; header: string; render?: (row: T) => ReactNode; hidden?: boolean; sortable?: boolean }

function RowMenu({ onEdit, onDelete }: { onEdit?: () => void; onDelete?: () => void }) {
  return <Menu.Root>
    <Menu.Trigger render={<Button variant="ghost" size="icon" aria-label="更多操作" />}><MoreHorizontal className="size-4" /></Menu.Trigger>
    <Menu.Portal><Menu.Positioner className="z-40" sideOffset={4} align="end"><Menu.Popup className="min-w-32 rounded-md border bg-surface p-1 shadow-lg outline-none">
      {onEdit && <Menu.Item className="flex cursor-default items-center gap-2 rounded px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-surface-muted" onClick={onEdit}><Pencil className="size-3.5" />编辑</Menu.Item>}
      {onDelete && <Menu.Item className="flex cursor-default items-center gap-2 rounded px-2 py-1.5 text-sm text-danger outline-none data-[highlighted]:bg-danger/10" onClick={onDelete}><Trash2 className="size-3.5" />删除</Menu.Item>}
    </Menu.Popup></Menu.Positioner></Menu.Portal>
  </Menu.Root>
}

export function DataTable<T extends object>({ data, columns, total, page, pageSize, search, onSearchChange, onPageChange, onPageSizeChange, onRowClick, onEdit, onDelete, loading, error, onRetry, searchable = true, searchPlaceholder = '搜索…（按 / 聚焦）', toolbar, title, headerActions }: {
  data: T[]; columns: AdminColumn<T>[]; total: number; page: number; pageSize: number; search: string
  onSearchChange: (value: string) => void; onPageChange: (page: number) => void; onPageSizeChange: (size: number) => void
  onRowClick?: (row: T) => void; onEdit?: (row: T) => void; onDelete?: (row: T) => void
  loading?: boolean; error?: string; onRetry?: () => void
  searchable?: boolean; searchPlaceholder?: string; toolbar?: ReactNode; title?: ReactNode; headerActions?: ReactNode
}) {
  const searchRef = useRef<HTMLInputElement>(null)
  const [sorting, setSorting] = useState<SortingState>([])
  const [visibility, setVisibility] = useState<VisibilityState>(() => Object.fromEntries(columns.filter((column) => column.hidden).map((column) => [column.key, false])))
  const tableColumns = useMemo<ColumnDef<T>[]>(() => [
    ...columns.map((column) => ({ id: column.key, accessorFn: (row: T) => getAtPath(row as Record<string, unknown>, column.key), header: column.header, enableSorting: column.sortable !== false, cell: column.render ? ({ row }: { row: { original: T } }) => column.render!(row.original) : ({ getValue }: { getValue: () => unknown }) => String(getValue() ?? '—') })),
    ...(onEdit || onDelete ? [{ id: '_actions', header: '', enableSorting: false, enableHiding: false, cell: ({ row }: { row: { original: T } }) => <RowMenu onEdit={onEdit ? () => onEdit(row.original) : undefined} onDelete={onDelete ? () => onDelete(row.original) : undefined} /> }] : []),
  ], [columns, onDelete, onEdit])
  const table = useReactTable({ data, columns: tableColumns, state: { sorting, columnVisibility: visibility }, onSortingChange: setSorting, onColumnVisibilityChange: setVisibility, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel() })
  const pageCount = Math.max(1, Math.ceil(total / pageSize))

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === '/' && !(event.target instanceof HTMLInputElement) && !(event.target instanceof HTMLTextAreaElement)) { event.preventDefault(); searchRef.current?.focus() }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  return <div className="grid gap-3">
    <div className="flex flex-wrap items-end justify-between gap-3">
      <div className="flex min-w-0 flex-1 flex-wrap items-end gap-2">{title && <h2 className="mr-2 self-center font-semibold">{title}</h2>}{searchable && <label className="relative block w-full max-w-md"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><Input ref={searchRef} className="pl-9" value={search} onChange={(event) => onSearchChange(event.target.value)} placeholder={searchPlaceholder} aria-label="搜索列表" /></label>}{toolbar}</div>
      <div className="flex items-center gap-2">{headerActions}<Menu.Root><Menu.Trigger render={<Button variant="secondary" />}><Columns3 className="size-4" />列<ChevronDown className="size-3.5" /></Menu.Trigger><Menu.Portal><Menu.Positioner className="z-40" align="end" sideOffset={4}><Menu.Popup className="min-w-40 rounded-md border bg-surface p-1 shadow-lg outline-none">{table.getAllLeafColumns().filter((column) => column.getCanHide()).map((column) => <Menu.CheckboxItem key={column.id} checked={column.getIsVisible()} onCheckedChange={column.toggleVisibility} className="grid cursor-default grid-cols-[1rem_minmax(0,1fr)] items-center gap-2 rounded px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-surface-muted"><span className="grid size-4 place-items-center"><Menu.CheckboxItemIndicator><Check className="size-4 text-primary" /></Menu.CheckboxItemIndicator></span><span>{column.columnDef.header as string}</span></Menu.CheckboxItem>)}</Menu.Popup></Menu.Positioner></Menu.Portal></Menu.Root></div>
    </div>
    {loading ? <LoadingState /> : error ? <ErrorState message={error} onRetry={onRetry} /> : data.length === 0 ? <EmptyState /> : <div className="admin-scroll overflow-x-auto rounded-md border bg-surface"><table className="w-full min-w-max border-collapse text-sm"><thead className="bg-surface-muted text-left text-xs text-muted-foreground">{table.getHeaderGroups().map((group) => <tr key={group.id}>{group.headers.map((header) => <th key={header.id} className="h-10 border-b px-3 font-medium"><button type="button" className="inline-flex items-center gap-1" disabled={!header.column.getCanSort()} onClick={header.column.getToggleSortingHandler()}>{flexRender(header.column.columnDef.header, header.getContext())}{header.column.getCanSort() && <ChevronsUpDown className="size-3" />}</button></th>)}</tr>)}</thead><tbody>{table.getRowModel().rows.map((row) => <tr key={row.id} className="border-b last:border-0 hover:bg-surface-muted/70" onClick={() => onRowClick?.(row.original)}>{row.getVisibleCells().map((cell) => <td key={cell.id} className="h-12 px-3" onClick={cell.column.id === '_actions' ? (event) => event.stopPropagation() : undefined}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>)}</tr>)}</tbody></table></div>}
    <div className="flex items-center justify-between text-xs text-muted-foreground"><span>显示 {total === 0 ? 0 : (page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)}，共 {total} 条</span><div className="flex items-center gap-2"><span>每页</span><Select className="w-20" value={String(pageSize)} onValueChange={(value) => onPageSizeChange(Number(value))} options={[20, 50, 100].map((value) => ({ value: String(value), label: String(value) }))} ariaLabel="每页条数" /><Button variant="secondary" size="icon" disabled={page <= 1} onClick={() => onPageChange(page - 1)}><ChevronLeft className="size-4" /></Button><span>{page} / {pageCount}</span><Button variant="secondary" size="icon" disabled={page >= pageCount} onClick={() => onPageChange(page + 1)}><ChevronRight className="size-4" /></Button></div></div>
  </div>
}
