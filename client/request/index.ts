import { CommentAPI } from "./comment";
import { HttpMgr } from "./http";
import { ShareAPI } from "./share";
import { TeamAPI } from "./team";
import { UsersAPI } from "./users";
import { DocumentAPI } from "./document";

const defaultToken = {
    getToken: () => {
        return localStorage.getItem('token') ?? undefined
    },
    setToken: (token: string | undefined) => {
        if (token) {
            localStorage.setItem('token', token)
        } else {
            localStorage.removeItem('token')
        }
    }
}

export class Request {
    user_api: UsersAPI
    team_api: TeamAPI
    comment_api: CommentAPI
    share_api: ShareAPI
    document_api: DocumentAPI
    private constructor(apiUrl: string, onUnauthorized: () => void, token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    } = defaultToken) {
        const httpmgr = new HttpMgr(apiUrl, onUnauthorized, token);
        this.user_api = new UsersAPI(httpmgr);
        this.team_api = new TeamAPI(httpmgr);
        this.comment_api = new CommentAPI(httpmgr);
        this.share_api = new ShareAPI(httpmgr);
        this.document_api = new DocumentAPI(httpmgr);
    }

    // todo
    watch(key: string, callback: (value: any) => void) {
    }

    // 需要单例模式。不然如有多个实例，可能同时刷新token，导致先刷新成功token的请求带上的token失效
    static instance: Request | null = null;
    static getInstance(apiUrl: string, onUnauthorized: () => void, token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    } = defaultToken) {
        if (!Request.instance) {
            Request.instance = new Request(apiUrl, onUnauthorized, token);
        }
        return Request.instance;
    }
}

export type { UserInfo, TeamInfo, ProjectInfo, LockedInfo } from "./types"

export { TeamPermType, LocketType } from "./types"

export { HttpCode } from "./httpcode"
export * from "./users"
export * from "./team"
export * from "./share"
export * from "./comment"
export * from "./document"

