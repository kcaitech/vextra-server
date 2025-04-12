import { HttpMgr } from './http'
import { UserInfoSchema, BaseResponseSchema, BaseResponse } from './types'
import { z } from 'zod';


// 用户信息模型
const UserProfileSchema = z.object({
    user_id: z.string(),
    nickname: z.string(),
    avatar: z.string()
})

export type UserProfile = z.infer<typeof UserProfileSchema>

// 用户KV存储模型
const UserKVStorageSchema = z.object({
    user_id: z.string(),
    key: z.string(),
    value: z.string()
})

export type UserKVStorage = z.infer<typeof UserKVStorageSchema>

// 用户反馈类型枚举
export enum FeedbackType {
    Report1 = 0, // 举报-欺诈
    Report2 = 1, // 举报-色情低俗
    Report3 = 2, // 举报-不正当言论
    Report4 = 3, // 举报-其他
    Last = 4    // 最后一个
}

// 用户反馈模型
const FeedbackSchema = z.object({
    user_id: z.string(),
    type: z.nativeEnum(FeedbackType),
    content: z.string(),
    image_path_list: z.string(),
    page_url: z.string()
})

export type Feedback = z.infer<typeof FeedbackSchema>

// 用户KV存储响应类型
const UserKVStorageResponseSchema = BaseResponseSchema.extend({
    data: z.record(z.string(), z.string())
})

export type UserKVStorageResponse = z.infer<typeof UserKVStorageResponseSchema>


// 用户信息响应类型
const UserInfoResponseSchema = BaseResponseSchema.extend({
    data: UserInfoSchema
})

export type UserInfoResponse = z.infer<typeof UserInfoResponseSchema>
export type UserInfoData = z.infer<typeof UserInfoResponseSchema.shape.data>



// 回收站列表响应类型
const RecycleListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.number(),
        name: z.string(),
        type: z.string(),
        parent_id: z.string(),
        created_at: z.string(),
        deleted_at: z.string()
    }))
})

export type RecycleListResponse = z.infer<typeof RecycleListResponseSchema>

// 项目列表响应类型
const ProjectListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.number(),
        name: z.string(),
        description: z.string(),
        created_at: z.string()
    }))
})

export type ProjectListResponse = z.infer<typeof ProjectListResponseSchema>

const RefreshTokenSchema = BaseResponseSchema.extend({
    data: z.object({
        token: z.string()
    })
})

export class UsersAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 获取用户信息
    async getInfo(): Promise<UserInfoResponse> {
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
            url: '/documents/feedback',
            method: 'post',
            data: formData,
        })
    }

    async refreshToken() {
        const result = await this.http.request({
            url: 'auth/refresh_token',
            method: 'post',
        })
        try {
            return RefreshTokenSchema.parse(result);
        } catch (error) {
            console.error('RefreshToken数据校验失败:', error);
            throw error;
        }
    }
}

