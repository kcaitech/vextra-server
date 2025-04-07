import { Cmd, ICoopNet, serialCmds, parseCmds, cloneCmds, RadixConvert } from "@kcdesign/data"
import * as timing_util from "./timing_util"
import { convert } from "./cmd_convert";
import { Connect, ConnectClient } from "./connect";
import { DataTypes } from "./types";

export class CoopNet extends ConnectClient implements ICoopNet {

    constructor(connect: Connect) {
        super(connect, DataTypes.Op)
    }

    // private send?: (data: any, isListened?: boolean, timeout?: number) => Promise<boolean>
    private watcherList: ((cmds: Cmd[]) => void)[] = []
    private errorWatcherList: ((errorInfo: {
        type: "duplicate",
        duplicateCmd: Cmd,
    }) => void)[] = []
    // private onClose?: () => void
    // private isConnected = false
    private pullCmdsPromiseList: Record<string, {
        resolve: (value: Cmd[]) => void,
        reject: (reason: any) => void
    }[]> = {}
    private radixRevert: RadixConvert = new RadixConvert(62)


    async _pullCmds(from?: string, to?: string): Promise<Cmd[]> {
        const ready = await this.waitReady()
        if (!ready) return [];
        from = from ? this.radixRevert.to(from).toString(10) : ""
        to = to ? this.radixRevert.to(to).toString(10) : ""
        console.log("pullCmds", from, to)
        this.send({
            type: "pullCmds",
            from: from,
            to: to,
        }, 0, 0)
        return new Promise<Cmd[]>((resolve, reject) => {
            const key = `${from}-${to}`
            let promiseList = this.pullCmdsPromiseList[key]
            if (!promiseList) promiseList = this.pullCmdsPromiseList[key] = [];
            promiseList.push({ resolve: resolve, reject: reject })
        })
    }

    pullCmds = timing_util.throttle(this._pullCmds.bind(this), 1000)

    async _postCmds(cmds: Cmd[]): Promise<boolean> {
        const ready = await this.waitReady()
        if (!ready) return false;
        console.log("postCmds", cloneCmds(cmds))
        this.send({
            type: "commit",
            cmds: serialCmds(cmds),
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

    watchError(watcher: (errorInfo: {
        type: "duplicate",
        duplicateCmd: Cmd,
    }) => void): void {
        this.errorWatcherList.push(watcher)
    }

    onMessage(data: any): void {
        const cmdsData = JSON.parse(data.cmds_data ?? '""') as any[]
        let cmds: Cmd[] | undefined
        let cmds1: Cmd[] | undefined
        if (Array.isArray(cmdsData)) {
            const data = cmdsData.map(item => {
                item.cmd.id = item.cmd_id
                item.cmd.version = this.radixRevert.from(item.id)
                item.cmd.previousVersion = this.radixRevert.from(item.previous_id)
                return item.cmd
            })
            cmds = parseCmds(data)
            cmds1 = parseCmds(data)
        }
        // pullCmdsResult update errorInvalidParams errorNoPermission errorInsertFailed errorPullCmdsFailed
        if (data.type === "pullCmdsResult" || data.type === "errorPullCmdsFailed") {
            if (data.type === "errorPullCmdsFailed") console.log("拉取数据失败");

            const from = typeof data.from === "string" ? data.from : ""
            const to = typeof data.to === "string" ? data.to : ""

            const key = `${from}-${to}`
            if (!this.pullCmdsPromiseList[key]) return;

            if (data.type === "pullCmdsResult") {
                if (!Array.isArray(cmds)) {
                    console.log("返回数据格式错误")
                    for (const item of this.pullCmdsPromiseList[key]) item.reject(new Error("返回数据格式错误"));
                } else {
                    console.log("pullCmdsResult", cmds1)
                    for (const item of this.pullCmdsPromiseList[key]) item.resolve(cmds);
                }
                // 有什么用？
                // if (typeof data.previous_id !== "string") {
                //     console.log("返回数据格式错误，缺少previous_id")
                // }
            } else {
                for (const item of this.pullCmdsPromiseList[key]) item.reject(new Error("拉取数据失败"));
            }

            delete this.pullCmdsPromiseList[key]
        } else if (data.type === "update") {
            if (!Array.isArray(cmds)) {
                console.log("返回数据格式错误")
                return
            }
            console.log("update", cmds1)
            // for (const watcher of this.watcherList) watcher(convert(cmds));
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
                duplicateCmd.cmd.version = this.radixRevert.from(duplicateCmd.id)
                duplicateCmd.cmd.previousVersion = this.radixRevert.from(duplicateCmd.previous_id)
                if (!duplicateCmd.cmd.ops) duplicateCmd.cmd.ops = [];
                const duplicateCmd1 = parseCmds([duplicateCmd.cmd])[0]
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
