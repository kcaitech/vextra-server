import { CommentAPI } from "./comment";
import { HttpMgr } from "./http";
import { ShareAPI } from "./share";
import { TeamAPI } from "./team";
import { UsersAPI } from "./users";
import { DocumentAPI } from "./document";
export class Request {
    user_api: UsersAPI
    team_api: TeamAPI
    comment_api: CommentAPI
    share_api: ShareAPI
    document_api: DocumentAPI
    constructor(apiUrl: string, onUnauthorized: () => void, token?: string) {
        const httpmgr = new HttpMgr(apiUrl, onUnauthorized, token);
        this.user_api = new UsersAPI(httpmgr);
        this.team_api = new TeamAPI(httpmgr);
        this.comment_api = new CommentAPI(httpmgr);
        this.share_api = new ShareAPI(httpmgr);
        this.document_api = new DocumentAPI(httpmgr);
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

