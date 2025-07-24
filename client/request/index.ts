import { CommentAPI } from "./comment";
import { HttpMgr } from "./http";
import { ShareAPI } from "./share";
import { TeamAPI } from "./team";
import { UsersAPI } from "./users";
import { DocumentAPI } from "./document";
import { AccessAPI } from "./access";

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
    access_api: AccessAPI
    private constructor(apiUrl: string, onUnauthorized: () => void, token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    } = defaultToken, timeout?: number) {
        const httpmgr = new HttpMgr(apiUrl, onUnauthorized, token, timeout);
        this.user_api = new UsersAPI(httpmgr);
        this.team_api = new TeamAPI(httpmgr);
        this.comment_api = new CommentAPI(httpmgr);
        this.share_api = new ShareAPI(httpmgr);
        this.document_api = new DocumentAPI(httpmgr);
        this.access_api = new AccessAPI(httpmgr);
    }

    // 单例模式
    static instance: Request | null = null;
    static getInstance(apiUrl: string, onUnauthorized: () => void, token: {
        getToken: () => string | undefined,
        setToken: (token: string | undefined) => void
    } = defaultToken, timeout?: number) {
        if (!Request.instance) {
            Request.instance = new Request(apiUrl, onUnauthorized, token, timeout);
        }
        return Request.instance;
    }
}



export { HttpCode } from "./httpcode"
export * from "./users"
export * from "./team"
export * from "./share"
export * from "./comment"
export * from "./document"
export * from "./access"

export type { ThumbnailResponseData } from "./types"