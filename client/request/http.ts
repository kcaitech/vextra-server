import axios, { AxiosResponse } from 'axios'
import { HttpCode } from './httpcode'
import * as base64 from "js-base64";

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

const REQUEST_TIMEOUT = 10000;

type HttpArgs<T = any> = {
    url: string,
    method: 'post' | 'get' | 'put' | 'delete',
    data?: T,
    params?: any
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
    // private readonly CACHE_EXPIRE_TIME = 30 * 60 * 1000 // 30分钟缓存过期时间

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

        // 检查缓存是否过期
        // if (Date.now() - item.timestamp > this.CACHE_EXPIRE_TIME) {
        //     this._cache.delete(cacheKey)
        //     return null
        // }

        return item
    }

    private setCachedItem(cacheKey: string, data: any, sha1: string) {
        this._cache.set(cacheKey, {
            data,
            sha1,
            timestamp: Date.now()
        })
    }

    private auth_request(config: any) {
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
        } else if (dataAxios.code === HttpCode.StatusBadRequest) {
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
    }) {
        this.onUnauthorized = onUnauthorized
        this.service = axios.create({
            baseURL: apiUrl,
            timeout: REQUEST_TIMEOUT,
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
}