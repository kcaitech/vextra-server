/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import { HttpArgs, HttpMgr } from './http'
import { BaseResponseSchema, BaseResponse, PermType } from './types';
import { UserInfoSchema, TeamInfoSchema, ProjectInfoSchema, DocumentInfoSchema } from '../common/types';
import { z } from 'zod';

export const ShareListResponseSchema1 = BaseResponseSchema.extend({
    data: z.array(z.object({
        document: DocumentInfoSchema,
        team: TeamInfoSchema.nullable(),
        project: ProjectInfoSchema.nullable(),
        user: UserInfoSchema,
        document_permission: z.object({
            id: z.string(),
            resource_type: z.number(),
            resource_id: z.string(),
            grantee_type: z.number(),
            grantee_id: z.string(),
            created_at: z.string(),
            updated_at: z.string(),
            deleted_at: z.string().nullable(),
            perm_type: z.nativeEnum(PermType),
            perm_source_type: z.number()
        })
    })),
});

export type ShareListResponse1 = z.infer<typeof ShareListResponseSchema1>;

export const ShareApplyListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        document: DocumentInfoSchema,
        team: TeamInfoSchema.nullable(),
        project: ProjectInfoSchema.nullable(),
        apply: z.object({
            id: z.string(),
            user_id: z.string(),
            document_id: z.string(),
            perm_type: z.nativeEnum(PermType),
            status: z.number(),
            created_at: z.string(),
            // updated_at: z.string(),
            deleted_at: z.string().nullable(),
            // first_displayed_at: z.string().nullable(),
            applicant_notes: z.string(),
            processed_at: z.string(),
            processed_by: z.string(),
            processor_notes: z.string(),
        }),
        user: UserInfoSchema.optional(),
        user_team_nickname: z.string().optional()
    }))
});

export type ShareApplyListResponse = z.infer<typeof ShareApplyListResponseSchema>;
export type ShareApplyListItem = z.infer<typeof ShareApplyListResponseSchema.shape.data.element>;


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

const ShareListItemSchema = z.object({
    document: DocumentInfoSchema,
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
    }),
    document_permission: z.object({
        id: z.string(),
        resource_type: z.number(),
        resource_id: z.string(),
        grantee_type: z.number(),
        grantee_id: z.string(),
        perm_type: z.nativeEnum(PermType),
        perm_source_type: z.number()
    })
})

// 共享文件列表响应类型
const ShareListResponseSchema = BaseResponseSchema.extend({
    data: z.array(ShareListItemSchema)
})

export type ShareListResponse = z.infer<typeof ShareListResponseSchema>

export class ShareAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    //查询某个文档对所有用户的分享列表
    async getShareGranteesList(params: { doc_id: string }): Promise<ShareListResponse1> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/share/grantees`,
            method: 'get',
            params: params,
        });
        try {
            return ShareListResponseSchema1.parse(result);
        } catch (error) {
            console.error('分享列表数据校验失败:', error);
            throw error;
        }
    }

    //查询某个文档对所有用户的分享列表
    async getShareReceivesLists(params: {
        cursor?: string;
        limit?: number;
    }): Promise<ShareListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/share/receives',
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

    //设置分享类型
    async setShateType(params: { doc_id: string; doc_type: DocType }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/share/set',
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
    async changeShareAuthority(params: { share_id: string; perm_type: PermType }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/share/',
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
    async delShareAuthority(params: { share_id: string }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/share/perm`,
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
    async applyDocumentAuthority(params: ShareApply): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/share/apply',
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
    async getApplyList(params: { start_time?: number, page?: number; page_size?: number }): Promise<ShareApplyListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/share/apply`,
            method: 'get',
            params: params,
        });
        try {
            return ShareApplyListResponseSchema.parse(result);
        } catch (error) {
            console.error('申请列表数据校验失败:', error);
            throw error;
        }
    }

    watchApplyList(params: () => { start_time?: number, page?: number; page_size?: number }, callback: (data: ShareApplyListResponse) => void, immediate: boolean) {
        const parse = (result: any) => {
            try {
                callback(ShareApplyListResponseSchema.parse(result))
            } catch (error) {
                console.error('申请列表数据校验失败:', error);
                throw error;
            }
        }
        const args = (): HttpArgs => ({
            url: '/share/apply',
            method: 'get',
            params: params(),
        })
        return this.http.watch(args, parse, immediate, true)
    }

    // 权限申请审核
    async permissionApplyAudit(params: ShareApplyAudit): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/share/apply/audit',
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

    //退出共享
    async exitSharing(params: {
        share_id: string;
    }): Promise<BaseResponse> {
        await this.http.refresh_token();
        return this.http.request({
            url: '/share/',
            method: 'delete',
            params: params,
        })
    }
}
