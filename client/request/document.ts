import { HttpMgr } from "./http"
import { BaseResponse, BaseResponseSchema, PermType, ProjectInfoSchema, TeamInfoSchema, UserInfoSchema, DocumentInfoSchema as DocumentSchema } from "./types"
import { z } from 'zod';

// 首先定义 DocumentListItemSchema
export const DocumentListItemSchema = z.object({
    document: DocumentSchema,
    user: UserInfoSchema,
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
    })
});

// 文档列表响应类型
export const DocumentListResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        list: z.array(DocumentListItemSchema),
        has_more: z.boolean(),
        next_cursor: z.string().optional(),
    })
});

export const DocumentRecycleListItemSchema = z.object({
    document: DocumentSchema,
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
    data: z.object({
        list: z.array(DocumentRecycleListItemSchema),
        has_more: z.boolean(),
        next_cursor: z.string().optional(),
    })
})
export type DocumentRecycleListResponse = z.infer<typeof DocumentRecycleListResponseSchema>
export type DocumentRecycleListItem = z.infer<typeof DocumentRecycleListItemSchema>
export type DocumentListResponse = z.infer<typeof DocumentListResponseSchema>
export type DocumentListItem = z.infer<typeof DocumentListItemSchema>

// 用户文档访问记录模型
const DocumentAccessRecordSchema = z.object({
    document: DocumentSchema,
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
    document: DocumentSchema,
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
    document_favorites: z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        is_favorite: z.boolean()
    }),
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

// 文档密钥类型
export const DocumentKeySchema = z.object({
    endpoint: z.string(),
    region: z.string(),
    access_key: z.string(),
    secret_access_key: z.string(),
    session_token: z.string(),
    bucket_name: z.string(),
    provider: z.string(),
});

export type DocumentKey = z.infer<typeof DocumentKeySchema>;

export const DocumentKeyResponseSchema = BaseResponseSchema.extend({
    data: DocumentKeySchema,
});

export type DocumentKeyResponse = z.infer<typeof DocumentKeyResponseSchema>;

export enum LocketType {
    LockedTypeMedia = 0, // 图片审核不通过
    LockedTypeText = 1, // 文本审核不通过
    LockedTypePage = 2, // 页面审核不通过
    LockedTypeComment = 3, // 评论审核不通过
}

// 文档信息类型
export const DocumentInfoSchema1 = z.object({
    document: DocumentSchema,
    team: TeamInfoSchema.nullable(),
    project: ProjectInfoSchema.nullable(),
    document_favorites: z.object({
        id: z.string(),
        is_favorite: z.boolean()
    }),
    document_access_record: z.object({
        id: z.string(),
        last_access_time: z.string()
    }),
    document_permission: z.object({
        id: z.string(),
        resource_type: z.number(),
        resource_id: z.string(),
        grantee_type: z.number(),
        grantee_id: z.string(),
        perm_type: z.number(),
        perm_source_type: z.number()
    }),
    document_permission_requests: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        perm_type: z.number(),
        status: z.number()
    })),
    shares_count: z.number(),
    application_count: z.number(),
    locked_info: z.array(z.object({
        id: z.string(),
        document_id: z.string(),
        created_at: z.string(),
        locked_reason: z.string(),
        locked_words: z.string(),
        locked_type: z.nativeEnum(LocketType),
        lock_target: z.string(),
    })).nullable(),
    user: UserInfoSchema
});

export type DocumentInfo = z.infer<typeof DocumentInfoSchema1>;

export const DocumentInfoResponseSchema = BaseResponseSchema.extend({
    data: DocumentInfoSchema1,
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
        cursor?: string;
        limit?: number;
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
    async getUserDocumentAccessRecordsList(params: {
        cursor?: string;
        limit?: number;
    }): Promise<DocumentListResponse> {
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

    async getResourceDocumentList(params: {
        cursor?: string;
        limit?: number;
    }) : Promise<ResourceDocumentListResponse> {
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

    async createResourceDocument(params: {
        doc_id: string;
        description: string;
    }): Promise<BaseResponse> {
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
        return this.http.request({
            url: 'documents/',
            method: 'delete',
            params: params,
        })
    }

    //获取文档权限
    async getDocumentAuthority(params: { doc_id: string }): Promise<DocumentPermission> {
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

    //获取文档密钥
    async getDocumentAccessKey(params: { doc_id: string }): Promise<DocumentKeyResponse> {
        const result = await this.http.request({
            url: `/documents/access_key`,
            method: 'get',
            params: params,
        });
        try {
            return DocumentKeyResponseSchema.parse(result);
        } catch (error) {
            console.error('文档密钥数据校验失败:', error);
            throw error;
        }
    }

    //获取文档信息
    async getDocumentInfo(params: { doc_id: string }): Promise<DocumentInfoResponse> {
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

}