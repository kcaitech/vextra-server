import { HttpArgs, HttpMgr } from "./http"
import { BaseResponse, BaseResponseSchema, PermType, AccessKeyResponseSchema, AccessKeyResponse, ThumbnailResponseDataSchema, ThumbnailResponse, ThumbnailResponseSchema } from "./types"
import { DocumentInfoSchemaEx, UserInfoSchema, TeamInfoSchema, ProjectInfoSchema, DocumentInfoSchema,  } from "../common/types"
import { z } from 'zod';

// 首先定义 DocumentListItemSchema
export const DocumentListItemSchema = z.object({
    document: DocumentInfoSchema,
    user: UserInfoSchema.optional(),
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
    document_favorites: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        is_favorite: z.boolean(),
        created_at: z.string(),
        updated_at: z.string()
    }),
    document_access_record: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        last_access_time: z.string(),
        created_at: z.string(),
        updated_at: z.string()
    }),
    thumbnail: ThumbnailResponseDataSchema.optional()
});

// 文档列表响应类型
export const DocumentListResponseSchema = BaseResponseSchema.extend({
    data: z.array(DocumentListItemSchema)
});

export const DocumentRecycleListItemSchema = z.object({
    document: DocumentInfoSchema,
    user: UserInfoSchema,
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
    document_favorites: z.object({
        id: z.string(),
        user_id: z.string(),
        is_favorite: z.boolean(),
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
    }).nullable(),
    delete_user: UserInfoSchema.nullable()
})

export const DocumentRecycleListResponseSchema = BaseResponseSchema.extend({
    data: z.array(DocumentRecycleListItemSchema)
})
export type DocumentRecycleListResponse = z.infer<typeof DocumentRecycleListResponseSchema>
export type DocumentRecycleListItem = z.infer<typeof DocumentRecycleListItemSchema>
export type DocumentListResponse = z.infer<typeof DocumentListResponseSchema>
export type DocumentListItem = z.infer<typeof DocumentListItemSchema>

// 用户文档访问记录模型
const DocumentAccessRecordSchema = z.object({
    document: DocumentInfoSchema,
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
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

const ResourceDocumentSchema = z.object({
    document: DocumentInfoSchema,
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
    document_favorites: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        is_favorite: z.boolean()
    }),
    thumbnail: ThumbnailResponseDataSchema.optional()
})

export type ResourceDocument = z.infer<typeof ResourceDocumentSchema>

const ResourceDocumentListResponseSchema = BaseResponseSchema.extend({
    data: z.array(ResourceDocumentSchema)
})

export type ResourceDocumentListResponse = z.infer<typeof ResourceDocumentListResponseSchema>

// 收藏列表响应类型
export type FavoriteListResponse = z.infer<typeof DocumentListResponseSchema>
export type FavoriteListItem = z.infer<typeof DocumentListItemSchema>

// 用户文档访问记录列表响应类型
const DocumentAccessRecordListResponseSchema = BaseResponseSchema.extend({
    data: z.array(DocumentAccessRecordSchema)
})

export type DocumentAccessRecordListResponse = z.infer<typeof DocumentAccessRecordListResponseSchema>

// 文档权限类型
export const DocumentPermissionSchema = BaseResponseSchema.extend({
    data: z.object({
        // document_id: z.string(),
        // permissions: z.array(z.string()),
        perm_type: z.nativeEnum(PermType),
    })
});

export type DocumentPermission = z.infer<typeof DocumentPermissionSchema>;

export const DocumentInfoResponseSchema = BaseResponseSchema.extend({
    data: DocumentInfoSchemaEx,
});

export type DocumentInfoResponse = z.infer<typeof DocumentInfoResponseSchema>;
export type DocumentInfoResponseData = z.infer<typeof DocumentInfoResponseSchema.shape.data>;

export class DocumentAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 移除历史记录
    async deleteAccessRecord(params: {
        access_record_id: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        return this.http.request({
            url: 'documents/access_record',
            method: 'delete',
            params: params,
        })
    }

    // 获取收藏列表
    async getFavoritesList(params: {
        cursor?: string;
        limit?: number;
    }): Promise<FavoriteListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: 'documents/favorites',
            method: 'get',
            params: params,
        })
        try {
            return DocumentListResponseSchema.parse(result)
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
        await this.http.refresh_token();
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
        cursor?: string;
        limit?: number;
    }): Promise<DocumentRecycleListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: 'documents/recycle_bin',
            method: 'get',
            params: params,
        });
        try {
            return DocumentRecycleListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目回收站列表数据校验失败:', error);
            throw error;
        }
    }

    //恢复文件
    async recoverFile(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
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
        await this.http.refresh_token();
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
        await this.http.refresh_token();
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
        await this.http.refresh_token();
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
        cursor?: string;
        limit?: number;
    }): Promise<DocumentListResponse> {
        await this.http.refresh_token();
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
    async getUserDocumentAccessRecordsList(params: {
        cursor?: string;
        limit?: number;
    }): Promise<DocumentListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/documents/access_records',
            method: 'get',
            params: params,
        });
        try {
            return DocumentListResponseSchema.parse(result);
        } catch (error) {
            console.error('用户文档访问记录列表数据校验失败:', error);
            throw error;
        }
    }

    // 获取资源文档列表
    async getResourceDocumentList(params: {
        cursor?: string;
        limit?: number;
    }): Promise<ResourceDocumentListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/documents/resource',
            method: 'get',
            params: params,
        });
        try {
            return ResourceDocumentListResponseSchema.parse(result);
        } catch (error) {
            console.error('资源文档列表数据校验失败:', error);
            throw error;
        }
    }

    // 创建资源文档
    async createResourceDocument(params: {
        doc_id: string;
        description: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        return this.http.request({
            url: '/documents/resource',
            method: 'post',
            data: params,
        })
    }

    //移动文件到回收站
    async moveFileToRecycleBin(params: {
        doc_id: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        return this.http.request({
            url: '/documents/',
            method: 'delete',
            params: params,
        })
    }

    //获取文档权限
    async getDocumentAuthority(params: { doc_id: string }): Promise<DocumentPermission> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/permission`,
            method: 'get',
            params: params,
        });
        try {
            return DocumentPermissionSchema.parse(result);
        } catch (error) {
            console.error('文档权限数据校验失败:', error);
            throw error;
        }
    }

    watchDocumentAuthority(params: () => { doc_id: string }, callback: (data: DocumentPermission) => void, immediate: boolean) {
        const parse = (result: any) => {
            try {
                callback(DocumentPermissionSchema.parse(result))
            } catch (error) {
                console.error('文档权限数据校验失败:', error);
                throw error;
            }
        }
        const args = (): HttpArgs => ({
            url: `/documents/permission`,
            method: 'get',
            params: params(),
        })
        return this.http.watch(args, parse, immediate, true)
    }

    //获取文档密钥
    async getDocumentAccessKey(params: { doc_id: string }): Promise<AccessKeyResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/access_key`,
            method: 'get',
            params: params,
        });
        try {
            return AccessKeyResponseSchema.parse(result);
        } catch (error) {
            console.error('文档密钥数据校验失败:', error);
            throw error;
        }
    }

    async getDocumentThumbnailAccessKey(params: { doc_id: string }): Promise<ThumbnailResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/thumbnail_access_key`,
            method: 'get',
            params: params,
        });
        try {
            return ThumbnailResponseSchema.parse(result);
        } catch (error) {
            console.error('文档缩略图密钥数据校验失败:', error);
            throw error;
        }
    }

    //获取文档信息
    async getDocumentInfo(params: { doc_id: string }): Promise<DocumentInfoResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/info`,
            method: 'get',
            params: params,
        });
        try {
            return DocumentInfoResponseSchema.parse(result);
        } catch (error) {
            console.error('文档信息数据校验失败:', error);
            throw error;
        }
    }

    watchDocumentInfo(params: () => { doc_id: string }, callback: (data: DocumentInfoResponse) => void, immediate: boolean) {
        const parse = (result: any) => {
            try {
                callback(DocumentInfoResponseSchema.parse(result))
            } catch (error) {
                console.error('文档信息数据校验失败:', error);
                throw error;
            }
        }
        const args = (): HttpArgs => ({
            url: `/documents/info`,
            method: 'get',
            params: params(),
        })
        return this.http.watch(args, parse, immediate, true)
    }

    // 重新审核文档
    async reReviewDocument(params: { doc_id: string }): Promise<BaseResponse> {
        return this.http.request({
            url: `/documents/review`,
            method: 'post',
            params: params,
            timeout: 60000,
        })
    }

}