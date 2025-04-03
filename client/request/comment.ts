// 导入axios实例
import { HttpMgr } from './http'
import { BaseResponseSchema, BaseResponse } from './types';
import { z } from 'zod';

// 评论菜单类型
const CommentListMenuSchema = z.object({
    text: z.string(),
    status_p: z.boolean()
})

export type CommentListMenu = z.infer<typeof CommentListMenuSchema>

// 评论状态枚举
export enum CommentStatus {
    Created = 0,
    Resolved = 1
}

// 评论项类型
const CommentItemSchema = z.object({
    id: z.string(),
    parent_id: z.string(),
    root_id: z.string(),
    doc_id: z.string(),
    page_id: z.string(),
    shape_id: z.string(),
    target_shape_id: z.string(),
    shape_frame: z.object({
        x1: z.number(),
        x2: z.number(),
        y1: z.number(),
        y2: z.number()
    }),
    user: z.string(),
    created_at: z.string(),
    record_created_at: z.string(),
    content: z.string(),
    status: z.number(),
    commentMenu: CommentListMenuSchema.array().optional()
})

export type CommentItem = z.infer<typeof CommentItemSchema>

const CommentItemsSchema = z.array(CommentItemSchema)

// 创建评论请求类型
const CreateCommentSchema = z.object({
    id: z.string(),
    parent_id: z.string().optional(),
    root_id: z.string().optional(),
    doc_id: z.string(),
    page_id: z.string(),
    shape_id: z.string(),
    target_shape_id: z.string(),
    shape_frame: z.object({
        x1: z.number(),
        x2: z.number(),
        y1: z.number(),
        y2: z.number()
    }),
    content: z.string()
})

export type CreateComment = z.infer<typeof CreateCommentSchema>

// 更新评论请求类型
const UpdateCommentSchema = z.object({
    id: z.string(),
    parent_id: z.string().optional(),
    root_id: z.string().optional(),
    page_id: z.string().optional(),
    shape_id: z.string().optional(),
    target_shape_id: z.string().optional(),
    shape_frame: z.object({
        x1: z.number(),
        x2: z.number(),
        y1: z.number(),
        y2: z.number()
    }).optional(),
    content: z.string().optional()
})

export type UpdateComment = z.infer<typeof UpdateCommentSchema>

// 设置评论状态请求类型
const SetCommentStatusSchema = z.object({
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
    data: CommentItemSchema
});

export type SingleCommentResponse = z.infer<typeof SingleCommentResponseSchema>;

export class CommentAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    //获取文档评论
    async getDocumentCommentAPI(params: {
        doc_id: string,
        page_id?: string,
        target_shape_id?: string,
        root_id?: string,
        parent_id?: string,
        user_id?: string,
        status?: CommentStatus
    }): Promise<CommentListResponse> {
        const result = await this.http.request({
            url: `/documents/comments`,
            method: 'get',
            params: params,
        })
        try {
            return CommentListResponseSchema.parse(result);
        } catch (error) {
            console.error('获取评论列表数据校验失败:', error);
            throw error
        }
    }

    // 创建评论
    async createCommentAPI(params: CreateComment): Promise<SingleCommentResponse> {
        const validatedParams = CreateCommentSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'post',
            data: validatedParams,
        })
        try {
            return SingleCommentResponseSchema.parse(result);
        } catch (error) {
            console.error('创建评论响应数据校验失败:', error);
            throw error;
        }
    }

    //设置评论状态
    async setCommentStatusAPI(params: SetCommentStatus): Promise<SingleCommentResponse> {
        const validatedParams = SetCommentStatusSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment/status`,
            method: 'put',
            data: validatedParams,
        })
        try {
            return SingleCommentResponseSchema.parse(result);
        } catch (error) {
            console.error('设置评论状态响应数据校验失败:', error);
            throw error;
        }
    }

    //编辑评论
    async editCommentAPI(params: UpdateComment): Promise<SingleCommentResponse> {
        const validatedParams = UpdateCommentSchema.parse(params);
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'put',
            data: validatedParams,
        })
        try {
            return SingleCommentResponseSchema.parse(result);
        } catch (error) {
            console.error('编辑评论响应数据校验失败:', error);
            throw error;
        }
    }

    //删除评论
    async deleteCommentAPI(params: { id: string }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: `/documents/comment`,
            method: 'delete',
            params: params,
        })
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('删除评论响应数据校验失败:', error);
            throw error;
        }
    }
}