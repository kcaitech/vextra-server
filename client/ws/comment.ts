import { Connect, ConnectClient } from "./connect";
import { DataTypes, DocCommentOpData } from "./types";

export class Comment extends ConnectClient {
    private updateHandlerSet = new Set<(data: DocCommentOpData) => void>()
    
    constructor(connect: Connect) {
        super(connect, DataTypes.Comment)
    }

    onMessage(data: any): void {
        // throw new Error("Method not implemented.");
        this._onUpdated(data as DocCommentOpData)
    }
    
    private _onUpdated(docCommentOpData: DocCommentOpData) {
        for (const handler of this.updateHandlerSet) handler(docCommentOpData);
    }

    public addUpdatedHandler(onUpdated: (docCommentOpData: DocCommentOpData) => void) {
        this.updateHandlerSet.add(onUpdated)
    }

    public removeUpdatedHandler(onUpdated: (docCommentOpData: DocCommentOpData) => void) {
        this.updateHandlerSet.delete(onUpdated)
    }
}