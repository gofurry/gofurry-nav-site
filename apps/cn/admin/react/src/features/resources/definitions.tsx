import { z } from 'zod'
import { StatusBadge } from '../../components/admin/status'
import { formatDate } from '../../lib/utils'
import type { ResourceDefinition, ResourceRecord } from './resource-types'

const optionalText = z.string().max(10_000)
const boolOptions = [{ label: '是', value: '1' }, { label: '否', value: '0' }]

export const resourceDefinitions: ResourceDefinition<ResourceRecord>[] = [
  {
    key: 'sayings', section: 'nav', title: '金句', description: '维护导航站展示的中英文金句。',
    listEndpoint: '/api/v1/nav/sayings', detailEndpoint: '/api/v1/nav/sayings',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'language', label: '语言' }, { key: 'author', label: '作者' }, { key: 'saying', label: '内容' }],
    fields: [{ key: 'language', label: '语言', type: 'select', options: [{ label: '中文', value: 'zh' }, { label: 'English', value: 'en' }] }, { key: 'author', label: '作者', type: 'text' }, { key: 'saying', label: '内容', type: 'textarea' }],
    defaults: { language: 'zh', author: '', saying: '' },
    schema: z.object({ language: z.enum(['zh', 'en']), author: optionalText, saying: z.string().trim().min(1, '请输入金句内容').max(10_000) }),
  },
  {
    key: 'update-notices', section: 'nav', title: '更新公告', description: '发布与维护站点更新公告。',
    listEndpoint: '/api/v1/nav/update-notices', detailEndpoint: '/api/v1/nav/update-notices',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'title', label: '中文标题' }, { key: 'title_en', label: '英文标题' }, { key: 'published_at', label: '发布时间', format: formatDate }],
    fields: [{ key: 'title', label: '中文标题', type: 'text', section: '标题与发布时间' }, { key: 'title_en', label: '英文标题', type: 'text' }, { key: 'published_at', label: '发布时间', type: 'datetime' }, { key: 'body', label: '中文正文', type: 'textarea', section: '公告正文' }, { key: 'body_en', label: '英文正文', type: 'textarea' }],
    defaults: { title: '', title_en: '', published_at: '', body: '', body_en: '' },
    schema: z.object({ title: z.string().trim().min(1, '请输入中文标题'), title_en: optionalText, published_at: z.string().min(1, '请选择发布时间'), body: optionalText, body_en: optionalText }),
  },
  {
    key: 'site-groups', section: 'nav', title: '网站分组', description: '维护用于内容组织的网站业务分组。',
    listEndpoint: '/api/v1/nav/site-groups', detailEndpoint: '/api/v1/nav/site-groups',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'name', label: '名称' }, { key: 'name_en', label: '英文名称' }, { key: 'priority', label: '优先级' }, { key: 'update_time', label: '最近更新', format: formatDate }],
    fields: [{ key: 'name', label: '中文名称', type: 'text', section: '基本内容' }, { key: 'name_en', label: '英文名称', type: 'text' }, { key: 'info', label: '中文简介', type: 'textarea' }, { key: 'info_en', label: '英文简介', type: 'textarea' }, { key: 'priority', label: '优先级', type: 'number', section: '展示顺序', help: '数值越高越靠前。' }],
    defaults: { name: '', name_en: '', info: '', info_en: '', priority: 0 },
    schema: z.object({ name: z.string().trim().min(1, '请输入中文名称'), name_en: optionalText, info: optionalText, info_en: optionalText, priority: z.coerce.number().int() }),
  },
  {
    key: 'tags', section: 'game', title: '标签', description: '维护游戏分类标签；游戏关联在 Game Workspace 中管理。',
    listEndpoint: '/api/v1/game/tags', detailEndpoint: '/api/v1/game/tags',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'name', label: '名称' }, { key: 'name_en', label: '英文名称' }, { key: 'prefix', label: '父标签' }, { key: 'update_time', label: '最近更新', format: formatDate }],
    fields: [{ key: 'id', label: '标签 ID', type: 'number', help: '仅创建时需要；必须与现有标签 ID 不冲突。', section: '标识' }, { key: 'name', label: '中文名称', type: 'text', section: '基本内容' }, { key: 'name_en', label: '英文名称', type: 'text' }, { key: 'info', label: '中文简介', type: 'textarea' }, { key: 'info_en', label: '英文简介', type: 'textarea' }, { key: 'prefix', label: '父标签 ID', type: 'number', help: '-1 表示没有父标签。', section: '分类关系' }],
    defaults: { id: 0, name: '', name_en: '', info: '', info_en: '', prefix: -1 },
    schema: z.object({ id: z.coerce.number().int().nonnegative(), name: z.string().trim().min(1, '请输入中文名称'), name_en: optionalText, info: optionalText, info_en: optionalText, prefix: z.coerce.number().int() }),
  },
  {
    key: 'comments', section: 'game', title: '评论', description: '审核与维护游戏评论内容。',
    listEndpoint: '/api/v1/game/comments', detailEndpoint: '/api/v1/game/comments',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'game_id', label: '游戏' }, { key: 'name', label: '昵称' }, { key: 'score', label: '评分' }, { key: 'content', label: '内容', format: (value: unknown) => <span className="block max-w-[36rem] overflow-hidden text-ellipsis whitespace-nowrap" title={String(value ?? '')}>{String(value ?? '')}</span> }, { key: 'create_time', label: '提交时间', format: formatDate }],
    fields: [{ key: 'game_id', label: '游戏', type: 'remote-select', optionEndpoint: '/api/v1/options/games', section: '评论对象' }, { key: 'region', label: '地区', type: 'text' }, { key: 'name', label: '昵称', type: 'text', section: '评论内容' }, { key: 'ip', label: 'IP', type: 'text' }, { key: 'score', label: '评分', type: 'number' }, { key: 'content', label: '内容', type: 'textarea' }],
    defaults: { game_id: '', region: '', name: '', ip: '', score: 0, content: '' },
    schema: z.object({ game_id: z.coerce.number().int().positive('请选择游戏'), region: optionalText, name: optionalText, ip: optionalText, score: z.coerce.number().min(0), content: z.string().min(1, '请输入评论内容') }),
  },
  {
    key: 'prizes', section: 'game', title: '抽奖', description: '维护抽奖活动、参与口令与奖品 Key。',
    listEndpoint: '/api/v1/game/prizes', detailEndpoint: '/api/v1/game/prizes',
    columns: [{ key: 'id', label: 'ID', hidden: true }, { key: 'title', label: '标题' }, { key: 'status', label: '状态', format: (value: unknown) => <StatusBadge tone={value ? 'success' : 'neutral'}>{value ? '启用' : '停用'}</StatusBadge> }, { key: 'start_time', label: '开始时间', format: formatDate }, { key: 'end_time', label: '结束时间', format: formatDate }],
    fields: [{ key: 'title', label: '标题', type: 'text', section: '活动信息' }, { key: 'desc', label: '描述', type: 'textarea' }, { key: 'key', label: '参与口令', type: 'text' }, { key: 'start_time', label: '开始时间', type: 'datetime', section: '有效期' }, { key: 'end_time', label: '结束时间', type: 'datetime' }, { key: 'status', label: '启用活动', type: 'boolean' }, { key: 'prize.title', label: '奖品标题', type: 'text', section: '奖品' }, { key: 'prize.platform', label: '平台', type: 'text' }, { key: 'prize.keys', label: '奖品 Key', type: 'string-array', help: '每行一个 Key。' }],
    defaults: { title: '', desc: '', key: '', start_time: '', end_time: '', status: true, prize: { title: '', platform: '', keys: [''] } },
    schema: z.object({ title: z.string().trim().min(1, '请输入标题'), desc: optionalText, key: optionalText, start_time: z.string().min(1, '请选择开始时间'), end_time: z.string().min(1, '请选择结束时间'), status: z.boolean(), prize: z.object({ title: z.string().min(1, '请输入奖品标题'), platform: optionalText, keys: z.array(z.string()).min(1) }) }),
  },
]

export function findResource(section?: string, key?: string) {
  return resourceDefinitions.find((definition) => definition.section === section && definition.key === key)
}

export { boolOptions }
