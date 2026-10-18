import type { NavUpdateIndexItem } from '../types/nav'
import { updatesMonth } from './updatesDate'

const repositoryCommitURL = 'https://github.com/gofurry/gofurry-nav-site/commit/'

export function updateCommit(sha: string | null) {
  return sha && /^[a-f0-9]{7,64}$/i.test(sha) ? { label: sha.slice(0, 7), href: repositoryCommitURL + sha } : null
}

export function groupUpdatesByMonth(items: NavUpdateIndexItem[]) {
  const groups: { month: string; items: NavUpdateIndexItem[] }[] = []
  for (const item of items) {
    const month = updatesMonth(item.published_at)
    let group = groups.find(value => value.month === month)
    if (!group) { group = { month, items: [] }; groups.push(group) }
    group.items.push(item)
  }
  return groups
}

export const updateCopy = {
  zh: {
    heading: 'GoFurry 更新公告',
    description: '这里记录 GoFurry 持续发布的重要功能、体验改进和维护更新。',
    seoTitle: 'GoFurry 更新公告 - 产品、导航与平台发布记录',
    count: '篇', latestDate: '最近更新', latest: '最新发布', history: '过往发布',
    read: '查看完整更新', loading: '正在读取更新公告。', empty: '暂时还没有更新公告。',
    error: '更新公告暂时不可用。', retry: '重试', all: '全部更新', more: '查看更多更新', moreError: '后续更新暂时不可用，请重试。',
    older: '上一篇', newer: '下一篇', fallback: '阅读 GoFurry 的功能更新、体验改进与维护记录。',
  },
  en: {
    heading: 'GoFurry Updates',
    description: 'New features, experience improvements and maintenance updates from GoFurry.',
    seoTitle: 'GoFurry Updates - Product and platform release notes',
    count: 'releases', latestDate: 'Latest', latest: 'Latest release', history: 'Previous releases',
    read: 'Read full release', loading: 'Loading release notes.', empty: 'No release notes yet.',
    error: 'Release notes are temporarily unavailable.', retry: 'Retry', all: 'All updates', more: 'More updates', moreError: 'More updates are temporarily unavailable. Please retry.',
    older: 'Previous', newer: 'Next', fallback: 'Read GoFurry feature updates, experience improvements and maintenance notes.',
  },
} as const

export function updateDetailSeo(item: NavUpdateIndexItem, lang: 'zh' | 'en') {
  return { title: `${item.version ? item.version + ' — ' : ''}${item.title} | GoFurry`, description: item.summary.trim() || updateCopy[lang].fallback }
}
