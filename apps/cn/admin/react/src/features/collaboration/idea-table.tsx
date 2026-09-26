import { Menu } from '@base-ui/react/menu'
import { DotsThree } from '@phosphor-icons/react'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/auth-context'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Dialog, ConfirmAction } from '../../components/ui/dialog'
import { StatusBadge } from '../../components/admin/status'
import { ApiError, sendJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { base, collaborationError, sourceURL, transition, useRefreshCollaboration } from './api'
import { ideaWorkspacePath } from './idea-context'
import { IdeaEditorDialog, LinkIdeaDialog } from './idea-dialogs'
import { statusLabels, type Idea } from './types'

const menuItem = 'cursor-pointer rounded px-3 py-2 text-sm outline-none data-[highlighted]:bg-surface-muted data-[disabled]:pointer-events-none data-[disabled]:opacity-50'

export function IdeaTableRow({ idea }: { idea: Idea & { idea_duplicate?: boolean } }) {
  const auth = useAuth(), navigate = useNavigate(), refresh = useRefreshCollaboration()
  const [editing, setEditing] = useState(false), [link, setLink] = useState(false)
  const [release, setRelease] = useState(false), [deleting, setDeleting] = useState(false)
  const [researchConflict, setResearchConflict] = useState('')
  const write = auth.can('collaboration.write')
  const active = ['idea', 'researching'].includes(idea.status)
  const canCreate = write && active && idea.kind !== 'other' && auth.can('content.write')
  const hasResource = idea.linked_kind && idea.linked_resource_id && auth.can('content.read')
  const openCreate = () => { if (idea.kind !== 'other') navigate(ideaWorkspacePath(idea.kind, 'new', idea.id)) }
  const mutation = useMutation({ mutationFn: async ({ action, open }: { action: string; open?: boolean }) => {
    if (action === 'delete') return sendJSON(`${base}/ideas/${idea.id}`, 'DELETE', { version: idea.version })
    try { return await transition(idea, action) } catch (error) {
      if (open && error instanceof ApiError && error.status === 409) setResearchConflict(error.message)
      throw error
    }
  }, onSuccess: async (_, variables) => { setRelease(false); setDeleting(false); await refresh(); if (variables.open) openCreate() } })
  const action = (value: string) => mutation.mutate({ action: value })
  const url = sourceURL(idea.source, idea.source_key)
  return <tr className="border-t align-middle">
    <td className="idea-kind"><span className="idea-cell">{idea.kind}</span></td>
    <td className="idea-title"><span className="idea-cell" title={idea.title || idea.source || ''}>{idea.title || idea.source || '—'}</span>{idea.idea_duplicate && <span title="内容池中有相同来源的想法"><StatusBadge tone="info">重复线索</StatusBadge></span>}</td>
    <td className="idea-source">{url ? <a href={url} target="_blank" rel="noreferrer" className="idea-cell text-primary" title={idea.source || ''}>{idea.source}</a> : <span className="idea-cell" title={idea.source || ''}>{idea.source || '—'}</span>}</td>
    <td className="idea-note"><span className="idea-cell" title={idea.note}>{idea.note || '—'}</span></td>
    <td className="idea-priority"><span className="idea-cell">{idea.priority === 'high' ? '优先' : '普通'}</span></td>
    <td className="idea-status"><StatusBadge>{statusLabels[idea.status]}</StatusBadge></td>
    <td className="idea-researcher"><span className="idea-cell" title={idea.researcher_name}>{idea.researcher_name || '—'}</span></td>
    <td className="idea-creator"><span className="idea-cell" title={idea.creator_name}>{idea.creator_name}</span></td>
    <td className="idea-created"><span className="idea-cell" title={formatDate(idea.created_at)}>{formatDate(idea.created_at)}</span></td>
    <td className="idea-actions"><div className="flex items-center justify-end gap-1.5">
      {canCreate && <Button className="idea-primary-action" size="sm" disabled={mutation.isPending} onClick={() => mutation.mutate({ action: 'research', open: true })}>录入{idea.kind === 'game' ? '游戏' : '网站'}</Button>}
      {write && <Button className="idea-edit-action" size="sm" variant="secondary" disabled={mutation.isPending} onClick={() => setEditing(true)}>编辑</Button>}
      <Menu.Root><Menu.Trigger render={<Button size="icon" variant="secondary" disabled={mutation.isPending} aria-label={`想法 ${idea.id} 的更多操作`} />}><DotsThree className="size-5" /></Menu.Trigger>
        <Menu.Portal><Menu.Positioner sideOffset={4} align="end" className="z-40"><Menu.Popup className="min-w-44 rounded-md border bg-surface p-1 shadow-lg outline-none">
          {write && <Menu.Item className={menuItem} onClick={() => setEditing(true)}>编辑想法</Menu.Item>}
          {canCreate && <Menu.Item className={menuItem} onClick={() => mutation.mutate({ action: 'research', open: true })}>录入{idea.kind === 'game' ? '游戏' : '网站'}</Menu.Item>}
          {write && active && <>
            <Menu.Item className={menuItem} onClick={() => action('research')}>开始整理</Menu.Item>
            {auth.can('content.read') && <Menu.Item className={menuItem} onClick={() => setLink(true)}>关联已有内容</Menu.Item>}
            {idea.kind === 'other' && <Menu.Item className={menuItem} onClick={() => action('land')}>标记已落地</Menu.Item>}
            <Menu.Item className={menuItem} onClick={() => action('shelve')}>搁置</Menu.Item>
          </>}
          {write && idea.status === 'researching' && <Menu.Item className={menuItem} onClick={() => idea.researching_by_account_id === auth.state?.identity?.account_id ? action('release') : setRelease(true)}>释放整理</Menu.Item>}
          {write && idea.status === 'shelved' && <Menu.Item className={menuItem} onClick={() => action('restore')}>恢复储备</Menu.Item>}
          {write && idea.status === 'landed' && <Menu.Item className={menuItem} onClick={() => action('reopen')}>重新打开</Menu.Item>}
          {hasResource && <Menu.Item className={menuItem} onClick={() => navigate(ideaWorkspacePath(idea.linked_kind!, idea.linked_resource_id!, idea.id))}>查看正式内容</Menu.Item>}
          {write && <Menu.Item className={`${menuItem} text-danger`} onClick={() => { mutation.reset(); setDeleting(true) }}>删除想法</Menu.Item>}
          {!write && !hasResource && <Menu.Item className={menuItem} disabled>当前为只读模式</Menu.Item>}
        </Menu.Popup></Menu.Positioner></Menu.Portal>
      </Menu.Root>
    </div>
    {mutation.error && !researchConflict && !deleting && <Alert tone="warning">{collaborationError(mutation.error)} <Button variant="secondary" onClick={() => { mutation.reset(); void refresh() }}>重新加载</Button></Alert>}
    {editing && <IdeaEditorDialog idea={idea} close={() => setEditing(false)} />}
    {link && <LinkIdeaDialog idea={idea} close={() => setLink(false)} />}
    <ConfirmAction open={deleting} onOpenChange={(open) => !mutation.isPending && setDeleting(open)} title="删除想法" description="此操作不可撤销，只删除内容想法池中的记录；已关联的正式游戏或网站会保留。" busy={mutation.isPending} onConfirm={() => action('delete')}>
      <p className="mt-3 truncate font-medium" title={idea.title || idea.source || ''}>{idea.title || idea.source}</p>
      {mutation.error && <Alert tone="warning">{collaborationError(mutation.error)}<Button variant="secondary" onClick={() => { setDeleting(false); mutation.reset(); void refresh() }}>重新加载</Button></Alert>}
    </ConfirmAction>
    <ConfirmAction open={release} onOpenChange={setRelease} title="释放其他成员的整理状态" description={`当前由 ${idea.researcher_name} 整理；释放会记录审计。`} confirmLabel="确认释放" variant="primary" busy={mutation.isPending} onConfirm={() => action('release')} />
    <Dialog open={Boolean(researchConflict)} onOpenChange={(open) => !open && setResearchConflict('')} title="协作状态已变化" footer={<><Button variant="secondary" onClick={() => setResearchConflict('')}>取消</Button><Button onClick={openCreate}>仍然打开录入页面</Button></>}><p>{researchConflict}</p><p className="mt-2 text-sm text-muted-foreground">继续打开不会改变当前整理人。</p></Dialog>
    </td>
  </tr>
}
