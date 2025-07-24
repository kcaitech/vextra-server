// 导入axios实例
import { HttpMgr } from './http'
import { BaseResponseSchema, BaseResponse } from './types';
import { UserInfoSchema } from '../common/types';
import { z } from 'zod';

// 评论状态枚举
export enum CommentStatus {
    Created = 0,
    Resolved = 1
}

// 评论项类型
const CommentItemSchema = z.object({
    id: z.string(),
    parent_id: z.string().optional(),
    doc_id: z.string(),
    page_id: z.string(),
    shape_id: z.string().optional(),
    offset_x: z.number(),
    offset_y: z.number(),
    root_x: z.number(),
    root_y: z.number(),
    user: UserInfoSchema,
    created_at: z.string(),
    record_created_at: z.string(),
    content: z.string(),
    status: z.nativeEnum(CommentStatus).optional()
})

export type CommentItem = z.infer<typeof CommentItemSchema>

const CommentItemsSchema = z.array(CommentItemSchema)

// 创建评论请求类型
const CreateCommentSchema = z.object({
    id: z.string(),
    parent_id: z.string().optional(),
    doc_id: z.string(),
    page_id: z.string(),
    shape_id: z.string().optional(),
    offset_x: z.number(),
    offset_y: z.number(),
    root_x: z.number(),
    root_y: z.number(),
    content: z.string(),
})

export type CreateComment = z.infer<typeof CreateCommentSchema>

// 更新评论请求类型
const CommentCommonSchema = z.object({
    doc_id: z.string(),
    id: z.string(),
    parent_id: z.string().optional(),
    page_id: z.string(),
    shape_id: z.string().optional(),
    offset_x: z.number(),
    offset_y: z.number(),
    root_x: z.number(),
    root_y: z.number(),
    content: z.string().optional(),
    status: z.nativeEnum(CommentStatus).optional()
})

export type CommentCommon = z.infer<typeof CommentCommonSchema>

// 设置评论状态参数类型
const SetCommentStatusSchema = z.object({
    doc_id: z.string(),
    id: z.string(),
    status: z.nativeEnum(CommentStatus)
})

export type SetCommentStatus = z.infer<typeof SetCommentStatusSchema>

// 评论列表响应类型
const CommentListResponseSchema = BaseResponseSchema.extend({
    data: CommentItemsSchema
});

export type CommentListResponse = z.infer<typeof CommentListResponseSchema>;

// 单个评论响应类型
const SingleCommentResponseSchema = BaseResponseSchema.extend({
    data: CommentCommonSchema
});

export type SingleCommentResponse = z.infer<typeof SingleCommentResponseSchema>;

export class CommentAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    //获取文档评论
    async list(params: { doc_id: string }): Promise<CommentListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/comments`,
            method: 'get',
            params: params,
        })
        return CommentListResponseSchema.parse(result);
    }

    // 创建评论
    async create(params: CreateComment): Promise<SingleCommentResponse> {
        await this.http.refresh_token();
        const validatedParams = CreateCommentSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'post',
            data: validatedParams,
        })
        return SingleCommentResponseSchema.parse(result);
    }

    // 设置评论状态
    async modifyStatus(params: SetCommentStatus): Promise<SingleCommentResponse> {
        await this.http.refresh_token();
        const validatedParams = SetCommentStatusSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment/status`,
            method: 'put',
            data: validatedParams,
        })
        return SingleCommentResponseSchema.parse(result);
    }

    // 编辑评论
    async modify(params: CommentCommon): Promise<SingleCommentResponse> {
        await this.http.refresh_token();
        const validatedParams = CommentCommonSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'put',
            data: validatedParams,
        })
        return SingleCommentResponseSchema.parse(result);
    }

    // 删除评论
    async remove(params: { comment_id: string, doc_id: string }): Promise<BaseResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'delete',
            params: params,
        })
        return BaseResponseSchema.parse(result);
    }
}