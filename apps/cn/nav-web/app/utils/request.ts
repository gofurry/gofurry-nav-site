import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

/*
* @Desc: Axios封装
* @Author: 福狼
* @Version: V1.0.1
* */

/*
* response.data {
*   code: 1,    code: 0,
*   data: {
*       ...
*   }
* }
* */

export function createRequest(baseURL?: string): AxiosInstance {
    const service: AxiosInstance = axios.create({
        baseURL,
        timeout: 10000,
        // withCredentials: true,
    })

    // 请求拦截器
    service.interceptors.request.use(
        (config: InternalAxiosRequestConfig) => {
            // const token = localStorage.getItem('token')
            // if (token) config.headers = { ...config.headers, Authorization: `Bearer ${token}` }
            return config
        },
        (error) => Promise.reject(error)
    )

    // 响应拦截器
    service.interceptors.response.use(
        (response) => {
            if (response.data.code !== 1) {
                console.error('接口返回错误:', response.data)
                return [] // 置空避免污染前端入参
            }
            return response.data.data
        },
        (error) => {
            console.error('网络错误:', error)
            return Promise.reject(error)
        }
    )

    return service
}
