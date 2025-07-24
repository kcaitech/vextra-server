import { z } from 'zod';

export enum Provider {
    minio = "minio",
    oss = "oss",
    s3 = "s3",
}

export const AccessKeyInfoSchema = z.object({
    access_key: z.string(),
    secret_access_key: z.string(),
    session_token: z.string(),
    signer_type: z.number(),
    provider: z.nativeEnum(Provider),
    region: z.string(),
    bucket_name: z.string(),
    endpoint: z.string(),
})

export type AccessKeyInfo = z.infer<typeof AccessKeyInfoSchema>;



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
})

export type DocumentInfo = z.infer<typeof DocumentInfoSchema>;


export enum LocketType {
    LockedTypeMedia = 0, // 图片审核不通过
    LockedTypeText = 1, // 文本审核不通过
    LockedTypePage = 2, // 页面审核不通过
    LockedTypeComment = 3, // 评论审核不通过
}

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
// 用户信息类型
export const UserInfoSchema = z.object({
    id: z.string(),
    nickname: z.string(),
    avatar: z.string(),
});

export type UserInfo = z.infer<typeof UserInfoSchema>;

// 文档信息类型
export const DocumentInfoSchemaEx = z.object({
    document: DocumentInfoSchema,
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
    locked_info: z.array(
        LockedInfoSchema
    ).optional(),
    user: UserInfoSchema.nullable(),
});

export type DocumentInfoEx = z.infer<typeof DocumentInfoSchemaEx>;
