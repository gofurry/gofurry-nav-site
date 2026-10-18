import type { NavUpdateDetail, NavUpdateDetailResponse, NavUpdateIndexItem } from '../../../app/types/nav'
import { runtimeTest, openRuntime, settleRuntime, assertRuntimeSurface, expect } from './insights-runtime'

export const updatesNow = '2026-09-28T08:00:00Z'
export const releaseSHA = '53f429a1234567890abcdef1234567890abcdef1234'
export const indexPath = '/api/v2/nav/updates'
export const articleBody = (en: boolean) => [
  en ? '# A clearer path through GoFurry' : '# 让每一次探索更清晰',
  en ? 'This release brings **site observations** and release notes closer together. Keep `Site` identity separate from its observation targets.' : '本次发布完善了 **站点观测** 与更新记录的阅读体验。站点 `Site` 与观测目标各自承担清晰的职责。',
  en ? '## What changed' : '## 本次更新',
  en ? '- Read protocol evidence in one workspace.\n- Switch languages without losing your place.' : '- 在同一工作区阅读协议观测证据。\n- 切换语言后，继续阅读当前发布记录。',
  en ? '> Observations describe technical evidence, not a safety verdict.' : '> 观测描述技术证据，不代表整体安全或可信度结论。',
  '```text\nGET /api/v2/nav/updates\nGET /api/v2/nav/updates/108?lang=zh\n```',
  en ? 'Read the [project documentation](https://github.com/gofurry/gofurry-nav-site) or return to [all updates](/updates).' : '阅读 [项目文档](https://github.com/gofurry/gofurry-nav-site)，或返回 [全部更新](/updates)。',
  '![GoFurry](/defaultLogo.svg)',
].join('\n\n')

export function releaseItems(locale: 'zh' | 'en'): NavUpdateIndexItem[] {
  const en = locale === 'en'
  const titles = en
    ? ['A new chapter for site discovery', 'Clearer observations, quieter reading', 'Navigation resources and route stability', 'Mobile navigation refinements', 'Year-end data maintenance']
    : ['站点探索体验全面升级', '更清晰的观测，更安静的阅读', '导航资源与路由稳定性调整', '移动端导航细节修复', '年末数据维护完成']
  const dates = ['2026-09-28T07:30:00Z', '2026-09-20T04:15:00Z', '2026-08-30T08:05:00Z', '2026-08-14T06:00:00Z', '2025-12-20T06:00:00Z']
  return titles.map((title, index) => ({ id: 109 - index, title, published_at: dates[index]!,
    version: index < 2 ? `v3.0.0-alpha.${10 - index}` : null, commit_sha: index < 2 ? releaseSHA : null,
    summary: index < 3 ? (en ? 'A focused release for resource discovery, technical observations and a more consistent reading experience.' : '聚焦资源发现、技术观测与阅读体验，让信息更有层次，让每次访问更从容。') : '',
  }))
}

export const test = runtimeTest(
  () => ({ indexState: 'ready' as 'ready' | 'empty' | 'error', detailFailure: 0, extraCount: 0, failedPage: 0 }),
  url => /^\/api\/v2\/nav\/updates(?:\/\d+)?$/.test(url.pathname) && ['zh', 'en'].includes(url.searchParams.get('lang') || ''),
  (url, _media, _body, state) => {
    const locale = url.searchParams.get('lang') === 'en' ? 'en' : 'zh'
    const items = releaseItems(locale)
    if (url.pathname === indexPath) {
      const page = Number(url.searchParams.get('page') || 1), size = Number(url.searchParams.get('page_size') || 100)
      const all = [...items, ...Array.from({ length: state.extraCount }, (_, i) => ({
        id: 104 - i, title: `${locale === 'en' ? 'Archive release' : '历史更新'} ${i + 1}`, summary: '', version: null,
        commit_sha: null, published_at: '2025-11-01T06:00:00Z',
      }))]
      const status = state.failedPage === page ? 'error' : state.indexState
      const visible = status === 'ready' ? all.slice((page - 1) * size, page * size) : []
      return { data: { schema_version: 1, state: status, generated_at: updatesNow, items: visible,
        page, page_size: size, total: status === 'empty' ? 0 : all.length, has_more: status === 'ready' && page * size < all.length } }
    }
    if (state.detailFailure) return { status: state.detailFailure }
    const index = items.findIndex(item => item.id === Number(url.pathname.split('/').at(-1)))
    if (index < 0) return { status: 404 }
    const item: NavUpdateDetail = { ...items[index]!, body: index < 3 ? articleBody(locale === 'en') : (locale === 'en' ? 'A plain text legacy release.' : '这是一条纯文本历史发布记录。') }
    const neighbor = (offset: number) => { const value = items[index + offset]; return value ? { id: value.id, title: value.title, version: value.version, published_at: value.published_at } : null }
    const data: NavUpdateDetailResponse = { schema_version: 1, generated_at: updatesNow, state: 'ready', item, previous: neighbor(1), next: neighbor(-1) }
    return { data }
  },
  { NUXT_PUBLIC_SITE_URL: 'https://go-furry.com', NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com' },
  { fixedTime: updatesNow },
)

export { expect, openRuntime, settleRuntime, assertRuntimeSurface }
