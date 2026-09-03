import { useQuery } from '@tanstack/react-query'
import { Browser, GameController, MagnifyingGlass, Stack, Tag } from '@phosphor-icons/react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getJSON } from '../../lib/api'
import type { OptionItem, PageResult } from '../../lib/types'
import { Input } from '../ui/input'
import { Dialog } from '../ui/dialog'
import { EmptyState, LoadingState } from './states'

type SearchGroup = { title: string; icon: typeof MagnifyingGlass; items: Array<OptionItem & { href: string }> }

async function searchEntities(keyword: string): Promise<SearchGroup[]> {
  const params = new URLSearchParams({ page_num: '1', page_size: '8', keyword })
  const [sites, games, tags, groups] = await Promise.all([
    getJSON<PageResult<OptionItem>>(`/api/v1/options/sites?${params}`),
    getJSON<PageResult<OptionItem>>(`/api/v1/options/games?${params}`),
    getJSON<PageResult<OptionItem>>(`/api/v1/options/tags?${params}`),
    getJSON<PageResult<OptionItem>>(`/api/v1/options/site-groups?${params}`),
  ])
  return [
    { title: '网站', icon: Browser, items: sites.list.map((item) => ({ ...item, href: `/nav/sites/${item.id}` })) },
    { title: '游戏', icon: GameController, items: games.list.map((item) => ({ ...item, href: `/game/games/${item.id}` })) },
    { title: '标签', icon: Tag, items: tags.list.map((item) => ({ ...item, href: `/game/tags?search=${encodeURIComponent(item.label)}` })) },
    { title: '网站分组', icon: Stack, items: groups.list.map((item) => ({ ...item, href: `/nav/site-groups?search=${encodeURIComponent(item.label)}` })) },
  ]
}

export function GlobalSearch({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const query = useQuery({ queryKey: ['global-search', keyword], queryFn: () => searchEntities(keyword.trim()), enabled: keyword.trim().length >= 2, staleTime: 30_000 })
  const handleOpenChange = (next: boolean) => { if (!next) setKeyword(''); onOpenChange(next) }
  const resultCount = query.data?.reduce((total, group) => total + group.items.length, 0) ?? 0
  const choose = (href: string) => { onOpenChange(false); navigate(href) }
  return <Dialog open={open} onOpenChange={handleOpenChange} title="全局搜索" description="搜索网站、游戏、标签与网站分组">
    <div className="relative"><MagnifyingGlass className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><Input autoFocus className="h-11 pl-9" value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="输入至少两个字符…" /></div>
    <div className="admin-scroll mt-4 max-h-[55vh] overflow-auto">
      {keyword.trim().length < 2 ? <div className="py-10 text-center text-sm text-muted-foreground">输入名称、域名、Steam AppID 或技术 ID</div> : query.isLoading ? <LoadingState label="正在搜索…" /> : query.error ? <p className="py-8 text-center text-sm text-danger">{query.error.message}</p> : resultCount === 0 ? <EmptyState title="没有匹配结果" /> : <div className="grid gap-4">{query.data?.map((group) => group.items.length > 0 && <section key={group.title}><h3 className="mb-1 flex items-center gap-2 px-2 text-xs font-medium text-muted-foreground"><group.icon className="size-3.5" />{group.title}</h3><div className="grid">{group.items.map((item) => <button key={item.id} type="button" onClick={() => choose(item.href)} className="flex items-center justify-between rounded-md px-2 py-2 text-left hover:bg-surface-muted"><span className="text-sm font-medium">{item.label}</span><span className="font-mono text-xs text-muted-foreground">{item.extra || `#${item.id}`}</span></button>)}</div></section>)}</div>}
    </div>
  </Dialog>
}
