import axios, { AxiosResponse } from 'axios'
import { HttpCode } from './httpcode'

declare module "axios" {
    interface AxiosResponse<T = any> {
        errorinfo: null
        code: number
        message: string
        error: object
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

type ReqItem = {
    reqid: number,
    args: HttpArgs,
    resolves: ((data: any) => void)[],
    rejects: ((reason?: any) => void)[],
    reqtime: number,
    requesting: boolean,
    sha1?: string,
    data?: any
}

function clone(data: any) {
    return JSON.parse(JSON.stringify(data))
}

export class HttpMgr {
    private _cache = new Map<string, ReqItem>()
    private _cache1 = new Map<number, ReqItem>()
    private reqlist: {
        args: HttpArgs,
        resolve: (data: any) => void,
        reject: (reason?: any) => void,
        cache?: ReqItem
    }[] = []
    private _reqid: number = 0
    private timer: any
    private service: any
    private onUnauthorized: () => void
    private _token?: string

    private get localStorage() {
        if (typeof window !== 'undefined' && window.localStorage) {
            return window.localStorage
        }
        return undefined
    }

    private get token() {
        if (this._token) return this._token
        return this.localStorage?.getItem('token') ?? undefined
    }
    private set token(value: string | undefined) {
        this._token = value
        if (value) {
            this.localStorage?.setItem('token', value)
        } else {
            this.localStorage?.removeItem('token')
        }
    }
    private auth_request(config: any) {
        const token = this.token
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    }

    private handle401() {
        if (this.token) {
            if (this.timer) clearTimeout(this.timer)
            this.timer = setTimeout(() => {
                // ElMessage({ duration: 3000, message: '登录失效，请重新登录', type: 'info' });
                this.timer = undefined;
            }, 500);
        }
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

    constructor(apiUrl: string, onUnauthorized: () => void, token?: string) {
        this._request = this._request.bind(this)
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
        const promise = new Promise<AxiosResponse<any, any>>((resolve, reject) => {
            this.reqlist.push({ args, resolve, reject })
        })
        if (!this.timer) {
            this.timer = setTimeout(this._request, 0)
        }
        return promise
    }

    private _request() {
        this.timer = undefined
        const _caches: ReqItem[] = []
        for (let i = 0; i < this.reqlist.length;) {
            const req = this.reqlist[i]
            const id = JSON.stringify(req.args)
            let cache = this._cache.get(id)
            if (cache && cache.requesting) {
                if (Date.now() - cache.reqtime > REQUEST_TIMEOUT) {
                    cache.rejects.forEach(f => f("Time Out"))
                    cache.rejects.length = 0
                    cache.resolves.length = 0
                } else {
                    cache.resolves.push(req.resolve)
                    cache.rejects.push(req.reject)
                    this.reqlist.splice(i, 1)
                    continue
                }
            }
            cache = {
                reqid: ++this._reqid,
                args: req.args,
                resolves: [req.resolve],
                rejects: [req.reject],
                reqtime: Date.now(),
                requesting: true,
                sha1: cache?.sha1,
                data: cache?.data,
            }
            req.cache = cache
            this._cache.set(id, cache)
            this._cache1.set(cache.reqid, cache)
            _caches.push(cache)
            ++i;
        }
        if (this.reqlist.length > 1) {
            // batch_request
            this.service({
                url: `/batch_request`,
                method: 'post',
                data: this.reqlist.map(v => ({ reqid: v.cache!.reqid, data: v.args, sha1: v.cache!.sha1 }))
            }).then((data: AxiosResponse<{ reqid: number, data?: any, error?: string, sha1?: string }[]>) => {
                for (let i = 0, len = data.data.length; i < len; ++i) {
                    const item = data.data[i]
                    const reqid = item.reqid
                    const sha1 = item.sha1
                    const cache = this._cache1.get(reqid)
                    this._cache1.delete(reqid)
                    if (item.error) console.log(item.error)
                    if (!cache) continue
                    cache.requesting = false
                    if (item.error) {
                        cache.rejects.forEach(f => f(item.error))
                        cache.rejects.length = 0
                        cache.resolves.length = 0
                        continue
                    }
                    if (sha1 && cache.sha1 && sha1 === cache.sha1) {
                        const data = cache.data
                        cache.resolves.forEach(f => f(clone(data)))
                    } else {
                        const data = item.data
                        cache.resolves.forEach(f => f(clone(data)))
                        cache.sha1 = sha1
                        cache.data = item.data
                    }
                    cache.rejects.length = 0
                    cache.resolves.length = 0
                }
            }).catch((err: any) => {
                console.log(err)
                _caches.forEach(c => {
                    c.rejects.forEach(f => f(err))
                    c.rejects.length = 0
                    c.resolves.length = 0
                })
            })
        } else if (this.reqlist.length === 1) {
            // normal_request
            const req = this.reqlist[0]
            this.service(req.args).then((data: AxiosResponse) => {
                const reqid = req.cache!.reqid
                const cache = this._cache1.get(reqid)
                this._cache1.delete(reqid)
                if (cache) {
                    cache.requesting = false
                    cache.data = data
                    cache.sha1 = undefined
                    cache.resolves.forEach(f => f(clone(data)))
                }
            }).catch((err: any) => {
                console.log(err)
                _caches.forEach(c => {
                    c.rejects.forEach(f => f(err))
                    c.rejects.length = 0
                    c.resolves.length = 0
                })
            })
        }

        this.reqlist.length = 0
    }
}