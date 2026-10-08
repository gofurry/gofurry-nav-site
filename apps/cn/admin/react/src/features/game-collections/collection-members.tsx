import { useMemo, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { RemoteSelect } from '../../components/admin/operations'
import { FormField } from '../../components/admin/page'
import { StatusBadge } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { useCompositionSafeSearch } from '../../hooks/use-composition-safe-search'
import { errorMessage, getJSON } from '../../lib/api'
import type { GameWorkspace } from '../../lib/types'
import { changeMemberOverride, compositionChanged, previewComposition } from './composition-draft'
import type { CollectionComposition, CollectionMember, CompositionDraft, EffectiveCollectionMember } from './types'

const sourceLabels = { automatic: '自动', manual: '人工', both: '自动 + 人工' }

export function CollectionCompositionEditor({ base, value, onChange, readonly, busy, onLoadingChange }: {
  base: CollectionComposition; value: CompositionDraft; onChange: (value: CompositionDraft) => void
  readonly: boolean; busy: boolean; onLoadingChange: (loading: boolean) => void
}) {
  const [error, setError] = useState('')
  const members = useMemo(() => previewComposition(base, value), [base, value])
  const dirty = compositionChanged(value, base)
  const override = (member: CollectionMember, action: 'pin' | 'unpin' | 'exclude' | 'restore') => onChange(changeMemberOverride(value, member, action))
  return <div className="grid min-w-0 gap-4">
    <div className="grid min-w-0 gap-2">
      <h3 className="text-sm font-medium">自动收录标签</h3>
      <p className="text-xs text-muted-foreground">匹配任意已绑定标签的游戏自动纳入（OR）；失效规则保留，恢复后重新生效。</p>
      {value.rule_tags.length === 0 ? <p className="text-sm text-muted-foreground">未绑定标签，仅使用人工固定成员。</p> : <ul aria-label="自动收录标签" className="flex flex-wrap gap-2">{value.rule_tags.map(tag => <li key={tag.tag_id} className="flex min-w-0 max-w-full flex-wrap items-center gap-x-2 gap-y-1 rounded-md border px-2 py-1 text-sm">
        <span className="min-w-0 break-words">{tag.name || tag.name_en}</span>
        {!tag.active && <StatusBadge tone="warning">失效</StatusBadge>}
        {!readonly && <Button className="text-foreground" variant="ghost" size="sm" disabled={busy} aria-label={`解绑 ${tag.name || tag.name_en}`} onClick={() => onChange({ ...value, rule_tags: value.rule_tags.filter(item => item.tag_id !== tag.tag_id) })}>解绑</Button>}
      </li>)}</ul>}
      {!readonly && <div className="max-w-md"><FormField label="添加自动收录标签"><RemoteSelect endpoint="/api/v1/options/tags" value={null} placeholder="搜索并绑定标签…" pageSize={20} debounceMs={300} disabled={busy} excludeIDs={value.rule_tags.map(tag => String(tag.tag_id))} onChange={option => {
        if (!option || value.rule_tags.some(tag => String(tag.tag_id) === option.id)) return
        // The existing options API lists only active Tags in active Categories.
        onChange({ ...value, rule_tags: [...value.rule_tags, { tag_id: Number(option.id), code: '', name: option.label, name_en: '', active: true }] })
      }} /></FormField></div>}
    </div>
    <div className="grid gap-2">
      <dl aria-label="已保存收录统计" className="flex flex-wrap gap-x-5 gap-y-2 text-sm">
        {([['自动匹配', base.counts.auto_matched], ['人工固定', base.counts.manual_pinned], ['已排除', base.counts.excluded], ['有效收录', base.counts.effective], ['SFW 可见', base.counts.sfw_visible]] as const).map(([label, count]) => <div key={label} className="flex gap-2"><dt className="text-muted-foreground">{label}</dt><dd className="font-medium tabular-nums">{count}</dd></div>)}
      </dl>
      {dirty && <p role="status" className="text-xs text-muted-foreground">收录设置尚未保存。上方数量为已保存结果；下方预览本地固定与排除，标签匹配和取消排除后的有效成员将在保存后确认。</p>}
      {(base.counts.auto_matched > 100 || members.length > 100) && <p className="text-xs text-muted-foreground">当前收录较多，列表分页展示；保存始终保留完整配置，不会截断成员。</p>}
    </div>
    <p className="text-sm text-muted-foreground">游戏顺序由首次可用时间与发行事实自动计算，当前成员列表顺序不会影响公开时间线。</p>
    {!readonly && <div className="max-w-md"><FormField label="人工固定游戏"><RemoteSelect endpoint="/api/v1/options/games" value={null} placeholder="搜索并固定游戏…" pageSize={20} debounceMs={300} disabled={busy} excludeIDs={value.manual_members.map(member => String(member.game_id))} onChange={async option => {
      if (!option || value.manual_members.some(member => String(member.game_id) === option.id)) return
      // Reuse known metadata; remote games and remote tags have independent owners.
      const known = [...base.effective_members, ...value.excluded_members].find(member => String(member.game_id) === option.id)
      if (known) { override(known, 'pin'); return }
      onLoadingChange(true); setError('')
      try {
        const { game, tags } = await getJSON<GameWorkspace>(`/api/v1/game/games/${option.id}/workspace`)
        override({ game_id: Number(game.id), name: game.name, name_en: game.name_en, appid: game.appid, adult: tags.some(tag => tag.code === 'adult') }, 'pin')
      } catch (err) { setError(errorMessage(err)) }
      finally { onLoadingChange(false) }
    }} /></FormField></div>}
    {error && <Alert tone="danger">{error}</Alert>}
    <MemberList members={members} excluded={false} renderActions={member => {
      const source = (member as EffectiveCollectionMember).source
      return <>
        <StatusBadge tone={source === 'automatic' ? 'info' : 'neutral'}>{sourceLabels[source]}</StatusBadge>
        {!readonly && <div className="flex flex-wrap gap-1">
          {source === 'automatic' ? <Button className="text-foreground" variant="ghost" disabled={busy} aria-label={`固定保留 ${member.name}`} onClick={() => override(member, 'pin')}>固定保留</Button> : <Button className="text-foreground" variant="ghost" disabled={busy} aria-label={`移除固定 ${member.name}`} onClick={() => override(member, 'unpin')}>移除固定</Button>}
          <Button className="text-foreground" variant="ghost" disabled={busy} aria-label={`排除 ${member.name}`} onClick={() => override(member, 'exclude')}>排除</Button>
        </div>}
      </>
    }} />
    <div className="grid gap-3 border-t pt-4">
      <h3 className="text-sm font-medium">人工排除（{value.excluded_members.length}）</h3>
      {value.excluded_members.length === 0 ? <p className="text-sm text-muted-foreground">没有人工排除的游戏。</p> : <MemberList members={value.excluded_members} excluded renderActions={member => !readonly && <div className="flex flex-wrap gap-1">
        <Button className="text-foreground" variant="ghost" disabled={busy} aria-label={`取消排除 ${member.name}`} onClick={() => override(member, 'restore')}>取消排除</Button>
        <Button className="text-foreground" variant="ghost" disabled={busy} aria-label={`固定保留 ${member.name}`} onClick={() => override(member, 'pin')}>固定保留</Button>
      </div>} />}
    </div>
  </div>
}

function MemberList({ members, excluded, renderActions }: { members: CollectionMember[]; excluded: boolean; renderActions: (member: CollectionMember) => ReactNode }) {
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const search = useCompositionSafeSearch({ value: keyword, onCommit: value => { setKeyword(value); setPage(1) } })
  const filtered = useMemo(() => {
    const query = keyword.trim().toLowerCase()
    return members.filter(member => [member.name, member.name_en, String(member.appid)].some(value => value.toLowerCase().includes(query)))
  }, [members, keyword])
  const pageSize = 20
  const pageCount = Math.max(1, Math.ceil(filtered.length / pageSize))
  const currentPage = Math.min(page, pageCount)
  const paged = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  if (page > pageCount) setPage(pageCount)
  const name = excluded ? '已排除游戏' : '已收录游戏'
  const pagination = excluded ? '排除成员' : '成员'
  return <div className="grid min-w-0 gap-3">
    <Input aria-label={`搜索${name}`} placeholder={`搜索${name}…`} {...search.inputProps} />
    {members.length === 0 ? <p className="text-sm text-muted-foreground">尚未收录游戏</p> : filtered.length === 0 ? <p className="text-sm text-muted-foreground">未找到匹配的{name}</p> : <ul aria-label={name} className="divide-y">{paged.map(member => <li key={member.game_id} className="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-2 py-3">
      <Link className="min-w-0 flex-1 basis-32 break-words text-primary" to={`/game/games/${member.game_id}`}>{member.name || member.name_en}</Link>
      {member.adult && <StatusBadge>Adult</StatusBadge>}
      {renderActions(member)}
    </li>)}</ul>}
    <div className="flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground" aria-label={`${pagination}分页`}>
      <span>显示 {filtered.length === 0 ? 0 : (currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filtered.length)}，共 {filtered.length} 个</span>
      <div className="flex shrink-0 items-center gap-2">
        <Button className="text-foreground" variant="secondary" size="icon" aria-label={`${pagination}上一页`} disabled={currentPage <= 1} onClick={() => setPage(currentPage - 1)}>‹</Button>
        <span>{currentPage} / {pageCount}</span>
        <Button className="text-foreground" variant="secondary" size="icon" aria-label={`${pagination}下一页`} disabled={currentPage >= pageCount} onClick={() => setPage(currentPage + 1)}>›</Button>
      </div>
    </div>
  </div>
}
