import { HttpMgr } from './http'
import { BaseResponseSchema, BaseResponse, DocumentListResponseSchema, DocumentListResponse } from './types';
import { z } from 'zod';

// 分享相关类型定义
export const ShareSchema = z.object({
    id: z.string(),
    document_id: z.string(),
    share_type: z.string(),
    created_at: z.string(),
    updated_at: z.string(),
    permissions: z.array(z.string()),
});

export type Share = z.infer<typeof ShareSchema>;

export const ShareListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        total: z.number(),
        items: z.array(z.object({
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
                name: z.string()
            }).nullable(),
            project: z.object({
                id: z.string(),
                name: z.string()
            }).nullable(),
            document_permission: z.object({
                id: z.string(),
                resource_type: z.number(),
                resource_id: z.string(),
                grantee_type: z.number(),
                grantee_id: z.string(),
                perm_type: z.number(),
                perm_source_type: z.number()
            })
        }))
    })),
});

export type ShareListResponse = z.infer<typeof ShareListResponseSchema>;


// type PermType uint8

// const (
// 	PermTypeNone        PermType = iota // 无权限
// 	PermTypeReadOnly                    // 只读
// 	PermTypeCommentable                 // 可评论
// 	PermTypeEditable                    // 可编辑
// )

export enum PermType {
    None = 0,
    ReadOnly = 1,
    Commentable = 2,
    Editable = 3,
}

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

// 文档信息类型
export const DocumentInfoSchema = z.object({
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
        name: z.string()
    }).nullable(),
    project: z.object({
        id: z.string(),
        name: z.string()
    }).nullable(),
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
    apply_list: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        document_id: z.string(),
        perm_type: z.number(),
        status: z.number(),
        first_displayed_at: z.string().nullable(),
        processed_at: z.string().nullable(),
        processed_by: z.string().nullable(),
        applicant_notes: z.string().nullable(),
        processor_notes: z.string().nullable()
    })),
    shares_count: z.number(),
    application_count: z.number(),
    locked_info: z.object({
        id: z.string(),
        document_id: z.string(),
        locked_at: z.string(),
        locked_reason: z.string(),
        locked_words: z.string()
    }).nullable()
});

export type DocumentInfo = z.infer<typeof DocumentInfoSchema>;

export const DocumentInfoResponseSchema = BaseResponseSchema.extend({
    data: DocumentInfoSchema,
});

export type DocumentInfoResponse = z.infer<typeof DocumentInfoResponseSchema>;

// 分享申请类型
export const ShareApplySchema = z.object({
    doc_id: z.string(),
    perm_type: z.nativeEnum(PermType),
    applicant_notes: z.string().optional(),
});

export type ShareApply = z.infer<typeof ShareApplySchema>;

// 分享申请审核类型
export const ShareApplyAuditSchema = z.object({
    apply_id: z.string(),
    approval_code: z.number(),
});

export type ShareApplyAudit = z.infer<typeof ShareApplyAuditSchema>;

// DocType 文档类型
export enum DocType {
    Private = 0,// 私有
    Shareable = 1,// 可分享（默认无权限，需申请）
    PublicReadable = 2,// 公共可读
    PublicCommentable = 3,// 公共可评论
    PublicEditable = 4,// 公共可编辑
}

export class ShareAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    //获取分享列表
    async getShareListAPI(params: { doc_id: string }): Promise<ShareListResponse> {
        const result = await this.http.request({
            url: `/documents/shares/all`,
            method: 'get',
            params: params,
        });
        try {
            return ShareListResponseSchema.parse(result);
        } catch (error) {
            console.error('分享列表数据校验失败:', error);
            throw error;
        }
    }

    //获取文档列表
    async getDoucmentListAPI(): Promise<DocumentListResponse> {
        const result = await this.http.request({
            url: '/documents/',
            method: 'get'
        });
        try {
            return DocumentListResponseSchema.parse(result);
        } catch (error) {
            console.error('文档列表数据校验失败:', error);
            throw error;
        }
    }

    //获取文档权限
    async getDocumentAuthorityAPI(params: { doc_id: string }): Promise<DocumentPermission> {
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
    async getDocumentKeyAPI(params: { doc_id: string }): Promise<DocumentKeyResponse> {
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
    async getDocumentInfoAPI(params: { doc_id: string }): Promise<DocumentInfoResponse> {
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

    //设置分享类型
    async setShateTypeAPI(params: { doc_id: string; doc_type: DocType }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/shares/set',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置分享类型响应数据校验失败:', error);
            throw error;
        }
    }

    //修改分享权限
    async putShareAuthorityAPI(params: { share_id: string; perm_type: PermType }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/shares',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('修改分享权限响应数据校验失败:', error);
            throw error;
        }
    }

    //移除分享权限
    async delShareAuthorityAPI(params: { share_id: string }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: `/documents/shares`,
            method: 'delete',
            params: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('移除分享权限响应数据校验失败:', error);
            throw error;
        }
    }

    // 申请文档权限
    async postDocumentAuthorityAPI(params: ShareApply): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/shares/apply',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('申请文档权限响应数据校验失败:', error);
            throw error;
        }
    }

    // 获取申请列表
    async getApplyListAPI(params: { start_time?: number, page?: number; page_size?: number }): Promise<ShareListResponse> {
        const result = await this.http.request({
            url: `/documents/shares/apply`,
            method: 'get',
            params: params,
        });
        try {
            return ShareListResponseSchema.parse(result);
        } catch (error) {
            console.error('申请列表数据校验失败:', error);
            throw error;
        }
    }

    // 权限申请审核
    async promissionApplyAuditAPI(params: ShareApplyAudit): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/shares/apply/audit',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('权限申请审核响应数据校验失败:', error);
            throw error;
        }
    }
}
