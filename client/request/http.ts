/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import axios, { AxiosResponse } from 'axios'
import { HttpCode } from './httpcode'
import * as base64 from "js-base64";
import { checkRefreshToken } from './refresh_token';

declare module "axios" {
    interface AxiosResponse<T = any> {
        errorinfo: null
        code: number
        message: string
        error: object
        sha1?: string
    }
    export function create(config?: AxiosRequestConfig): AxiosInstance;
}

const REQUEST_TIMEOUT = 10000; // 默认10秒超时

export type HttpArgs<T = any> = {
    url: string,
    method: 'post' | 'get' | 'put' | 'delete',
    data?: T,
    params?: any,
    timeout?: number // 可选的超时时间，单位毫秒
}

type CacheItem = {
    data: any,
    sha1: string,
    timestamp: number
}

function getTokenExp(token: string | undefined): number {
    try {
        if (!token) return 0;
        const res = token.split(".");
        if (res.length !== 3) return 0; // jwt格式不正确
        const r = base64.decode(res[1]);
        const payload = JSON.parse(r);
        return (payload.exp ?? 0) * 1000;
    } catch (e) {
        console.log("parse jwt error", e);
        return 0;
    }
}

export class HttpMgr {
    private service: any
    private onUnauthorized: () => void
    private _token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    }
    private _token_exp_cache: string | undefined = undefined
    private _token_exp: number = 0
    private _cache = new Map<string, CacheItem>()
    private readonly CACHE_EXPIRE_TIME = 30 * 60 * 1000 // 30分钟缓存过期时间

    public get token() {
        return this._token.getToken()
    }

    public set token(value: string | undefined) {
        this._token.setToken(value)
    }

    public get token_remain() {
        const _token = this.token
        if (this._token_exp_cache && this._token_exp_cache === _token) {
            return this._token_exp - Date.now()
        }
        this._token_exp_cache = _token
        this._token_exp = getTokenExp(_token)
        return this._token_exp - Date.now()
    }

    private generateCacheKey(args: HttpArgs): string {
        const { url, method, data, params } = args
        const keyString = JSON.stringify({ url, method, data, params })
        return this.simpleHash(keyString)
    }

    private simpleHash(str: string): string {
        let hash = 0
        if (str.length === 0) return hash.toString()

        for (let i = 0; i < str.length; i++) {
            const char = str.charCodeAt(i)
            hash = ((hash << 5) - hash) + char
            hash = hash & hash // Convert to 32bit integer
        }

        return Math.abs(hash).toString(36)
    }

    private getCachedItem(cacheKey: string): CacheItem | null {
        const item = this._cache.get(cacheKey)
        if (!item) return null
        item.timestamp = Date.now() // 更新缓存时间, 避免被清除
        return item
    }

    private clearExpiredCache() {
        const now = Date.now()
        for (const [key, item] of this._cache) {
            if (now - item.timestamp > this.CACHE_EXPIRE_TIME) {
                this._cache.delete(key)
            }
        }
    }

    private setCachedItem(cacheKey: string, data: any, sha1: string) {
        this._cache.set(cacheKey, {
            data,
            sha1,
            timestamp: Date.now()
        })
    }

    private auth_request(config: any) { // todo 将refresh token 放到这里,但有部分请求不需要refresh token
        const token = this.token
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    }

    private handle401() {
        this.onUnauthorized()
        this.token = undefined
    }

    private handle_response(response: any) {
        const dataAxios = response?.data || {}
        if (dataAxios.code === HttpCode.StatusOK) {
            return Promise.resolve(dataAxios)
        } else {
            return Promise.reject(response)
        }
    }

    private handle_response_err(error: any) {
        const dataAxios = error?.response?.data || {}
        if (error?.status === HttpCode.StatusUnauthorized) {
            this.handle401();
            return Promise.reject(error?.response)
        } else if (dataAxios.code === HttpCode.StatusBadRequest) { // todo 这些错误应该reject的
            return Promise.resolve(dataAxios)
        } else if (dataAxios.code === HttpCode.StatusForbidden) {
            return Promise.resolve(dataAxios)
        } else if (dataAxios.code === HttpCode.StatusInternalServerError) {
            return Promise.resolve(error?.response)
        } else if (dataAxios.code === HttpCode.StatusContentReviewFail) {
            return Promise.resolve(error?.response)
        } else {
            return Promise.reject(error);
        }
    }

    constructor(apiUrl: string, onUnauthorized: () => void, token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    }, timeout: number = REQUEST_TIMEOUT) {
        this.onUnauthorized = onUnauthorized
        this.service = axios.create({
            baseURL: apiUrl,
            timeout: timeout,
        })
        this._token = token

        // 请求拦截器
        this.service.interceptors.request.use(this.auth_request.bind(this), (error: any) => {
            return Promise.reject(error);
        });

        // 响应拦截器
        this.service.interceptors.response.use(
            this.handle_response.bind(this),
            this.handle_response_err.bind(this));

        // 定时清除过期缓存
        setInterval(this.clearExpiredCache.bind(this), this.CACHE_EXPIRE_TIME)
    }

    async request(args: HttpArgs): Promise<AxiosResponse<any, any>> {
        const cacheKey = this.generateCacheKey(args)
        const cachedItem = this.getCachedItem(cacheKey)

        if (cachedItem) { // 有缓存，则添加sha1参数
            if (args.params) {
                args.params.sha1 = cachedItem.sha1
            } else {
                args.params = { sha1: cachedItem.sha1 }
            }
        }

        const response = await this.service(args)

        if (!response.sha1) {
            this._cache.delete(cacheKey)
            return response
        }

        if (cachedItem && response.sha1 === cachedItem.sha1) {
            return {
                ...response,
                data: cachedItem.data
            }
        }
        this.setCachedItem(cacheKey, response.data, response.sha1)
        return response
    }

    // 清除缓存的方法
    clearCache() {
        this._cache.clear()
    }

    // 清除特定请求的缓存
    clearRequestCache(args: HttpArgs) {
        const cacheKey = this.generateCacheKey(args)
        this._cache.delete(cacheKey)
    }

    // watch
    private watch_interval = 1000 * 30 // 30秒
    private watch_interval_id: NodeJS.Timeout | null = null
    private watch_map = new Map<number, { args: () => HttpArgs, callback: ((data: any) => void), refresh_token: boolean }>()
    private watch_id = 0

    watch(args: () => HttpArgs, callback: (data: any) => void, immediate: boolean, refresh_token: boolean) {
        const id = this.watch_id++
        this.watch_map.set(id, { args, callback, refresh_token })
        if (immediate) {
            this.request(args()).then(callback)
        }
        this.start_watch()
        return () => {
            this.unwatch(id)
        }
    }

    private unwatch(id: number) {
        this.watch_map.delete(id)
        if (this.watch_map.size === 0) {
            this.stop_watch()
        }
    }

    private async request_watch() {
        const watch_map = this.watch_map

        const group_watch = new Map<string, { args: HttpArgs, callbacks: ((data: any) => void)[], refresh_token: boolean }>()
        for (const [, item] of watch_map) {
            const args = item.args()
            const cacheKey = this.generateCacheKey(args)
            const group_item = group_watch.get(cacheKey)
            if (group_item) {
                group_item.callbacks.push(item.callback)
            } else {
                group_watch.set(cacheKey, { args, callbacks: [item.callback], refresh_token: item.refresh_token })
            }
        }
        for (const [, item] of group_watch) {
            if (item.refresh_token) {
                await this.refresh_token()
            }
            this.request(item.args).then((data) => {
                for (const callback of item.callbacks) { // 这里需要优化，因为可能存在多个回调
                    try {
                        callback(data)
                    } catch (e) {
                        console.error('watch callback error', e)
                    }
                }
            })
        }
    }

    private start_watch() {
        if (this.watch_interval_id) {
            return
        }
        this.watch_interval_id = setInterval(this.request_watch.bind(this), this.watch_interval)
    }

    private stop_watch() {
        if (this.watch_interval_id) {
            clearInterval(this.watch_interval_id)
            this.watch_interval_id = null
        }
    }

    // refresh token
    async refresh_token() {
        await checkRefreshToken(this)
    }
}