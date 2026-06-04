import { createRequest } from '@/utils/request'
import type {Site, Group, SayingModel} from '@/types/nav'

// 响应拦截器
const navRequest = createRequest(import.meta.env.VITE_NAV_API_BASE_URL)

// 获取站点列表
export function getSites(lang: string): Promise<Site[]> {
    return navRequest.get(`/nav/page/site/list`, { params: { lang } })
}

// 获取分组
export function getGroups(lang: string): Promise<Group[]> {
    return navRequest.get(`/nav/page/group/list`, { params: { lang } })
}

// 获取延迟信息
export function getPing(): Promise<Record<string, string>> {
    return navRequest.get(`/nav/page/ping/list`)
}

// 增加浏览量
export function addCount() {
    navRequest.get(`/nav/stat/add/count`)
}

// 随机金句
export function getSaying(): Promise<SayingModel> {
    return navRequest.get('/nav/page/header/getSaying')
}

export function getImageUrl(type: string): Promise<string> {
    return navRequest.get('/nav/page/header/image/url', { params: { type } })
}

