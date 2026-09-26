import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError, errorMessage, getJSON, sendJSON } from '../../lib/api'
import type { Idea, IdeaInput, PreviewRow } from './types'

export const base = '/api/v1/collaboration'
export const conflictMessage = '此内容刚刚被其他成员修改，请重新加载后重试。'
export function collaborationError(error: unknown) { return error instanceof ApiError && error.status === 409 ? conflictMessage : errorMessage(error) }
export function useIdea(id: string | null, enabled = true) { return useQuery({ queryKey: ['collaboration', 'idea', id], queryFn: () => getJSON<Idea>(`${base}/ideas/${id}`), enabled: Boolean(id) && enabled, retry: false }) }
export function preview(items: IdeaInput[]) { return sendJSON<PreviewRow[]>(`${base}/ideas/batch-preview`, 'POST', { items }) }
export function transition(idea: Idea, action: string, fields: { kind?: string; resource_id?: number } = {}) { return sendJSON<Idea>(`${base}/ideas/${idea.id}/${action}`, 'POST', { version: idea.version, ...fields }) }
export function useRefreshCollaboration() {
  const client = useQueryClient()
  return async () => { await Promise.all([client.invalidateQueries({ queryKey: ['collaboration'] }), client.invalidateQueries({ queryKey: ['workbench'] })]) }
}
export function sourceURL(source: string | null, key?: string | null) {
  if (key?.startsWith('steam:')) return `https://store.steampowered.com/app/${key.slice(6)}/`
  if (!source) return undefined
  try { const url = new URL(source.includes('://') ? source : `https://${source}`); return ['https:', 'http:'].includes(url.protocol) && !url.username && !url.password && url.hostname.includes('.') ? url.href : undefined } catch { return undefined }
}
