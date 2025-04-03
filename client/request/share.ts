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
  data: z.object({
    total: z.number(),
    items: z.array(ShareSchema),
  }),
});

export type ShareListResponse = z.infer<typeof ShareListResponseSchema>;

// 文档权限类型
export const DocumentPermissionSchema = z.object({
  document_id: z.string(),
  permissions: z.array(z.string()),
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
  id: z.string(),
  name: z.string(),
  type: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
  size: z.number(),
  owner_id: z.string(),
  parent_id: z.string().optional(),
  is_favorite: z.boolean(),
  is_deleted: z.boolean(),
});

export type DocumentInfo = z.infer<typeof DocumentInfoSchema>;

export const DocumentInfoResponseSchema = BaseResponseSchema.extend({
  data: DocumentInfoSchema,
});

export type DocumentInfoResponse = z.infer<typeof DocumentInfoResponseSchema>;

// 分享申请类型
export const ShareApplySchema = z.object({
  document_id: z.string(),
  permissions: z.array(z.string()),
  reason: z.string().optional(),
});

export type ShareApply = z.infer<typeof ShareApplySchema>;

// 分享申请审核类型
export const ShareApplyAuditSchema = z.object({
  apply_id: z.string(),
  status: z.enum(['approved', 'rejected']),
  reason: z.string().optional(),
});

export type ShareApplyAudit = z.infer<typeof ShareApplyAuditSchema>;

export class ShareAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    //获取分享列表
    async getShareListAPI(params: { page?: number; page_size?: number }): Promise<ShareListResponse> {
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
    async getDocumentAuthorityAPI(params: { document_id: string }): Promise<DocumentPermission> {
        const result = await this.http.request({
            url: `/documents/permission`,
            method: 'get',
            params: params,
        });
        try {
            return DocumentPermissionSchema.parse(result.data);
        } catch (error) {
            console.error('文档权限数据校验失败:', error);
            throw error;
        }
    }

    //获取文档密钥
    async getDocumentKeyAPI(params: { document_id: string }): Promise<DocumentKeyResponse> {
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
    async getDocumentInfoAPI(params: { document_id: string }): Promise<DocumentInfoResponse> {
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
    async setShateTypeAPI(params: { document_id: string; share_type: string }): Promise<BaseResponse> {
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
    async putShareAuthorityAPI(params: { document_id: string; permissions: string[] }): Promise<BaseResponse> {
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
    async delShareAuthorityAPI(params: { document_id: string }): Promise<BaseResponse> {
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
    async getApplyListAPI(params: { page?: number; page_size?: number }): Promise<ShareListResponse> {
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
