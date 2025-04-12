// 在 client/request/types.ts 中定义通用类型
import { z } from 'zod';

// 通用响应类型
export const BaseResponseSchema = z.object({
  code: z.number(),
  message: z.string().optional(),
  data: z.any().optional(),
});

export type BaseResponse = z.infer<typeof BaseResponseSchema>;

// 用户信息类型
export const UserInfoSchema = z.object({
  id: z.string(),
  nickname: z.string(),
  avatar: z.string(),
});

export type UserInfo = z.infer<typeof UserInfoSchema>;

// 文档类型
// export const DocumentSchema = z.object({
//   id: z.string(),
//   name: z.string(),
//   type: z.string(),
//   created_at: z.string(),
//   updated_at: z.string(),
//   size: z.number(),
//   owner_id: z.string(),
//   parent_id: z.string().optional(),
//   is_favorite: z.boolean(),
//   is_deleted: z.boolean(),
// });

// export type Document = z.infer<typeof DocumentSchema>;

