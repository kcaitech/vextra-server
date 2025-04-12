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

export enum PermType {
  None = 0,
  ReadOnly = 1,
  Commentable = 2,
  Editable = 3,
}
