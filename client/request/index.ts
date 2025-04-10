import { CommentAPI } from "./comment";
import { HttpMgr } from "./http";
import { ShareAPI } from "./share";
import { TeamAPI } from "./team";
import { UsersAPI } from "./users";

export class Request {
    user_api: UsersAPI
    team_api: TeamAPI
    comment_api: CommentAPI
    share_api: ShareAPI

    constructor(apiUrl: string, onUnauthorized: () => void, token?: string) {
        const httpmgr = new HttpMgr(apiUrl, onUnauthorized, token);
        this.user_api = new UsersAPI(httpmgr);
        this.team_api = new TeamAPI(httpmgr);
        this.comment_api = new CommentAPI(httpmgr);
        this.share_api = new ShareAPI(httpmgr);
    }
}

export { HttpCode } from "./httpcode"
export * from "./users"
export * from "./team"
export * from "./share"
export * from "./comment"