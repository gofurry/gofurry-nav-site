// 把字符串类型的对象体解析成对象
// 如 '{"ip":"1.1.1.1","country":"US"}' => {"ip":"1.1.1.1","country":"US"}
import axios from "axios";

export function safeJsonParse<T>(data: any): T | null {
    try {
        if (!data) return null
        return typeof data === 'string' ? JSON.parse(data) : data
    } catch (e) {
        console.error('JSON 解析错误:', e)
        return null
    }
}

// 格式化本地时间为 YYYY-MM-DD HH:mm:ss
// 使用本地时间, 不做 UTC 转换
export function formatLocalDateTime(d: Date): string {
    const pad = (n: number) => n.toString().padStart(2, '0')

    return (
        `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
        `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    )
}

// 获取当前本地时间字符串 YYYY-MM-DD HH:mm:ss
export function nowLocalDateTime(): string {
    return formatLocalDateTime(new Date())
}

// 判断一个值是否为有效值 非 null / undefined / 空字符串
// isValidValue('abc') // true
// isValidValue('')    // false
// isValidValue(null)  // false
// isValidValue(0)     // true
export function isValidValue(val: any): boolean {
    return val !== null && val !== undefined && val !== ''
}

// 安全格式化数字
// formatNumber(4.567)        => "4.6"
// formatNumber(4.567, 2)     => "4.57"
// formatNumber(null)         => "0.0"
// formatNumber(NaN, 1, '-')  => "-"
export function formatNumber(
    value: number | null | undefined,
    fixed = 1,
    fallback = '0.0'
): string {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return fallback
    }
    return value.toFixed(fixed)
}

// 防抖函数
export function debounce<T extends (...args: any[]) => void>(
    fn: T,
    delay = 300
): (...args: Parameters<T>) => void {
    let timer: number | undefined

    return (...args: Parameters<T>) => {
        if (timer) window.clearTimeout(timer)
        timer = window.setTimeout(() => {
            fn(...args)
        }, delay)
    }
}

// 节流函数
export function throttle<T extends (...args: any[]) => void>(
    fn: T,
    delay = 300
): (...args: Parameters<T>) => void {
    let last = 0

    return (...args: Parameters<T>) => {
        const now = Date.now()
        if (now - last > delay) {
            last = now
            fn(...args)
        }
    }
}

// 安全截取数组前 N 项
export function take<T>(arr: T[] | null | undefined, count: number): T[] {
    if (!Array.isArray(arr)) return []
    return arr.slice(0, count)
}

// 从 URLSearchParams 中安全获取字符串参数
export function getQueryString(
    params: URLSearchParams,
    key: string,
    defaultValue = ''
): string {
    return params.get(key) ?? defaultValue
}

// 复制文本到剪贴板
export async function copyToClipboard(text: string): Promise<boolean> {
    try {
        await navigator.clipboard.writeText(text)
        return true
    } catch (e) {
        console.error('复制失败:', e)
        return false
    }
}

export function setCookie(name: string, value: string) {
    document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=315360000`
}

export function getCookie(name: string): string | null {
    const match = document.cookie.match(
        new RegExp('(?:^|; )' + name + '=([^;]*)')
    )

    if (!match || !match[1]) return null
    return decodeURIComponent(match[1])
}

export function deleteCookie(name: string) {
    document.cookie = `${name}=; path=/; max-age=0`
}

export function getMarkdown(url: string): Promise<string> {
    return axios.get(url, {
        responseType: 'text',
    }).then(res => res.data)
}