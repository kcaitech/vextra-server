/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import { Connect } from "./connect";
import { CoopNet } from "./op";
import { Resource } from "./resource";
import { Selection } from "./selection"
import { Comment } from "./comment"
import { DocUpload, ExportFunc, MediasMgr } from "./upload";
import { BindResponse, BindResponseSchema, Cmd, DataTypes, IClientContext, NetworkStatusType } from "./types";
import { Version } from "./version";
import { Thumbnail } from "./thumbnail";
export { Connect } from "./connect"
export { Selection as SelectionSync } from "./selection"
export { DocUpload } from "./upload"
export * from "./types"

export class WSClient {
    connect: Connect

    private op: CoopNet
    private resource: Resource
    private selection?: Selection;
    private comment: Comment;

    private bind_document_id?: string;
    private _started?: IClientContext;
    private version: Version;
    private thumbnail: Thumbnail;

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

    // from: server 表示是服务端发起的连接，client 表示是客户端发起的连接. 默认是client
    constructor(wsUrl: string, token?: string, from?: 'server' | 'client') {
        this.connect = new Connect(wsUrl, token, from);
        this.connect.addOnChange(this.onNetChange.bind(this))
        this.op = new CoopNet(this.connect)
        this.resource = new Resource(this.connect)
        this.comment = new Comment(this.connect)
        this.version = new Version(this.connect)
        this.thumbnail = new Thumbnail(this.connect)
    }

    // 先bind，再start
    async bind(document_id: string): Promise<BindResponse> {
        await this.connect.waitReady()
        const ret = await this.connect.send(DataTypes.Bind, {
            document_id
        })
        // data {
        //     "doc_info":   docInfo,
        //     "access_key": accessKey,
        // }
        this.bind_document_id = document_id
        return BindResponseSchema.parse(ret);
    }

    async start(context: IClientContext) {
        if (!this.bind_document_id) throw new Error("need bind first");
        await this.connect.waitReady()
        const ret = await this.connect.send(DataTypes.Start, { last_cmd_version: context.lastRemoteCmdVersion() })
        this._started = context;
        return ret;
    }

    upload(name: string, data: ArrayBuffer): Promise<boolean> {
        return this.resource.upload(name, data)
    }
    genThumbnail(name: string, cotentType: string, data: ArrayBuffer): Promise<boolean> {
        return this.thumbnail.upload(name, cotentType, data)
    }
    hasConnected(): boolean {
        return this.connect.isReady;
    }
    pullCmds(from: number, to?: number): Promise<Cmd[]> {
        return this.op.pullCmds(from, to);
    }
    postCmds(cmds: Cmd[], serial:(cmds: Cmd[])=> string): Promise<boolean> {
        return this.op.postCmds(cmds, serial);
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

    getSelectionSync() {
        if (this.selection) return this.selection;
        if (!this._started) throw new Error("need start first");
        this.selection = new Selection(this.connect, this._started)
        return this.selection;
    }

    getCommentSync() {
        return this.comment
    }

    getVersionSync() {
        return this.version
    }

    uploadDocument(exportExForm: ExportFunc, mediasMgr: MediasMgr, project_id?: string) {
        return new DocUpload(this.connect).upload(exportExForm, mediasMgr, project_id)
    }
}