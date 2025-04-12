import { HttpMgr } from "./http"
import { BaseResponse, BaseResponseSchema } from "./types"
import { z } from 'zod';

// 文档列表响应类型
export const DocumentListResponseSchema = BaseResponseSchema.extend({
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

export type DocumentListResponse = z.infer<typeof DocumentListResponseSchema>

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

// 用户文档访问记录列表响应类型
const DocumentAccessRecordListResponseSchema = BaseResponseSchema.extend({
    data: z.array(DocumentAccessRecordSchema)
})

export type DocumentAccessRecordListResponse = z.infer<typeof DocumentAccessRecordListResponseSchema>

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

export class DocumentAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 移除历史记录
    async deleteList(params: {
        access_record_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/access_record',
            method: 'delete',
            params: params,
        })
    }

    // 获取收藏列表
    async getFavoritesList(params: {
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
    async setFavoriteStatus(params: {
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
    async getRecycleList(params: {
        team_id?: string;
        project_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<DocumentListResponse> {
        const result = await this.http.request({
            url: 'documents/recycle_bin',
            method: 'get',
            params: params,
        });
        try {
            return DocumentListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目回收站列表数据校验失败:', error);
            throw error;
        }
    }

    //恢复文件
    async recoverFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/recycle_bin',
            method: 'put',
            data: params,
        })
    }

    //彻底删除文件
    async deleteFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/recycle_bin',
            method: 'delete',
            params: params,
        })
    }

    //文件重命名
    async setFileName(params: {
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
    async copyFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/copy',
            method: 'post',
            data: params,
        })
    }

    //获取文档列表
    async getDocumentList(params: {
        team_id?: string;
        project_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<DocumentListResponse> {
        const result = await this.http.request({
            url: '/documents/',
            method: 'get',
            params: params,
        });
        try {
            return DocumentListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目文件列表数据校验失败:', error);
            throw error;
        }
    }
    // 获取用户文档访问记录列表
    async getUserDocumentAccessRecordsList(): Promise<DocumentAccessRecordListResponse> {
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
    async deleteUserDocumentAccessRecord(params: {
        access_record_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/access_record',
            method: 'delete',
            params: params,
        })
    }


    //收到的共享文件列表
    async shareLists(params: {
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


    //移动文件到回收站
    async moveFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: 'documents/',
            method: 'delete',
            params: params,
        })
    }

}