import * as timing_util from "./timing_util"
import { convert } from "./cmd_convert";
import { Connect, ConnectClient } from "./connect";
import { Cmd, DataTypes } from "./types";

export interface CoopNetError {
    type: "duplicate";
    duplicateCmd: Cmd;
}

export class CoopNet extends ConnectClient {

    constructor(connect: Connect) {
        super(connect, DataTypes.Op)
    }

    private watcherList: ((cmds: Cmd[]) => void)[] = []
    private errorWatcherList: ((errorInfo: CoopNetError) => void)[] = []

    async _pullCmds(from?: number, to?: number): Promise<Cmd[]> {
        const ready = await this.waitReady()
        if (!ready) return [];
        const result = await this.send({
            type: "pullCmds",
            from: from,
            to: to,
        }, 0, 0)

        const cmdsData = JSON.parse(result.data.cmds_data ?? '""') as any[]
        let cmds: Cmd[] | undefined
        if (Array.isArray(cmdsData)) {
            cmds = cmdsData;
        }
        return cmds || []
    }

    pullCmds = timing_util.throttle(this._pullCmds.bind(this), 1000)

    async _postCmds(cmds: Cmd[], serial:(cmds: Cmd[])=> string): Promise<boolean> {
        const ready = await this.waitReady()
        if (!ready) return false;
        this.send({
            type: "commit",
            cmds: serial(cmds),
        }, 0, 0)
        return true;
    }

    postCmds = timing_util.throttle(this._postCmds.bind(this), 1000)

    watchCmds(watcher: (cmds: Cmd[]) => void) {
        this.watcherList.push(watcher)
        return () => {
            const watcherIndex = this.watcherList.findIndex(w => w === watcher);
            if (watcherIndex > -1) this.watcherList.splice(watcherIndex, 1);
        }
    }

    watchError(watcher: (errorInfo: CoopNetError) => void): void {
        this.errorWatcherList.push(watcher)
    }

    onMessage(data: any): void {
        const cmdsData = JSON.parse(data.cmds_data ?? '""') as any[]
        let cmds: Cmd[] | undefined
        if (Array.isArray(cmdsData)) {
            cmds = cmdsData;
        }
        if (data.type === "update") {
            if (!Array.isArray(cmds)) {
                console.log("返回数据格式错误")
                return
            }
            console.log("update", cmds)
            this.watcherList.slice(0).forEach(watcher => watcher(convert(cmds)));
        } else if (data.type === "errorInvalidParams") {
            console.log("参数错误")
        } else if (data.type === "errorNoPermission") {
            console.log("无权限")
        } else if (data.type === "errorInsertFailed") {
            console.log("数据插入失败", data.cmd_id_list)
            if (!Array.isArray(data.cmd_id_list)) {
                console.log("返回数据格式错误")
                return
            }
            if (data.data?.type === "duplicate") {
                const duplicateCmd = data.data?.duplicateCmd
                console.log("数据重复", duplicateCmd)
                if (!duplicateCmd) {
                    console.log("返回数据格式错误")
                    return
                }
                duplicateCmd.cmd.id = duplicateCmd.cmd_id
                if (!duplicateCmd.cmd.ops) duplicateCmd.cmd.ops = [];
                const duplicateCmd1 = duplicateCmd.cmd
                for (const watcher of this.errorWatcherList) {
                    watcher({
                        type: "duplicate",
                        duplicateCmd: duplicateCmd1,
                    })
                }
            }
        } else {
            console.log("未知的数据类型", data.type)
        }
    }
}
