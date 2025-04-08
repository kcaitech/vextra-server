// import { Cmd, Document, WatchableObject } from "@kcdesign/data";
// import { IContext, INet } from "@kcdesign/editor";
import { Connect } from "./connect";
import { CoopNet } from "./op";
import { Resource } from "./resource";
import { Selection } from "./selection"
import { Comment } from "./comment"
import { DocUpload, ExportFunc, MediasMgr } from "./upload";
import { Cmd, DataTypes, IContext, NetworkStatusType } from "./types";
export { Connect } from "./connect"
export { Selection as SelectionSync } from "./selection"
export { DocUpload } from "./upload"

export class WSClient {
    // private context: IContext
    connect: Connect

    private op: CoopNet
    private resource: Resource
    private selection?: Selection;
    private comment: Comment

    private bind_document_id?: string;
    private _started?: IContext;

    private async onNetChange(networkStatus: NetworkStatusType) {
        // 如果先走了ready,也没事
        if (networkStatus !== NetworkStatusType.Online) return;
        if (this.bind_document_id) {
            this.bind(this.bind_document_id) // todo 有可能这时没权限了,正常没权限了要刷新页面
        }
        if (this._started) {
            this.start(this._started);
        }
    }

    constructor() {
        // super();
        // this.context = context;
        this.connect = new Connect();
        this.connect.addOnChange(this.onNetChange.bind(this))
        this.op = new CoopNet(this.connect)
        this.resource = new Resource(this.connect)
        this.comment = new Comment(this.connect)
    }

    // 先bind，再start
    async bind(document_id: string) {
        await this.connect.waitReady()
        const ret = await this.connect.send(DataTypes.Bind, {
            document_id
        })
        this.bind_document_id = document_id
        return ret;
    }

    async start(context: IContext) {
        await this.connect.waitReady()
        const ret = await this.connect.send(DataTypes.Start, { last_cmd_version: context.lastRemoteCmdVersion() })
        this._started = context;
        return ret;
    }

    upload(name: string, data: ArrayBuffer): Promise<boolean> {
        return this.resource.upload(name, data)
    }
    hasConnected(): boolean {
        return this.connect.isReady;
    }
    pullCmds(from: string, to?: string): Promise<Cmd[]> {
        return this.op.pullCmds(from, to);
    }
    postCmds(cmds: Cmd[]): Promise<boolean> {
        return this.op.postCmds(cmds);
    }

    watchCmds(watcher: (cmds: Cmd[]) => void) {
        return this.op.watchCmds(watcher);
    }
    watchError(watcher: (errorInfo: any
    ) => void): void {
        this.op.watchError(watcher);
    }

    close() {
        this.connect.close();
    }

    getSelectionSync(context: IContext) {
        if (this.selection) return this.selection;
        this.selection = new Selection(this.connect, context)
        return this.selection;
    }

    getCommentSync() {
        return this.comment
    }

    uploadDocument(exportExForm: ExportFunc, mediasMgr: MediasMgr, project_id?: string) {
        return new DocUpload(this.connect).upload(exportExForm, mediasMgr, project_id)
    }
}