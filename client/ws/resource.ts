import { Connect, ConnectClient } from "./connect";
import { DataTypes, ResourceHeader } from "./types";

export class Resource extends ConnectClient {
    constructor(connect: Connect) {
        super(connect, DataTypes.Resource)
    }

    onMessage(data: any): void {
    }

    public async upload(name: string, data: ArrayBuffer): Promise<boolean> {

        // let count = 0
        // while (count++ < 3 && !await this.docResourceUpload.uploadResource(name, data)) {
        //     await new Promise(resolve => setTimeout(resolve, 1000))
        // }
        // return count < 3;
        await this.waitReady()
        return !!this.sendBinary({
            name: name,
        } as ResourceHeader, data, 1000, 3);
    }
}