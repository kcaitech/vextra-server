import { HttpMgr } from './http'
import { UserInfoSchema, BaseResponseSchema, BaseResponse } from './types'
import { z } from 'zod';

const KVSSchema = z.object({
    key: z.enum(['Preferences', 'FontList']),
    value: z.union([z.string(), z.object({})]).optional()
})
export type KVS = z.infer<typeof KVSSchema>

// // 登录请求
// export function PostLogin(params = {}) {
//     return httpRequest({
//         url: 'auth/login/wx',
//         method: 'post',
//         data: params,
//     })
// }

// // 微信小程序登录请求
// export function PostWxLogin(params = {}) {
//     return httpRequest({
//         url: '/auth/login/wx_mp',
//         method: 'post',
//         data: params,
//     })
// }

//获取小程序码
// export function GetminiProgramCode(params = {}) {
//     return httpRequest({
//         url: '/documents/shares/wx_mp_code',
//         method: 'get',
//         params: params,
//     })
// }

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

// 用户文档访问记录模型
const DocumentAccessRecordSchema = z.object({
    document: z.object({
        id: z.string(),
        user_id: z.string(),
        path: z.string(),
        doc_type: z.number(),
        name: z.string(),
        size: z.number(),
        version_id: z.string(),
        team_id: z.string(),
        project_id: z.string(),
        created_at: z.string(),
        updated_at: z.string(),
        deleted_at: z.string().nullable()
    }),
    team: z.object({
        id: z.string(),
        name: z.string()
    }).nullable(),
    project: z.object({
        id: z.string(),
        name: z.string()
    }).nullable(),
    document_favorites: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        is_favorite: z.boolean()
    }),
    document_access_record: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        last_access_time: z.string()
    })
})

export type DocumentAccessRecord = z.infer<typeof DocumentAccessRecordSchema>

// 用户KV存储响应类型
const UserKVStorageResponseSchema = BaseResponseSchema.extend({
    data: z.record(z.string(), z.string())
})

export type UserKVStorageResponse = z.infer<typeof UserKVStorageResponseSchema>

// 用户文档访问记录列表响应类型
const DocumentAccessRecordListResponseSchema = BaseResponseSchema.extend({
    data: z.array(DocumentAccessRecordSchema)
})

export type DocumentAccessRecordListResponse = z.infer<typeof DocumentAccessRecordListResponseSchema>

// 用户信息响应类型
const UserInfoResponseSchema = BaseResponseSchema.extend({
    data: UserInfoSchema
})

export type UserInfoResponse = z.infer<typeof UserInfoResponseSchema>

// 文档列表响应类型
const UsersDocumentListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        document: z.object({
            id: z.string(),
            user_id: z.string(),
            path: z.string(),
            doc_type: z.number(),
            name: z.string(),
            size: z.number(),
            version_id: z.string(),
            team_id: z.string().nullable(),
            project_id: z.string().nullable(),
            created_at: z.string(),
            updated_at: z.string(),
            deleted_at: z.string().nullable()
        }),
        team: z.object({
            id: z.string(),
            name: z.string(),
            description: z.string().optional(),
            avatar: z.string().optional()
        }).nullable(),
        project: z.object({
            id: z.string(),
            name: z.string(),
            description: z.string().optional()
        }).nullable(),
        document_favorites: z.object({
            id: z.string(),
            user_id: z.string(),
            document_id: z.string(),
            created_at: z.string(),
            updated_at: z.string()
        }).nullable(),
        document_access_record: z.object({
            id: z.string(),
            user_id: z.string(),
            document_id: z.string(),
            last_access_time: z.string(),
            created_at: z.string(),
            updated_at: z.string()
        }).nullable()
    }))
})

export type UsersDocumentListResponse = z.infer<typeof UsersDocumentListResponseSchema>

// 收藏列表响应类型
const FavoriteListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.number(),
        name: z.string(),
        type: z.string(),
        parent_id: z.string(),
        created_at: z.string(),
        updated_at: z.string()
    }))
})

export type FavoriteListResponse = z.infer<typeof FavoriteListResponseSchema>

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

// 共享文件列表响应类型
const ShareListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.number(),
        name: z.string(),
        type: z.string(),
        parent_id: z.string(),
        created_at: z.string(),
        shared_by: z.string()
    }))
})

export type ShareListResponse = z.infer<typeof ShareListResponseSchema>

// 团队列表响应类型
const TeamListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.number(),
        name: z.string(),
        description: z.string(),
        created_at: z.string()
    }))
})

export type TeamListResponse = z.infer<typeof TeamListResponseSchema>

// 团队成员列表响应类型
const TeamMemberListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        user_id: z.string(),
        nickname: z.string(),
        avatar: z.string(),
        perm_type: z.number()
    }))
})

export type TeamMemberListResponse = z.infer<typeof TeamMemberListResponseSchema>

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

export class UsersAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 获取用户信息
    async GetInfo(): Promise<UserInfoResponse> {
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

    //获取历史记录
    async GetDocumentsList(params: {
        page?: number;
        page_size?: number;
        parent_id?: string;
        type?: string;
    }): Promise<UsersDocumentListResponse> {
        const result = await this.http.request({
            url: '/documents/',
            method: 'get',
            params,
        })
        try {
            return UsersDocumentListResponseSchema.parse(result)
        } catch (error) {
            console.error('文档列表数据校验失败:', error)
            throw error
        }
    }

    // 移除历史记录
    async DeleteList(params: {
        access_record_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/access_record',
            method: 'delete',
            params: params,
        })
    }

    // 获取收藏列表
    async GetFavoritesList(params: {
        page?: number;
        page_size?: number;
    }): Promise<FavoriteListResponse> {
        const result = await this.http.request({
            url: 'documents/favorites',
            method: 'get',
            params: params,
        })
        try {
            return FavoriteListResponseSchema.parse(result)
        } catch (error) {
            console.error('收藏列表数据校验失败:', error)
            throw error
        }
    }

    //设置收藏列表
    async SetFavoriteStatus(params: {
        doc_id: string;
        status: boolean;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/favorites',
            method: 'put',
            data: params,
        })
    }

    //获取回收站列表
    async GetRecycleList(params: {
        page?: number;
        page_size?: number;
    }): Promise<RecycleListResponse> {
        const result = await this.http.request({
            url: 'documents/recycle_bin',
            method: 'get',
            params: params,
        })
        try {
            return RecycleListResponseSchema.parse(result)
        } catch (error) {
            console.error('回收站列表数据校验失败:', error)
            throw error
        }
    }

    //恢复文件
    async RecoverFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/recycle_bin',
            method: 'put',
            data: params,
        })
    }

    //彻底删除文件
    async DeleteFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/recycle_bin',
            method: 'delete',
            params: params,
        })
    }

    //退出共享
    async ExitSharing(params: {
        share_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/share',
            method: 'delete',
            params: params,
        })
    }

    //移动文件到回收站
    async MoveFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/',
            method: 'delete',
            params: params,
        })
    }

    //收到的共享文件列表
    async ShareLists(params: {
        page?: number;
        page_size?: number;
    }): Promise<ShareListResponse> {
        const result = await this.http.request({
            url: 'documents/shares',
            method: 'get',
            params: params,
        })
        try {
            return ShareListResponseSchema.parse(result)
        } catch (error) {
            console.error('共享文件列表数据校验失败:', error)
            throw error
        }
    }

    //设置用户头像
    async SetAvatar(file: File): Promise<BaseResponse> {
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
    async SetNickname(params: {
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

    //文件重命名
    async SetFileName(params: {
        doc_id: string;
        name: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/name',
            method: 'put',
            data: params,
        })
    }

    //复制文档
    async CopyFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/copy',
            method: 'post',
            data: params,
        })
    }

    //创建团队
    async CreateTeam(params: {
        name: string;
        description: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team',
            method: 'post',
            data: params,
        })
    }

    //获取团队列表
    async GetteamList(params: {
        page?: number;
        page_size?: number;
    }): Promise<TeamListResponse> {
        const result = await this.http.request({
            url: '/documents/team/list',
            method: 'get',
            params: params,
        })
        try {
            return TeamListResponseSchema.parse(result)
        } catch (error) {
            console.error('团队列表数据校验失败:', error)
            throw error
        }
    }

    //获取团队成员
    async GetteamMember(params: {
        team_id: string;
        page?: number;
        page_size?: number;
    }): Promise<TeamMemberListResponse> {
        const result = await this.http.request({
            url: '/documents/team/member/list',
            method: 'get',
            params: params,
        })
        try {
            return TeamMemberListResponseSchema.parse(result)
        } catch (error) {
            console.error('团队成员列表数据校验失败:', error)
            throw error
        }
    }

    //设置团队信息
    // async Setteaminfo(params: {
    //     team_id: string;
    //     name: string;
    //     description: string;
    // }): Promise<BaseResponse> {
    //     return this.http.request({
    //         url: '/documents/team/info',
    //         method: 'put',
    //         data: params,
    //     })
    // }


    //获取项目列表
    // async GetProjectLists(params: {
    //     page?: number;
    //     page_size?: number;
    // }): Promise<ProjectListResponse> {
    //     const result = await this.http.request({
    //         url: '/documents/project/list',
    //         method: 'get',
    //         params: params,
    //     })
    //     try {
    //         return ProjectListResponseSchema.parse(result)
    //     } catch (error) {
    //         console.error('项目列表数据校验失败:', error)
    //         throw error
    //     }
    // }

    //创建项目
    // async CreateProject(params: {
    //     name: string;
    //     description: string;
    // }): Promise<BaseResponse> {
    //     return this.http.request({
    //         url: '/documents/project',
    //         method: 'post',
    //         data: params,
    //     })
    // }

    //获取文档列表
    async getDoucmentListAPI(params: {
        page?: number;
        page_size?: number;
        parent_id?: string;
        type?: string;
    }): Promise<UsersDocumentListResponse> {
        const result = await this.http.request({
            url: '/documents/list',
            method: 'get',
            params: params,
        })
        try {
            return UsersDocumentListResponseSchema.parse(result)
        } catch (error) {
            console.error('文档列表数据校验失败:', error)
            throw error
        }
    }

    //获取KV存储
    async GetKVStorage(params: {
        key: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/users/kv_storage',
            method: 'get',
            params: params,
        })
    }

    //设置KV存储
    async SetKVStorage(params: {
        key: string;
        value: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/users/kv_storage',
            method: 'put',
            data: params,
        })
    }

    // 获取用户KV存储
    async getUserKVStorageAPI(params: {
        key: string;
    }): Promise<UserKVStorageResponse> {
        const result = await this.http.request({
            url: '/users/user_kv_storage',
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
    async setUserKVStorageAPI(params: {
        key: string;
        value: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/users/user_kv_storage',
            method: 'post',
            data: params,
        })
    }

    // 获取用户文档访问记录列表
    async getUserDocumentAccessRecordsListAPI(): Promise<DocumentAccessRecordListResponse> {
        const result = await this.http.request({
            url: '/documents/access_records',
            method: 'get',
        });
        try {
            return DocumentAccessRecordListResponseSchema.parse(result);
        } catch (error) {
            console.error('用户文档访问记录列表数据校验失败:', error);
            throw error;
        }
    }

    // 删除用户文档访问记录
    async deleteUserDocumentAccessRecordAPI(params: {
        access_record_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/access_record',
            method: 'delete',
            params: params,
        })
    }

    // 提交用户反馈
    async submitFeedbackAPI(params: {
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
}

