// 在 client/request/types.ts 中定义通用类型
import { z } from 'zod';

export enum LocketType {
  LockedTypeMedia = 0, // 图片审核不通过
  LockedTypeText = 1, // 文本审核不通过
  LockedTypePage = 2, // 页面审核不通过
  LockedTypeComment = 3, // 评论审核不通过
}

// 团队权限类型枚举
export enum TeamPermType {
  None = 0,     // 无权限
  ReadOnly = 1,  // 只读
  Commentable = 2,
  Editable = 3,  // 可编辑
  Admin = 4,     // 管理员
  Creator = 5,   // 创建者
  Null = 255, // 无权限
}

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

export const TeamInfoSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  avatar: z.string().optional(),
  invited_perm_type: z.nativeEnum(TeamPermType),
  open_invite: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
  deleted_at: z.string().nullable()
})

export type TeamInfo = z.infer<typeof TeamInfoSchema>;

export const ProjectInfoSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  team_id: z.string(),
  need_approval: z.boolean(),
  perm_type: z.nativeEnum(TeamPermType),
  is_public: z.boolean(),
  open_invite: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
  deleted_at: z.string().nullable()
})

export type ProjectInfo = z.infer<typeof ProjectInfoSchema>;


export const DocumentInfoSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  path: z.string(),
  doc_type: z.number(),
  name: z.string(),
  size: z.number(),
  version_id: z.string(),
  delete_by: z.string(),
  team_id: z.string().nullable(),
  project_id: z.string().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  deleted_at: z.string().nullable(),
  thumbnail: z.string(),
})

export type DocumentInfo = z.infer<typeof DocumentInfoSchema>;

export const LockedInfoSchema = z.object({
  id: z.string(),
  document_id: z.string(),
  created_at: z.string(),
  locked_reason: z.string(),
  locked_words: z.string().optional(),
  lock_type: z.nativeEnum(LocketType),
  lock_target: z.string().optional(),
  deleted_at: z.string().nullable(),
  updated_at: z.string(),
})

export type LockedInfo = z.infer<typeof LockedInfoSchema>;
