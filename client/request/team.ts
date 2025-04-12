import { HttpMgr } from './http'
import { BaseResponseSchema, BaseResponse, UserInfoSchema } from './types'
import { z } from 'zod'

// 团队权限类型枚举
export enum TeamPermType {
    None = 0,     // 无权限
    ReadOnly = 1,  // 只读
    Commentable = 2,
    Editable = 3,  // 可编辑
    Admin = 4,     // 管理员
    Creator = 5,   // 创建者
}

// 团队模型
const TeamSchema = z.object({
    team: z.object({
        id: z.string(),
        name: z.string(),
        description: z.string().optional(),
        avatar: z.string().optional(),
        invited_perm_type: z.nativeEnum(TeamPermType),
        open_invite: z.boolean(),
        created_at: z.string(),
        updated_at: z.string()
    }),
    creator: UserInfoSchema,
    self_perm_type: z.nativeEnum(TeamPermType)
})

export type Team = z.infer<typeof TeamSchema>

// 团队成员模型
const TeamMemberSchema = z.object({
    id: z.string(),
    team_id: z.string(),
    user_id: z.string(),
    perm_type: z.nativeEnum(TeamPermType),
    nickname: z.string(),
    created_at: z.string(),
    updated_at: z.string()
})

export type TeamMember = z.infer<typeof TeamMemberSchema>

// 团队成员带用户信息模型
const TeamMemberWithUserSchema = z.object({
    team: TeamSchema,
    team_member: TeamMemberSchema,
    user: UserInfoSchema
})

export type TeamMemberWithUser = z.infer<typeof TeamMemberWithUserSchema>

// 团队列表响应类型
const TeamListResponseSchema = BaseResponseSchema.extend({
    data: z.array(TeamSchema)
})

export type TeamListResponse = z.infer<typeof TeamListResponseSchema>

// 团队成员列表响应类型
const TeamMemberListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        team: z.object({
            id: z.string(),
            name: z.string(),
            description: z.string().optional(),
            avatar: z.string().optional(),
            invited_perm_type: z.nativeEnum(TeamPermType),
            open_invite: z.boolean(),
            created_at: z.string(),
            updated_at: z.string()
        }),
        team_member: z.object({
            id: z.string(),
            team_id: z.string(),
            user_id: z.string(),
            perm_type: z.nativeEnum(TeamPermType),
            nickname: z.string(),
            created_at: z.string(),
            updated_at: z.string()
        }),
        user: z.object({
            id: z.string(),
            nickname: z.string(),
            avatar: z.string()
        })
    }))
})

export type TeamMemberListResponse = z.infer<typeof TeamMemberListResponseSchema>
export type TeamMemberListResponseData = z.infer<typeof TeamMemberListResponseSchema.shape.data>
// 团队信息响应类型
const TeamInfoResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        id: z.string(),
        name: z.string(),
        self_perm_type: z.nativeEnum(TeamPermType),
        invited_perm_type: z.nativeEnum(TeamPermType),
        invited_switch: z.boolean()
    })
})

export type TeamInfoResponse = z.infer<typeof TeamInfoResponseSchema>

// 项目申请列表响应类型
const ProjectApplyListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        project_id: z.string(),
        perm_type: z.number(),
        status: z.number(),
        first_displayed_at: z.string().nullable(),
        processed_at: z.string().nullable(),
        processed_by: z.string().nullable(),
        applicant_notes: z.string().nullable(),
        processor_notes: z.string().nullable()
    }))
})

export type ProjectApplyListResponse = z.infer<typeof ProjectApplyListResponseSchema>
export type ProjectApplyListItem = z.infer<typeof ProjectApplyListResponseSchema.shape.data.element>
// 团队申请列表响应类型
const TeamApplyListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        team_id: z.string(),
        perm_type: z.nativeEnum(TeamPermType),
        status: z.number(),
        first_displayed_at: z.string().nullable(),
        processed_at: z.string().nullable(),
        processed_by: z.string().nullable(),
        applicant_notes: z.string().nullable(),
        processor_notes: z.string().nullable()
    }))
})

export type TeamApplyListResponse = z.infer<typeof TeamApplyListResponseSchema>
export type TeamApplyListItem = z.infer<typeof TeamApplyListResponseSchema.shape.data.element>

// 项目列表响应类型
const TeamProjectListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        project: z.object({
            id: z.string(),
            team_id: z.string(),
            name: z.string(),
            description: z.string().optional(),
            is_public: z.boolean(),
            perm_type: z.nativeEnum(TeamPermType),
            open_invite: z.boolean(),
            need_approval: z.boolean(),
            created_at: z.string(),
            updated_at: z.string()
        }),
        creator: UserInfoSchema,
        creator_team_nickname: z.string().optional(),
        self_perm_type: z.nativeEnum(TeamPermType),
        is_in_team: z.boolean(),
        is_invited: z.boolean()
    }))
})

export type TeamProjectListResponse = z.infer<typeof TeamProjectListResponseSchema>
export type TeamProjectListItem = z.infer<typeof TeamProjectListResponseSchema.shape.data.element>
// 项目成员列表响应类型
const ProjectMemberListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        project: z.object({
            id: z.string(),
            name: z.string(),
            description: z.string().optional(),
            is_public: z.boolean(),
            perm_type: z.number(),
            open_invite: z.boolean(),
            need_approval: z.boolean(),
            created_at: z.string(),
            updated_at: z.string()
        }),
        project_member: z.object({
            id: z.string(),
            project_id: z.string(),
            user_id: z.string(),
            perm_type: z.number(),
            created_at: z.string(),
            updated_at: z.string()
        }),
        user: z.object({
            id: z.string(),
            nickname: z.string(),
            avatar: z.string()
        })
    }))
})

export type ProjectMemberListResponse = z.infer<typeof ProjectMemberListResponseSchema>
export type ProjectMemberListResponseData = z.infer<typeof ProjectMemberListResponseSchema.shape.data>

// 项目收藏列表响应类型
const ProjectFavoriteListResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        project_id: z.string(),
        created_at: z.string(),
        updated_at: z.string()
    }))
})

export type ProjectFavoriteListResponse = z.infer<typeof ProjectFavoriteListResponseSchema>

// 项目邀请信息响应类型
const ProjectInvitedInfoResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        id: z.string(),
        name: z.string(),
        self_perm_type: z.number(),
        invited_perm_type: z.number(),
        invited_switch: z.boolean()
    })
})

export type ProjectInvitedInfoResponse = z.infer<typeof ProjectInvitedInfoResponseSchema>

// 项目申请通知响应类型
const ProjectNoticeResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        project_id: z.string(),
        perm_type: z.number(),
        status: z.number(),
        first_displayed_at: z.string().nullable(),
        processed_at: z.string().nullable(),
        processed_by: z.string().nullable(),
        applicant_notes: z.string().nullable(),
        processor_notes: z.string().nullable()
    }))
})

export type ProjectNoticeResponse = z.infer<typeof ProjectNoticeResponseSchema>
export type ProjectNoticeListItem = z.infer<typeof ProjectNoticeResponseSchema.shape.data.element>

// 团队申请通知响应类型
const TeamNoticeResponseSchema = BaseResponseSchema.extend({
    data: z.array(z.object({
        id: z.string(),
        user_id: z.string(),
        team_id: z.string(),
        perm_type: z.nativeEnum(TeamPermType),
        status: z.number(),
        first_displayed_at: z.string().nullable(),
        processed_at: z.string().nullable(),
        processed_by: z.string().nullable(),
        applicant_notes: z.string().nullable(),
        processor_notes: z.string().nullable()
    }))
})

export type TeamNoticeResponse = z.infer<typeof TeamNoticeResponseSchema>
export type TeamNoticeListItem = z.infer<typeof TeamNoticeResponseSchema.shape.data.element>
export class TeamAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    // 创建团队
    async createTeam(params: {
        name: string;
        description?: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('创建团队响应数据校验失败:', error);
            throw error;
        }
    }

    // 创建项目
    async createProject(params: {
        team_id: string;
        name: string;
        description?: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('创建项目响应数据校验失败:', error);
            throw error;
        }
    }

    //获取项目申请列表
    async getTeamProjectApply(params: {
        team_id?: string;
        project_id?: string;
        status?: number;
        page?: number;
        page_size?: number;
        start_time?: number;
    }): Promise<ProjectApplyListResponse> {
        const result = await this.http.request({
            url: `/documents/team/project/apply`,
            method: 'get',
            params: params,
        });
        try {
            return ProjectApplyListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目申请列表数据校验失败:', error);
            throw error;
        }
    }

    //项目加入审核
    async postTeamProjectAudit(params: {
        apply_id: string;
        approval_code: number;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: `/documents/team/project/apply/audit`,
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('项目加入审核响应数据校验失败:', error);
            throw error;
        }
    }

    //获取团队申请列表
    async getTeamApply(params: {
        team_id?: string;
        status?: number;
        page?: number;
        page_size?: number;
        start_time? : number;
    }): Promise<TeamApplyListResponse> {
        const result = await this.http.request({
            url: `/documents/team/apply`,
            method: 'get',
            params: params,
        });
        try {
            return TeamApplyListResponseSchema.parse(result);
        } catch (error) {
            console.error('团队申请列表数据校验失败:', error);
            throw error;
        }
    }

    //团队加入审核
    async postTeamAudit(params: {
        apply_id: string;
        approval_code: number;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: `/documents/team/apply/audit`,
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('团队加入审核响应数据校验失败:', error);
            throw error;
        }
    }

    //获取项目收藏列表
    async getProjectFavoriteLists(params: {
        team_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<ProjectFavoriteListResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/favorite/list',
            method: 'get',
            params: params,
        });
        try {
            return ProjectFavoriteListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目收藏列表数据校验失败:', error);
            throw error;
        }
    }

    //是否收藏
    async setProjectIsFavorite(params: {
        project_id: string;
        is_favor: boolean;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/favorite',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置项目收藏状态响应数据校验失败:', error);
            throw error;
        }
    }

    //设置项目信息
    async setProjectInfo(params: {
        project_id: string;
        name?: string;
        description?: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/info',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置项目信息响应数据校验失败:', error);
            throw error;
        }
    }

    //设置项目邀请信息
    async setProjectInvitedInfo(params: {
        project_id: string;
        invited_perm_type?: number;
        invited_switch?: boolean;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/invited',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置项目邀请信息响应数据校验失败:', error);
            throw error;
        }
    }

    // 获取项目邀请信息
    async getProjectInvitedInfo(params: {
        project_id: string;
    }): Promise<ProjectInvitedInfoResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/info/invited',
            method: 'get',
            params: params,
        });
        try {
            return ProjectInvitedInfoResponseSchema.parse(result);
        } catch (error) {
            console.error('获取项目邀请信息数据校验失败:', error);
            throw error;
        }
    }

    //申请加入项目
    async applyJoinProject(params: {
        project_id: string;
        applicant_notes?: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/apply',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('申请加入项目响应数据校验失败:', error);
            throw error;
        }
    }

    //获取项目列表
    async getProjectLists(params: {
        team_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<TeamProjectListResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/list',
            method: 'get',
            params: params,
        });
        try {
            return TeamProjectListResponseSchema.parse(result);
        } catch (error) {
            console.error('团队项目列表数据校验失败:', error);
            throw error;
        }
    }

    //获取项目成员列表
    async getProjectMemberList(params: {
        project_id: string;
        page?: number;
        page_size?: number;
    }): Promise<ProjectMemberListResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/member/list',
            method: 'get',
            params: params,
        });
        try {
            return ProjectMemberListResponseSchema.parse(result);
        } catch (error) {
            console.error('项目成员列表数据校验失败:', error);
            throw error;
        }
    }

    //退出项目
    async exitProject(params: {
        project_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/exit',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('退出项目响应数据校验失败:', error);
            throw error;
        }
    }

    //设置项目成员权限
    async setProjectMemberPerm(params: {
        project_id: string;
        user_id: string;
        perm_type: number;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/member/perm',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置项目成员权限响应数据校验失败:', error);
            throw error;
        }
    }

    //将成员移出项目组
    async delProjectMember(params: {
        project_id: string;
        user_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/member',
            method: 'delete',
            params: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('将成员移出项目组响应数据校验失败:', error);
            throw error;
        }
    }

    //转让项目创建者
    async transferProjectCreator(params: {
        project_id: string;
        user_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/creator',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('转让项目创建者响应数据校验失败:', error);
            throw error;
        }
    }

    //删除项目
    async delProject(params: {
        project_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project',
            method: 'delete',
            params: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('删除项目响应数据校验失败:', error);
            throw error;
        }
    }

    //移动文档
    async moveDocument(params: {
        document_id: string;
        target_team_id?: string;
        target_project_id?: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/document/move',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('移动文档响应数据校验失败:', error);
            throw error;
        }
    }

    // 获取项目的申请通知信息
    async getProjectNotice(params: {
        team_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<ProjectNoticeResponse> {
        const result = await this.http.request({
            url: '/documents/team/project/self_apply',
            method: 'get',
            params: params,
        });
        try {
            return ProjectNoticeResponseSchema.parse(result);
        } catch (error) {
            console.error('获取项目申请通知信息数据校验失败:', error);
            throw error;
        }
    }

    //获取团队的申请通知信息
    async getTeamNotice(params: {
        team_id?: string;
        page?: number;
        page_size?: number;
    }): Promise<TeamNoticeResponse> {
        const result = await this.http.request({
            url: '/documents/team/self_apply',
            method: 'get',
            params: params,
        });
        try {
            return TeamNoticeResponseSchema.parse(result);
        } catch (error) {
            console.error('获取团队申请通知信息数据校验失败:', error);
            throw error;
        }
    }

    //设置团队成员昵称
    async setTeamMemberNickname(params: {
        team_id: string;
        user_id: string;
        nickname: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/member/nickname',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置团队成员昵称响应数据校验失败:', error);
            throw error;
        }
    }

    // 获取团队列表
    async getTeamList(params: {
        page?: number;
        page_size?: number;
    }): Promise<TeamListResponse> {
        const result = await this.http.request({
            url: '/documents/team/list',
            method: 'get',
            params,
        });
        try {
            return TeamListResponseSchema.parse(result);
        } catch (error) {
            console.error('团队列表数据校验失败:', error);
            throw error;
        }
    }

    // 获取团队成员列表
    async getTeamMemberList(params: {
        team_id: string;
        page?: number;
        page_size?: number;
    }): Promise<TeamMemberListResponse> {
        const result = await this.http.request({
            url: '/documents/team/member/list',
            method: 'get',
            params,
        });
        try {
            return TeamMemberListResponseSchema.parse(result);
        } catch (error) {
            console.error('团队成员列表数据校验失败:', error);
            throw error;
        }
    }

    // 设置团队信息
    async setTeamInfo(params: {
        team_id: string;
        name?: string;
        description?: string;
        avatar?: File
    }): Promise<BaseResponse> {
        const formData = new FormData();
        formData.append('team_id', params.team_id);
        if (params.name) {
            formData.append('name', params.name);
        }
        if (params.description) {
            formData.append('description', params.description);
        }
        if (params.avatar) {
            formData.append('avatar', params.avatar);
        }
        const result = await this.http.request({
            url: '/documents/team/info',
            method: 'put',
            data: formData,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置团队信息响应数据校验失败:', error);
            throw error;
        }
    }

    //设置团队邀请选项
    async setTeamInviteInfo(params: {
        team_id: string;
        open_invite?: boolean;
        invited_perm_type?: number;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/invited',
            method: 'put',
            data: params,
        })
    }

    //转移团队创建者
    async setTeamCreator(params: {
        team_id: string;
        user_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/creator',
            method: 'put',
            data: params,
        })
    }

    //获取团队信息
    async getTeamInfo(params: {
        team_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/info/invited',
            method: 'get',
            params: params,
        })
    }

    //申请加入团队
    async joinTeam(params: {
        team_id: string;
        applicant_notes?: string
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/apply',
            method: 'post',
            data: params,
        })
    }

    //获取申请列表
    async getJoinList(params: {
        team_id: string;
        page?: number;
        page_size?: number;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/join/list',
            method: 'get',
            params: params,
        })
    }

    //删除团队成员
    async deletTeamMember(params: {
        team_id: string;
        user_id: string;
    }): Promise<BaseResponse> {
        return this.http.request({
            url: '/documents/team/member',
            method: 'delete',
            params: params,
        })
    }

    // 设置团队成员权限
    async setTeamMemberPermission(params: {
        team_id: string;
        user_id: string;
        perm_type: TeamPermType;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/member/perm',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('设置团队成员权限响应数据校验失败:', error);
            throw error;
        }
    }

    // 退出团队
    async exitTeam(params: {
        team_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/exit',
            method: 'post',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('退出团队响应数据校验失败:', error);
            throw error;
        }
    }

    // 删除团队
    async deleteTeam(params: {
        team_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team',
            method: 'delete',
            params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('删除团队响应数据校验失败:', error);
            throw error;
        }
    }

    // 转移团队创建者
    async changeTeamCreator(params: {
        team_id: string;
        user_id: string;
    }): Promise<BaseResponse> {
        const result = await this.http.request({
            url: '/documents/team/creator',
            method: 'put',
            data: params,
        });
        try {
            return BaseResponseSchema.parse(result);
        } catch (error) {
            console.error('转移团队创建者响应数据校验失败:', error);
            throw error;
        }
    }
}