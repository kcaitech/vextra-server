import { HttpMgr } from './http'
import { UserInfoSchema, BaseResponseSchema, BaseResponse } from './types'
import { z } from 'zod';

// 用户反馈类型枚举
export enum FeedbackType {
    Report1 = 0, // 举报-欺诈
    Report2 = 1, // 举报-色情低俗
    Report3 = 2, // 举报-不正当言论
    Report4 = 3, // 举报-其他
    Last = 4    // 最后一个
}

// 用户KV存储响应类型
const UserKVStorageResponseSchema = BaseResponseSchema.extend({
    data: z.record(z.string(), z.string())
})

export type UserKVStorageResponse = z.infer<typeof UserKVStorageResponseSchema>


// 用户信息响应类型
const UserInfoResponseSchema = BaseResponseSchema.extend({
    data: UserInfoSchema
})

const UserInfoWithTokenResponseSchema = UserInfoSchema.extend({
    token: z.string(),
})

export type UserInfoResponse = z.infer<typeof UserInfoResponseSchema>
export type UserInfoWithTokenResponse = z.infer<typeof UserInfoWithTokenResponseSchema>
export type UserInfoData = z.infer<typeof UserInfoResponseSchema.shape.data>


export class UsersAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 获取用户信息
    async getInfo(): Promise<UserInfoResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/users/info',
            method: 'get',
        })
        try {
            return UserInfoResponseSchema.parse(result)
        } catch (error) {
            console.error('用户信息数据校验失败:', error)
            throw error
        }
    }

    //设置用户头像
    async setAvatar(file: File): Promise<BaseResponse> {
        await this.http.refresh_token();
        const formData = new FormData()
        formData.append('file', file)

        const result = await this.http.request({
            url: '/users/info/avatar',
            method: 'put',
            data: formData,
        })
        try {
            return BaseResponseSchema.parse(result)
        } catch (error) {
            console.error('设置头像响应数据校验失败:', error)
            throw error
        }
    }

    //设置用户昵称
    async setNickname(params: {
        nickname: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/users/info/nickname',
            method: 'put',
            data: params,
        })
        try {
            return BaseResponseSchema.parse(result)
        } catch (error) {
            console.error('设置昵称响应数据校验失败:', error)
            throw error
        }
    }

    // 获取用户KV存储
    async getKVStorage(params: {
        key: string;
    }): Promise<UserKVStorageResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/users/kv_storage',
            method: 'get',
            params: params,
        });
        try {
            return UserKVStorageResponseSchema.parse(result);
        } catch (error) {
            console.error('用户KV存储数据校验失败:', error);
            throw error;
        }
    }

    // 设置用户KV存储
    async setKVStorage(params: {
        key: string;
        value: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        return this.http.request({
            url: '/users/kv_storage',
            method: 'post',
            data: params,
        })
    }

    // 提交用户反馈
    async submitFeedback(params: {
        type: FeedbackType;
        content: string;
        page_url?: string;
        files?: ArrayBuffer[];
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const formData = new FormData();
        formData.append('type', params.type.toString());
        formData.append('content', params.content);
        if (params.page_url) {
            formData.append('page_url', params.page_url);
        }
        else {
            formData.append('page_url', '');
        }
        if (params.files) {
            for (let i = 0; i < params.files.length; i++) {
                formData.append('files', new Blob([params.files[i]]));
            }
        }
        return this.http.request({
            url: '/feedback',
            method: 'post',
            data: formData,
        })
    }

    // async refreshToken() {
    //     return refreshToken(this.http);
    // }

    async getLoginUrl(): Promise<{ url: string, client_id?: string }> {
        const result = await this.http.request({
            url: 'auth/login_url',
            method: 'get',
        })
        // 需要处理一下url，如果url中包含~，则需要替换为当前域名
        let url = result.data.url;
        if (url.includes('~')) {
            const host = typeof window !== 'undefined' ? window.location.host : 'localhost';
            url = url.replace('~', host);
        }
        if (!url.startsWith('http://') && !url.startsWith('https://')) {
            const protocol = typeof window !== 'undefined' ? window.location.protocol : 'http:';
            url = protocol + '//' + url;
        }
        return { url: url, client_id: result.data.client_id };
    }

    async loginCallback(code: string): Promise<UserInfoWithTokenResponse> {
        const result = await this.http.request({
            url: 'auth/login/callback',
            method: 'get',
            params: {
                code: code,
            },
        })
        try {
            return UserInfoWithTokenResponseSchema.parse(result.data)
        } catch (error) {
            console.error('用户信息数据校验失败:', error)
            throw error
        }
    }

    async logout(): Promise<BaseResponse> {
        return this.http.request({
            url: 'auth/logout',
            method: 'get',
        })
    }

    // 微信小程序登录
    async wechatMiniProgramLogin(code: string): Promise<UserInfoWithTokenResponse> {
        const result = await this.http.request({
            url: 'auth/login/mini_program',
            method: 'get',
            params: {
                code: code,
            },
        })
        try {
            return UserInfoWithTokenResponseSchema.parse(result.data)
        } catch (error) {
            console.error('微信小程序登录响应数据校验失败:', error)
            throw error
        }
    }
}

