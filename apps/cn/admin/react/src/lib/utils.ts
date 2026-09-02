import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(value: unknown) {
  if (!value) return '—'
  const date = new Date(String(value))
  if (Number.isNaN(date.getTime())) return String(value)
  return new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}

export function getAtPath(record: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((value, key) => {
    if (value && typeof value === 'object') return (value as Record<string, unknown>)[key]
    return undefined
  }, record)
}

export function setAtPath(record: Record<string, unknown>, path: string, value: unknown) {
  const keys = path.split('.')
  let cursor = record
  keys.slice(0, -1).forEach((key) => {
    const nested = cursor[key]
    if (!nested || typeof nested !== 'object' || Array.isArray(nested)) cursor[key] = {}
    cursor = cursor[key] as Record<string, unknown>
  })
  cursor[keys.at(-1)!] = value
}
