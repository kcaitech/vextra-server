// import { exportExForm, Document } from "@kcdesign/data";
import { v4 } from "uuid";
import { Connect, ConnectClient } from "./connect";
import { DataTypes } from "./types";

export type ExportFunc = () => Promise<{
    media_names: string[]
    // todo
}>

export interface MediasMgr {
    get(name: string): Promise<{
        buff: Uint8Array
    }>
}

export class DocUpload extends ConnectClient {
    constructor(connect: Connect) {
        super(connect, DataTypes.DocUpload)
    }

    onMessage(data: any): void {
        // 
    }

    public async upload(exportExForm: ExportFunc, mediasMgr: MediasMgr, project_id?: string): Promise<undefined | { document_id: string, version_id: string }> {

        const netReady = await this.waitReady();
        if (!netReady) return;
        const document_id = v4();
        let data
        let ret = false;
        try {
            data = await exportExForm()
            ret = !!await this.send({ document_id, project_id, export: data }, 60000) // 一分钟超时
        } catch (e) {
            console.log(e)
            return
        }
        if (!ret) {
            console.log("upload-send fail")
            return;
        }
        for (let i = 0, len = data.media_names.length; i < len; i++) {
            const name = data.media_names[i];
            const buffer = await mediasMgr.get(name)
            if (buffer !== undefined) {
                ret = !!await this.sendBinary({ document_id, media: name }, buffer.buff.buffer, 60000); // 一分钟超时
                if (!ret) {
                    console.log("upload-sendBinary fail")
                    return;
                }
            }
        }
        const confirm = await this.send({ document_id, commit: true }, 60000) // 一分钟超时
        if (confirm.err !== undefined) {
            console.log("upload confirm fail", confirm.err)
            return;
        }
        return (confirm.data as { document_id: string, version_id: string });
    }
}