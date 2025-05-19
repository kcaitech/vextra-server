import { Connect, ConnectClient } from "./connect";
import { DataTypes, VersionData } from "./types";

export class Version extends ConnectClient {
    private updateHandlerSet = new Set<(data: VersionData) => void>()
    
    constructor(connect: Connect) {
        super(connect, DataTypes.GenerateVersion)
    }

    onMessage(data: any): void {
        this._onUpdated(data as VersionData)
    }
    
    private _onUpdated(versionData: VersionData) {
        // @ts-ignore
        for (const handler of this.updateHandlerSet) handler(versionData);
    }

    public addUpdatedHandler(onUpdated: (versionData: VersionData) => void) {
        this.updateHandlerSet.add(onUpdated)
    }

    public removeUpdatedHandler(onUpdated: (versionData: VersionData) => void) {
        this.updateHandlerSet.delete(onUpdated)
    }
}