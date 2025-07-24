// 在 client/request/types.ts 中定义通用类型
import { z } from 'zod';
import { AccessKeyInfoSchema } from '../common/types';

// 通用响应类型
export const BaseResponseSchema = z.object({
  code: z.number(),
  message: z.string().optional(),
  data: z.any().optional(),
  sha1: z.string().optional(),
  has_more: z.boolean().optional(),
  next_cursor: z.string().optional(),
});

export const CopyDocumentResponseSchema = z.object({
  copy_id: z.string(),
});

export type CopyDocumentResponse = z.infer<typeof CopyDocumentResponseSchema>;

export type BaseResponse = z.infer<typeof BaseResponseSchema>;


export enum PermType {
  None = 0,
  ReadOnly = 1,
  Commentable = 2,
  Editable = 3,
}


export const AccessKeyResponseSchema = BaseResponseSchema.extend({
    data: AccessKeyInfoSchema,
});

export type AccessKeyResponse = z.infer<typeof AccessKeyResponseSchema>;


export const ThumbnailResponseDataSchema = AccessKeyInfoSchema.extend({
  object_key: z.string(),
});

export type ThumbnailResponseData = z.infer<typeof ThumbnailResponseDataSchema>;

export const ThumbnailResponseSchema = BaseResponseSchema.extend({
  data: ThumbnailResponseDataSchema,
});

export type ThumbnailResponse = z.infer<typeof ThumbnailResponseSchema>;